// Code generated by go-swagger; DO NOT EDIT.

package services

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/metaparticle-io/metaparticle-ast/models"
)

// GetServicesOKCode is the HTTP code returned for type GetServicesOK
const GetServicesOKCode int = 200

/*GetServicesOK list the services

swagger:response getServicesOK
*/
type GetServicesOK struct {

	/*
	  In: Body
	*/
	Payload []*models.Service `json:"body,omitempty"`
}

// NewGetServicesOK creates GetServicesOK with default headers values
func NewGetServicesOK() *GetServicesOK {
	return &GetServicesOK{}
}

// WithPayload adds the payload to the get services o k response
func (o *GetServicesOK) WithPayload(payload []*models.Service) *GetServicesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get services o k response
func (o *GetServicesOK) SetPayload(payload []*models.Service) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetServicesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		payload = make([]*models.Service, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}

}