package compiler

import (
	"fmt"
	"io"

	"github.com/metaparticle-io/metaparticle-ast/models"
)

type dockerCompiler struct{}

type dockerPlan struct {
	service *models.Service
}

type dockerDeletePlan struct {
	service *models.Service
}

func NewDockerCompiler() Compiler {
	return &dockerCompiler{}
}

func (d *dockerCompiler) Compile(svc *models.Service) (Plan, error) {
	return &dockerPlan{svc}, nil
}

func (d *dockerCompiler) Delete(svc *models.Service) (Plan, error) {
	return &dockerDeletePlan{svc}, nil
}

func (d *dockerPlan) Execute(dryrun bool) error {
	for ix := range d.service.Services {
		if err := d.runService(d.service.Services[ix], d.service.Serve, dryrun); err != nil {
			return err
		}
	}
	return nil
}

func (d *dockerPlan) runService(spec *models.ServiceSpecification, serve *models.ServeSpecification, dryrun bool) error {
	if spec.Replicas > 1 || spec.ShardSpec != nil {
		return fmt.Errorf("docker runtime doesn't support replication or sharding")
	}
	image := *spec.Containers[0].Image
	cmd := []string{"docker", "run", "--name", *spec.Name, "-d"}

	for _, port := range spec.Ports {
		cmd = append(cmd, "-p", fmt.Sprintf("%d:%d", *port.Number, *port.Number))
	}

	if len(spec.Containers[0].Env) > 0 {
		cmd = append(cmd, "-e")
		for _, env := range spec.Containers[0].Env {
			cmd = append(cmd, fmt.Sprintf("%s=%s", *env.Name, *env.Value))
		}
	}

	cmd = append(cmd, image)

	return executeCommand(cmd, dryrun)
}

func (d *dockerPlan) Dump(dir string) error {
	return fmt.Errorf("unimplemented")
}

func (d *dockerDeletePlan) Execute(dryrun bool) error {
	for ix := range d.service.Services {
		if err := d.deleteService(d.service.Services[ix], dryrun); err != nil {
			return err
		}
	}
	return nil
}

func (d *dockerDeletePlan) deleteService(spec *models.ServiceSpecification, dryrun bool) error {
	cmd := []string{"docker", "rm", "-f", *spec.Name}
	return executeCommand(cmd, dryrun)
}

func (d *dockerDeletePlan) Dump(dir string) error {
	return fmt.Errorf("unimplemented")
}

func (d *dockerCompiler) Logs(svc *models.Service, stdout, stderr io.Writer) error {
	cmd := []string{"docker", "logs", *svc.Services[0].Name}
	return executeCommandStreaming(cmd, stdout, stderr)
}
