package compiler

import (
	"io"

	"github.com/metaparticle-io/metaparticle-ast/models"
)

type CompilerOptions struct {
	WorkingDirectory string
}

// Compiler is an interface for things that know how to compile metaparticle models
type Compiler interface {
	// Compile a model
	Compile(opts *CompilerOptions, svc *models.Service) (Plan, error)
	// Delete a model
	Delete(opts *CompilerOptions, svc *models.Service) (Plan, error)
	// Tail the logs for an existing service
	Logs(svc *models.Service, stdout, stderr io.Writer) error
}

type Plan interface {
	Execute(dryrun bool) error
	Dump(directory string) error
}
