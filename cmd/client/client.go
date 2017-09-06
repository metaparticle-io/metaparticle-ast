package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/glog"
	"github.com/metaparticle-io/metaparticle-ast/client"
	"github.com/metaparticle-io/metaparticle-ast/client/services"
	"github.com/metaparticle-io/metaparticle-ast/models"
	flag "github.com/spf13/pflag"
)

var (
	port = flag.Int("port", 8080, "The port to connect to.")
	host = flag.String("host", "localhost", "The host to connect to")
	file = flag.StringP("file", "f", "", "The config file to load")
)

func main() {
	// TODO: fix this for glog, see: https://github.com/kubernetes/kubernetes/pull/3342/files
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)

	tc := client.DefaultTransportConfig().WithHost(addr)
	c := client.NewHTTPClientWithConfig(nil, tc)

	if len(*file) != 0 {
		obj := &models.Service{}
		bytes, err := ioutil.ReadFile(*file)
		if err != nil {
			glog.Fatalf("Couldn't read file: %v", err)
		}
		if err := obj.UnmarshalBinary(bytes); err != nil {
			glog.Fatalf("Couldn't parse file: %v", err)
		}
		glog.Infof("Parsed: %#v", obj)

		params := &services.GetServiceParams{Name: *obj.Name}
		params = params.WithTimeout(5 * time.Second)
		resp, err := c.Services.GetService(params)
		if err != nil {
			modelErr := err.(*services.GetServiceDefault).Payload
			if modelErr.Code != 404 {
				glog.Fatalf("Failed to get service: %#v", modelErr)
			}
			glog.Infof("Didn't find service.")
			updateParams := services.NewCreateOrUpdateServiceParamsWithTimeout(5 * time.Second)
			updateParams.Body = obj
			updateParams.Name = *obj.Name
			_, err := c.Services.CreateOrUpdateService(updateParams)
			if err != nil {
				glog.Fatalf("Failed to update: %#v", err.(*services.CreateOrUpdateServiceDefault).Payload)
			}
			return
		}
		glog.Infof("Found: %#v", resp)
	}

	resp, err := c.Services.GetServices(nil)
	if err != nil {
		glog.Fatalf("Error: %v\n", err)
	}
	glog.Infof("[")
	for _, obj := range resp.Payload {
		bytes, err := obj.MarshalBinary()
		if err != nil {
			glog.Warningf("Unexpected error: %v", err)
			continue
		}
		glog.Infof("%s\n", string(bytes))
	}
	glog.Infof("]")
}
