package compiler

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/metaparticle-io/metaparticle-ast/models"
)

type aciCompiler struct{}

type aciPlan struct {
	service *models.Service
}

type aciDeletePlan struct {
	service *models.Service
}

func NewAciCompiler() Compiler {
	return &aciCompiler{}
}

func (a *aciCompiler) Compile(svc *models.Service) (Plan, error) {
	return &aciPlan{svc}, nil
}

func (a *aciCompiler) Delete(svc *models.Service) (Plan, error) {
	return &aciDeletePlan{svc}, nil
}

func (a *aciPlan) Execute(dryrun bool) error {
	rg := "test"
	for ix := range a.service.Services {
		if err := a.runService(a.service.Services[ix], rg, a.service.Serve, dryrun); err != nil {
			return err
		}
	}
	return nil
}

func (a *aciPlan) runService(spec *models.ServiceSpecification, resourceGroup string, serve *models.ServeSpecification, dryrun bool) error {
	if spec.Replicas > 1 || spec.ShardSpec != nil {
		return fmt.Errorf("ACI runtime doesn't support replication or sharding")
	}
	image := *spec.Containers[0].Image
	cmd := []string{"az", "container", "create", "-g", resourceGroup, "-n", *a.service.Name, "--image", image}

	switch len(spec.Ports) {
	case 0:
		break
	case 1:
		cmd = append(cmd, "--port", strconv.Itoa(int(*spec.Ports[0].Number)))
	default:
		// TODO: Use ACI API directly and fix this...
		return fmt.Errorf("ACI runtime doesn't support multiple ports (for now)")
	}

	if len(spec.Containers[0].Env) > 0 {
		cmd = append(cmd, "-e")
		for _, env := range spec.Containers[0].Env {
			cmd = append(cmd, fmt.Sprintf("%s=%s", *env.Name, *env.Value))
		}
	}

	if serve != nil {
		if *serve.Name == *spec.Name && serve.Public {
			cmd = append(cmd, "--ip-address", "public")
		}
	}

	return executeCommand(cmd, dryrun)
}

func (a *aciPlan) Dump(dir string) error {
	return fmt.Errorf("unimplemented")
}

func (a *aciDeletePlan) Execute(dryrun bool) error {
	rg := "test"
	for ix := range a.service.Services {
		if err := a.deleteService(a.service.Services[ix], rg, dryrun); err != nil {
			return err
		}
	}
	return nil
}

func (a *aciDeletePlan) deleteService(spec *models.ServiceSpecification, resourceGroup string, dryrun bool) error {
	cmd := []string{"az", "container", "delete", "-g", resourceGroup, "-n", *a.service.Name}
	return executeCommand(cmd, dryrun)
}

func (a *aciDeletePlan) Dump(dir string) error {
	return fmt.Errorf("unimplemented")
}

func (k *aciCompiler) Logs(svc *models.Service, stdout, stderr io.Writer) error {
	// TODO: fix this hard-code 'test'
	cmd := []string{"az", "container", "logs", "-g", "test", "-n", *svc.Services[0].Name}
	for {
		if err := executeCommandStreaming(cmd, stdout, stderr); err != nil {
			return err
		}
		time.Sleep(2 * time.Second)
	}
}
