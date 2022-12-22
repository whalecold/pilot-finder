the mock pilot agent for test 

### Introduction


### command

#### mock


Usage:
```shell
mock the multi agents connecting to pilot

Usage:
  pilot-finder mock [flags]

Flags:
      --ap string             the address prefix, use for generator the address of workload entry (default "192.168")
  -h, --help                  help for mock
      --namespace string      specify the namespace where se located. (default "default")
      --pilotAddress string   specify the pilot-discovery address (default "127.0.0.1:15010")
      --rate float            specify rate limit to connect to pilot. (default 10)
      --sn int                serviceEntry number (default 1)
      --wn int                workloadEntry number per serviceEntry (default 1)
```

Example:
```shell
pilot-finder mock --namespace=test1 --pilotAddress=127.0.0.1:15010 --rate=5 --sn=10 --wn=2
```

#### xds

Usage:
```shell
mock a the request to the pilot discovery

Usage:
  pilot-finder xds [flags]

Flags:
      --agent string              agent type (default "sidecar")
      --discoveryAddress string   specify the pilot-discovery address (default "127.0.0.1:15010")
  -h, --help                      help for xds
      --ip string                 the address of the agent (default "127.0.0.1")
      --namespace string          the agent location namespace (default "default")
      --podName string            the agent name (default "pod")
      --suffix string             the cluster domain suffix (default "svc.cluster.local")
      --url string                specify the url type of xds, options for lds|rds|cds (default "lds")

```
Example:
```shell
pilot-finder mock --url=lds --namespace=test --ip=127.0.0.1

```

### RoadMap

- [ ] support encryption connection 
- [ ] add static information 
- [ ] support more xDS protocol
- [ ] helm install
- [ ] makefile
