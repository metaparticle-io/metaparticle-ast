package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/glog"
	"github.com/metaparticle-io/metaparticle-ast/client"
	"github.com/metaparticle-io/metaparticle-ast/client/services"
	"github.com/metaparticle-io/metaparticle-ast/models"
	flag "github.com/spf13/pflag"
	apps_v1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	port   = flag.Int("port", 8080, "The port to connect to.")
	host   = flag.String("host", "", "The host to connect to")
	file   = flag.StringP("file", "f", "", "The config file to load")
	name   = flag.StringP("name", "n", "", "The name of the service to compile")
	dryrun = flag.Bool("dry-run", false, "If true, only output the execution plan, don't actually enact it.")
)

func output(obj interface{}) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	os.Stdout.Write(data)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// TODO: fix this for glog, see: https://github.com/kubernetes/kubernetes/pull/3342/files
	flag.Parse()

	var c *client.AnApplicationForEasierDistributedApplicationGeneration
	if len(*host) > 0 {
		addr := fmt.Sprintf("%s:%d", *host, *port)
		tc := client.DefaultTransportConfig().WithHost(addr)
		c = client.NewHTTPClientWithConfig(nil, tc)
	}

	if len(*file) == 0 && len(*name) == 0 {
		log.Fatalf("--file/-f or --name/-n is required.")
	}
	obj := &models.Service{}
	if len(*file) > 0 {
		bytes, err := ioutil.ReadFile(*file)
		if err != nil {
			glog.Fatalf("Couldn't read file: %v", err)
		}
		if err := obj.UnmarshalBinary(bytes); err != nil {
			glog.Fatalf("Couldn't parse file: %v", err)
		}
		if c != nil {
			updateParams := services.NewCreateOrUpdateServiceParamsWithTimeout(5 * time.Second)
			updateParams.Body = obj
			updateParams.Name = *obj.Name
			_, err = c.Services.CreateOrUpdateService(updateParams)
			if err != nil {
				glog.Fatalf("Failed to update: %s", err.Error())
			}
		}
	}
	if len(*name) > 0 {
		params := &services.GetServiceParams{Name: *name}
		params = params.WithTimeout(5 * time.Second)
		resp, err := c.Services.GetService(params)
		if err != nil {
			glog.Fatalf("Failed to get service: %s", err.Error())
		}
		obj = resp.Payload
	}

	compile(obj, *kubeconfig)
}

func envvars(container *models.Container) []v1.EnvVar {
	envvars := []v1.EnvVar{}
	for _, env := range container.Env {
		envvars = append(envvars, v1.EnvVar{
			Name:  *env.Name,
			Value: *env.Value,
		})
	}
	return envvars
}

func containers(service *models.ServiceSpecification) []v1.Container {
	containers := []v1.Container{}
	for ix, c := range service.Containers {
		containers = append(containers, v1.Container{
			Name:  fmt.Sprintf("%s-%d", *service.Name, ix),
			Image: *c.Image,
			Env:   envvars(c),
		})
	}
	return containers
}

func deploy(service *models.ServiceSpecification, client *kubernetes.Clientset) {
	name := *service.Name

	deployment := &v1beta1.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &service.Replicas,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: containers(service),
				},
			},
		},
	}

	if *dryrun {
		output(deployment)
		return
	}

	if _, err := client.ExtensionsV1beta1().Deployments("default").Create(deployment); err != nil {
		log.Fatalf(err.Error())
	}
}

func deployStateful(service *models.ServiceSpecification, client *kubernetes.Clientset) {
	name := *service.Name

	deployment := &apps_v1beta1.StatefulSet{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: apps_v1beta1.StatefulSetSpec{
			Replicas: &service.Replicas,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: containers(service),
				},
			},
		},
	}

	if *dryrun {
		output(deployment)
	} else {
		if _, err := client.AppsV1beta1().StatefulSets("default").Create(deployment); err != nil {
			log.Fatalf(err.Error())
		}
	}

	name = name + "-sharder"
	shardDeployment := &v1beta1.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &service.Replicas,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						v1.Container{
							Name:  "sharder",
							Image: "brendanburns/sharder",
						},
					},
				},
			},
		},
	}

	if *dryrun {
		output(shardDeployment)
		return
	}

	if _, err := client.ExtensionsV1beta1().Deployments("default").Create(shardDeployment); err != nil {
		log.Fatalf(err.Error())
	}
}

func getPorts(service *models.ServiceSpecification) []v1.ServicePort {
	ports := []v1.ServicePort{}
	for px := range service.Ports {
		port := service.Ports[px]
		ports = append(ports, v1.ServicePort{
			Port:     *port.Number,
			Protocol: "TCP",
		})
	}
	return ports
}

func createLoadBalancedService(service *models.ServiceSpecification, public bool, client *kubernetes.Clientset) {
	name := *service.Name

	svc := &v1.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: getPorts(service),
		},
	}

	if public {
		svc.Spec.Type = "LoadBalancer"
	}

	if *dryrun {
		output(svc)
		return
	}

	if _, err := client.CoreV1().Services("default").Create(svc); err != nil {
		log.Fatalf(err.Error())
	}
}

func createStatefulService(service *models.ServiceSpecification, client *kubernetes.Clientset) {
	name := *service.Name

	svc := &v1.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports:     getPorts(service),
			ClusterIP: "None",
		},
	}

	if *dryrun {
		output(svc)
		return
	}

	if _, err := client.CoreV1().Services("default").Create(svc); err != nil {
		log.Fatalf(err.Error())
	}
}

func compile(service *models.Service, kubeconfig string) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	for ix := range service.Services {
		if service.Services[ix].Replicas > 0 && service.Services[ix].ShardSpec != nil {
			log.Fatalf("%v: Replicas and shards are mutually exclusive.", service.Services[ix].Name)
		}
		if service.Services[ix].Replicas > 0 {
			deploy(service.Services[ix], clientset)
			public := *service.Serve.Name == *service.Services[ix].Name && service.Serve.Public
			createLoadBalancedService(service.Services[ix], public, clientset)
		}
		if service.Services[ix].ShardSpec != nil && service.Services[ix].ShardSpec.Shards > 0 {
			deployStateful(service.Services[ix], clientset)
		}
	}
}
