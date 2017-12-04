# Metaparticle Compiler

This compiler transforms the metaparticle domain-specific language into execution
on various backends (e.g. Kubernetes)

# Status

This documentation is woefully incomplete, but fortunately, this tool is an implementation
detail.

# Schema
The OpenAPI Schema for the metaparticle DSL can be found [here](api.yaml)

# Specification examples

## Simple server

```json
{
    "name": "server",
    "guid": 1234567, 
    "services": [ 
        {
                "name": "server",
            "replicas": 4,
            "containers": [
                { "image": "nginx" }
            ],
            "ports": [{
                "number": 80
            }]
        }
    ],
    "serve": {
        "name": "server",
        "public": true
    }
}
```

## Sharded server

```json
{
    "name": "name",
    "guid": 1234567, 
    "services": [ 
        {
            "name": "server",
            "shardSpec": {
                "shards": 4,
                "urlPattern": "user/(.*)/"
            },
            "containers": [
                { "image": "brendanburns/node-hostname" }
            ],
            "ports": [{
                "number": 80
            }]
        }
    ],
    "serve": {
        "name": "server",
        "public": true
    }
}
```


# Command line examples

```sh
# Run a file in kubernetes
mp-compiler -f metaparticle-spec.json

# Attach to the logs, but don't re-deploy
mp-compiler -f metaparticle-spec.json --attach=true --deploy=false

# Tear down an existing service
mp-compiler -f metaparticle-spec.json --deploy=false --delete=true
```

