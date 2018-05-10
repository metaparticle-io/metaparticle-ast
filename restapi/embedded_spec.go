// Code generated by go-swagger; DO NOT EDIT.

package restapi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
)

// SwaggerJSON embedded version of the swagger document used at generation time
var SwaggerJSON json.RawMessage

func init() {
	SwaggerJSON = json.RawMessage([]byte(`{
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "schemes": [
    "http"
  ],
  "swagger": "2.0",
  "info": {
    "description": "The metaparticle API",
    "title": "An application for easier distributed application generation",
    "version": "0.0.1"
  },
  "paths": {
    "/services": {
      "get": {
        "tags": [
          "services"
        ],
        "operationId": "listServices",
        "responses": {
          "200": {
            "description": "list the services",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/service"
              }
            }
          }
        }
      }
    },
    "/services/{name}": {
      "get": {
        "tags": [
          "services"
        ],
        "operationId": "getService",
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/service"
            }
          },
          "default": {
            "description": "error",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "put": {
        "tags": [
          "services"
        ],
        "operationId": "createOrUpdateService",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "schema": {
              "$ref": "#/definitions/service"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#/definitions/service"
            }
          },
          "default": {
            "description": "error",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "delete": {
        "tags": [
          "services"
        ],
        "operationId": "deleteService",
        "responses": {
          "204": {
            "description": "Deleted"
          },
          "default": {
            "description": "error",
            "schema": {
              "$ref": "#/definitions/error"
            }
          }
        }
      },
      "parameters": [
        {
          "type": "string",
          "name": "name",
          "in": "path",
          "required": true
        }
      ]
    }
  },
  "definitions": {
    "build": {
      "type": "object",
      "properties": {
        "imageName": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "path": {
          "type": "string"
        }
      }
    },
    "container": {
      "type": "object",
      "required": [
        "image"
      ],
      "properties": {
        "env": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/envVar"
          }
        },
        "image": {
          "type": "string"
        },
        "volumeMounts": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/volumeMount"
          }
        }
      }
    },
    "envVar": {
      "type": "object",
      "required": [
        "name",
        "value"
      ],
      "properties": {
        "name": {
          "type": "string"
        },
        "value": {
          "type": "string"
        }
      }
    },
    "error": {
      "type": "object",
      "required": [
        "message"
      ],
      "properties": {
        "code": {
          "type": "integer",
          "format": "int64"
        },
        "message": {
          "type": "string"
        }
      }
    },
    "jobSpecification": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "containers": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/container"
          }
        },
        "name": {
          "type": "string"
        },
        "replicas": {
          "type": "integer",
          "format": "int32"
        },
        "schedule": {
          "type": "string"
        },
        "volumes": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/volume"
          }
        }
      }
    },
    "serveSpecification": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "name": {
          "type": "string"
        },
        "public": {
          "type": "boolean"
        }
      }
    },
    "service": {
      "type": "object",
      "required": [
        "guid",
        "name"
      ],
      "properties": {
        "guid": {
          "type": "integer",
          "format": "int64"
        },
        "jobs": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/jobSpecification"
          }
        },
        "name": {
          "type": "string",
          "minLength": 1
        },
        "serve": {
          "type": "object",
          "$ref": "#/definitions/serveSpecification"
        },
        "services": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/serviceSpecification"
          }
        }
      }
    },
    "servicePort": {
      "type": "object",
      "required": [
        "number"
      ],
      "properties": {
        "number": {
          "type": "integer",
          "format": "int32"
        },
        "protocol": {
          "type": "string"
        }
      }
    },
    "serviceSpecification": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "containers": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/container"
          }
        },
        "depends": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "ports": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/servicePort"
          }
        },
        "reference": {
          "type": "string"
        },
        "replicas": {
          "type": "integer",
          "format": "int32"
        },
        "shardSpec": {
          "$ref": "#/definitions/shardSpecification"
        },
        "volumes": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/volume"
          }
        }
      }
    },
    "shardSpecification": {
      "type": "object",
      "properties": {
        "fieldPath": {
          "type": "string"
        },
        "shards": {
          "type": "integer",
          "format": "int32"
        },
        "urlPattern": {
          "type": "string"
        }
      }
    },
    "volume": {
      "type": "object",
      "required": [
        "name",
        "persistentVolumeClaim"
      ],
      "properties": {
        "name": {
          "type": "string"
        },
        "persistentVolumeClaim": {
          "type": "string"
        }
      }
    },
    "volumeMount": {
      "type": "object",
      "required": [
        "name",
        "mountPath"
      ],
      "properties": {
        "mountPath": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "subPath": {
          "type": "string"
        }
      }
    }
  }
}`))
}
