package restapi

import (
	"sync"

	middleware "github.com/go-openapi/runtime/middleware"
	"github.com/metaparticle-io/metaparticle-ast/models"
	"github.com/metaparticle-io/metaparticle-ast/restapi/operations/services"
)

type Impl struct {
	sync.Mutex
	services map[string]*models.Service
}

// HandleGetServices implements the GetServiceHandler interface
func (i *Impl) HandleListServices(params services.ListServicesParams) middleware.Responder {
	i.Lock()
	defer i.Unlock()
	result := []*models.Service{}
	for _, value := range i.services {
		result = append(result, value)
	}
	return services.NewGetServicesOK().WithPayload(result)
}

// HandleDestroyOne implements the DestroyOneHanlder interface
func (i *Impl) HandleDestroyOne(param services.DeleteServiceParams) middleware.Responder {
	i.Lock()
	defer i.Unlock()
	delete(i.services, param.Name)
	return services.NewDeleteServiceNoContent()
}

// HandleGetOne implements the GetOneHandler interface
func (i *Impl) HandleGetOne(param services.GetServiceParams) middleware.Responder {
	i.Lock()
	defer i.Unlock()
	service, found := i.services[param.Name]
	if !found || service == nil {
		return services.NewGetServiceDefault(404).WithPayload(&models.Error{Code: 404})
	}
	return services.NewGetServiceOK().WithPayload(service)
}

// HandlUpdateOne implements the UpdateOneHandler interface
func (i *Impl) HandleUpdateOne(param services.CreateOrUpdateServiceParams) middleware.Responder {
	i.Lock()
	defer i.Unlock()
	if i.services == nil {
		i.services = map[string]*models.Service{}
	}
	i.services[param.Name] = param.Body
	return services.NewCreateOrUpdateServiceOK().WithPayload(param.Body)
}
