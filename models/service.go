// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Service service
// swagger:model service
type Service struct {

	// guid
	// Required: true
	GUID *int64 `json:"guid"`

	// jobs
	Jobs ServiceJobs `json:"jobs"`

	// name
	// Required: true
	// Min Length: 1
	Name *string `json:"name"`

	// serve
	Serve *ServeSpecification `json:"serve,omitempty"`

	// services
	Services ServiceServices `json:"services"`

	// tf jobs
	TfJobs ServiceTfJobs `json:"tfJobs"`
}

// Validate validates this service
func (m *Service) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateGUID(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if err := m.validateServe(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Service) validateGUID(formats strfmt.Registry) error {

	if err := validate.Required("guid", "body", m.GUID); err != nil {
		return err
	}

	return nil
}

func (m *Service) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	if err := validate.MinLength("name", "body", string(*m.Name), 1); err != nil {
		return err
	}

	return nil
}

func (m *Service) validateServe(formats strfmt.Registry) error {

	if swag.IsZero(m.Serve) { // not required
		return nil
	}

	if m.Serve != nil {

		if err := m.Serve.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("serve")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Service) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Service) UnmarshalBinary(b []byte) error {
	var res Service
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
