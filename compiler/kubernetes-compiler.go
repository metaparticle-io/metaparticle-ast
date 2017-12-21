package compiler

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/metaparticle-io/metaparticle-ast/ktail"
	"github.com/metaparticle-io/metaparticle-ast/models"
	apps_v1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	foreground    = meta.DeletePropagationForeground
	deleteOptions = &meta.DeleteOptions{
		PropagationPolicy: &foreground,
	}
)

type kubernetesCompiler struct {
	opts      *CompilerOptions
	clientset *kubernetes.Clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// NewKubernetesCompiler creates an Kubernetes Compiler instance
func NewKubernetesCompiler() (Compiler, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kubernetesCompiler{
		clientset: clientset,
	}, nil
}

func makeSharderName(name string) string {
	return fmt.Sprintf("%s-sharder", name)
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

func (k *kubernetesPlan) deleteService(service *models.ServiceSpecification, client *kubernetes.Clientset) error {
	if service.ShardSpec != nil {
		return k.deleteShardedService(service, client)
	}
	return k.deleteReplicatedService(service, client)
}

func (k *kubernetesPlan) deleteReplicatedService(service *models.ServiceSpecification, client *kubernetes.Clientset) error {
	name := *service.Name
	if k.dryrun {
		glog.Infof("Would have deleted deployment and service %s\n", name)
		return nil
	}
	if err := client.ExtensionsV1beta1().Deployments("default").Delete(name, deleteOptions); err != nil {
		return err
	}
	return client.CoreV1().Services("default").Delete(name, nil)
}

func (k *kubernetesPlan) deleteShardedService(service *models.ServiceSpecification, client *kubernetes.Clientset) error {
	name := *service.Name
	shardName := makeSharderName(name)

	if k.dryrun {
		glog.Infof("Would have deleted deployment and service %s &%s\n", name, shardName)
		return nil
	}

	if err := client.ExtensionsV1beta1().Deployments("default").Delete(shardName, deleteOptions); err != nil {
		return err
	}
	if err := client.AppsV1beta1().StatefulSets("default").Delete(name, deleteOptions); err != nil {
		return err
	}
	if err := client.CoreV1().Services("default").Delete(shardName, deleteOptions); err != nil {
		return err
	}
	return client.CoreV1().Services("default").Delete(name, deleteOptions)
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

func (k *kubernetesPlan) deploy(service *models.ServiceSpecification, client *kubernetes.Clientset) {
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

	k.output(deployment, name+"-deploy")
	if k.dryrun {
		return
	}

	if _, err := client.ExtensionsV1beta1().Deployments("default").Create(deployment); err != nil {
		log.Fatalf(err.Error())
	}
}

func (k *kubernetesPlan) deployStateful(service *models.ServiceSpecification, client *kubernetes.Clientset) {
	name := *service.Name

	deployment := &apps_v1beta1.StatefulSet{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: apps_v1beta1.StatefulSetSpec{
			Replicas:    &service.ShardSpec.Shards,
			ServiceName: name,
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

	k.output(deployment, name+"-stateful-set")
	if !k.dryrun {
		if _, err := client.AppsV1beta1().StatefulSets("default").Create(deployment); err != nil {
			log.Fatalf(err.Error())
		}
	}

	name = makeSharderName(name)
	shardDeployment := &v1beta1.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.DeploymentSpec{
			Replicas: &service.ShardSpec.Shards,
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
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name:  "SHARD_ADDRESSES",
									Value: getShardAddresses(service),
								},
							},
						},
					},
				},
			},
		},
	}

	k.output(shardDeployment, name+"shard-router")
	if k.dryrun {
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

func (k *kubernetesPlan) createLoadBalancedService(service *models.ServiceSpecification, public bool, client *kubernetes.Clientset) {
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

	k.output(svc, name+"-load-balancer")
	if k.dryrun {
		return
	}

	if _, err := client.CoreV1().Services("default").Create(svc); err != nil {
		log.Fatalf(err.Error())
	}
}

func getShardAddresses(service *models.ServiceSpecification) string {
	name := *service.Name
	// TODO: multi-port here?
	port := int(*service.Ports[0].Number)
	pieces := []string{}
	for ix := 0; int32(ix) < service.ShardSpec.Shards; ix++ {
		pieces = append(pieces, fmt.Sprintf("http://%s-%d.%s:%d", name, ix, name, port))
	}
	return strings.Join(pieces, ",")
}

func (k *kubernetesPlan) createStatefulService(service *models.ServiceSpecification, public bool, client *kubernetes.Clientset) {
	name := *service.Name

	statefulSvc := &v1.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: v1.ServiceSpec{
			Ports:     getPorts(service),
			ClusterIP: "None",
			Selector: map[string]string{
				"app": name,
			},
		},
	}

	k.output(statefulSvc, name+"-shards-service")
	if !k.dryrun {
		if _, err := client.CoreV1().Services("default").Create(statefulSvc); err != nil {
			log.Fatalf(err.Error())
		}
	}

	svc := &v1.Service{
		ObjectMeta: meta.ObjectMeta{
			Name: makeSharderName(name),
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": makeSharderName(name),
			},
			Ports: getPorts(service),
		},
	}
	if public {
		svc.Spec.Type = "LoadBalancer"
	} 

	k.output(svc, name+"-shard-router-service")
	if k.dryrun {
		return
	}

	if _, err := client.CoreV1().Services("default").Create(svc); err != nil {
		log.Fatalf(err.Error())
	}
}

