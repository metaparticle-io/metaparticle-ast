package main

import (
	goflag "flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"github.com/metaparticle-io/metaparticle-ast/client"
	"github.com/metaparticle-io/metaparticle-ast/client/services"
	"github.com/metaparticle-io/metaparticle-ast/compiler"
	"github.com/metaparticle-io/metaparticle-ast/models"
	flag "github.com/spf13/pflag"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	port   = flag.Int("port", 8080, "The port to connect to.")
	host   = flag.String("host", "", "The host to connect to")
	file   = flag.StringP("file", "f", "", "The config file to load")
	name   = flag.StringP("name", "n", "", "The name of the service to compile")
	dryrun = flag.Bool("dry-run", false, "If true, only output the execution plan, don't actually enact it.")
	del    = flag.Bool("delete", false, "If true, instead of creating, delete the service.")
	exec   = flag.String("executor", "kubernetes", "The executor to use. Default is 'kubernetes'")
	attach = flag.Bool("attach", false, "If true, then attach to the service in question.")
	deploy = flag.Bool("deploy", true, "If true, deploy or update the service")
)

func main() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	goflag.CommandLine.Parse([]string{})
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

	var cmp compiler.Compiler
	var err error
	switch *exec {
	case "kubernetes":
		cmp, err = compiler.NewKubernetesCompiler()
	case "aci":
		cmp = compiler.NewAciCompiler()
	case "docker":
		cmp = compiler.NewDockerCompiler()
	default:
		glog.Fatalf("Unknown executor: %s", *exec)
	}
	if err != nil {
		glog.Fatalf(err.Error())
	}

	var plan compiler.Plan
	var opts *compiler.CompilerOptions
	if !*dryrun {
		wd, err := os.Getwd()
		if err != nil {
			panic(err.Error())
		}
		dir := path.Join(wd, ".metaparticle")
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err.Error())
		}
		opts = &compiler.CompilerOptions{
			WorkingDirectory: dir,
		}
	}
	if *deploy {
		if *del {
			plan, err = cmp.Delete(opts, obj)
		} else {
			plan, err = cmp.Compile(opts, obj)
		}
	}

	if err != nil {
		glog.Fatalf(err.Error())
	}
	if plan != nil {
		if err := plan.Execute(*dryrun); err != nil {
			glog.Fatalf(err.Error())
		}
	}
	if *attach {
		if err := cmp.Logs(obj, os.Stdout, os.Stderr); err != nil {
			glog.Fatalf(err.Error())
		}
	}
}