func (k *kubernetesCompiler) Compile(opts *CompilerOptions, obj *models.Service) (Plan, error) {
	return &kubernetesPlan{service: obj, clientset: k.clientset, opts: opts}, nil
}

type kubernetesPlan struct {
	opts      *CompilerOptions
	service   *models.Service
	clientset *kubernetes.Clientset
	dryrun    bool
	delete    bool
}

func (k *kubernetesPlan) Dump(dir string) error {
	return fmt.Errorf("unimplemented")
}

func (k *kubernetesPlan) Execute(dryrun bool) error {
	k.dryrun = dryrun
	if k.delete {
		for ix := range k.service.Services {
			if err := k.deleteService(k.service.Services[ix], k.clientset); err != nil {
				return err
			}
		}
		return nil
	}
	service := k.service
	for ix := range service.Services {
		if service.Services[ix].Replicas > 0 && service.Services[ix].ShardSpec != nil {
			return fmt.Errorf("%v: Replicas and shards are mutually exclusive", service.Services[ix].Name)
		}
		public := *service.Serve.Name == *service.Services[ix].Name && service.Serve.Public
		if service.Services[ix].Replicas > 0 {
			k.deploy(service.Services[ix], k.clientset)
			k.createLoadBalancedService(service.Services[ix], public, k.clientset)
		}
		if service.Services[ix].ShardSpec != nil && service.Services[ix].ShardSpec.Shards > 0 {
			k.deployStateful(service.Services[ix], k.clientset)
			k.createStatefulService(service.Services[ix], public, k.clientset)
		}
	}
	return nil
}

func (k *kubernetesCompiler) Delete(opts *CompilerOptions, obj *models.Service) (Plan, error) {
	return &kubernetesPlan{service: obj, clientset: k.clientset, delete: true, opts: opts}, nil
}

func (k *kubernetesPlan) output(obj interface{}, name string) {
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	var stream io.Writer
	if k.opts != nil && len(k.opts.WorkingDirectory) > 0 {
		file := name + ".json"
		iofile, err := os.OpenFile(path.Join(k.opts.WorkingDirectory, file), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			panic(err.Error())
		}
		stream = iofile
		defer iofile.Close()
	} else {
		stream = os.Stdout
	}
	stream.Write(data)
}

func (k *kubernetesCompiler) Logs(svc *models.Service, stdout, stderr io.Writer) error {
	// TODO: choose namespace here?
	namespace := v1.NamespaceDefault

	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	formatPod := func(pod *v1.Pod) string {
		return pod.Name
	}

	formatPodAndContainer := func(pod *v1.Pod, container *v1.Container) string {
		return fmt.Sprintf("%s:%s", formatPod(pod), container.Name)
	}

	tmplString := "{{.Pod.Name}}:{{.Container.Name}} {{.Message}}\n"

	tmpl, err := template.New("line").Parse(tmplString)
	if err != nil {
		return err
	}

	labelSelector := labels.Set{
		"app": *svc.Name,
	}.AsSelectorPreValidated()
	containerPatterns := make([]*regexp.Regexp, 0)
	quiet := false
	var stdoutMutex sync.Mutex
	controller := ktail.NewController(k.clientset, namespace, labelSelector,
		ktail.Callbacks{
			OnEvent: func(event ktail.LogEvent) {
				stdoutMutex.Lock()
				defer stdoutMutex.Unlock()
				tmpl.Execute(os.Stdout, event)
			},
			OnEnter: func(pod *v1.Pod, container *v1.Container) bool {
				if len(containerPatterns) > 0 {
					match := false
					for _, r := range containerPatterns {
						if r.MatchString(pod.Name) || r.MatchString(container.Name) {
							match = true
							break
						}
					}
					if !match {
						return false
					}
				}
				if !quiet {
					yellow.Fprintf(os.Stderr, "==> Detected container %s\n", formatPodAndContainer(pod, container))
				}
				return true
			},
			OnExit: func(pod *v1.Pod, container *v1.Container) {
				if !quiet {
					yellow.Fprintf(os.Stderr, "==> Leaving container %s\n", formatPodAndContainer(pod, container))
				}
			},
			OnError: func(pod *v1.Pod, container *v1.Container, err error) {
				red.Fprintf(os.Stderr, "==> Warning: Error while tailing container %s: %s\n",
					formatPodAndContainer(pod, container), err)
			},
		})
	controller.Run()
	return nil
}

/*
func (k *kubernetesCompiler) Logs(svc *models.Service, stdout, stderr io.Writer) error {
	cmd := []string{"ktail", "-l", "app=" + *svc.Name, "--template", "\"{{.Message}}\\"}
	return executeCommandStreaming(cmd, stdout, stderr)
}
*/
