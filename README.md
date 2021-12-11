stunning-octo-enigma
====================

`stunning-octo-enigma` (`enigma` for short) is a dependency aware autoscaler. It makes use of `istio` as a service mesh along with `kiali` to maintain a graph of dependencies between the microservices of the deployed application. It uses the kubernetes metrics server to fetch pod and deployment resource metrics.

Along with the graph and resource metrics, `enigma` also takes in a configuration file which defines the desired overall throughput of the application along with service wise resource thresholds. When the application is in violation of either factors (application throughput or per service resource utilizations), a scaling cycle begins.

`enigma` provides better scaling as when services are determined to be scaled, corresponding downstream services are also scaled (if needed) to avoid bottleneck shifting. These downstream services are validated if they require scaling by estimating queue lengths at each service and comparing them against pre computed thresholds.

Repository organization:
------------------------

The packages of the repository are organized as follows

-	**Load Generator** (`pkg/load-generator`\): This package is responsible for load testing and generating load that simulates user interaction with the application. It internally uses the `vegeta` HTTP load testing package.

-	**Trigger** (`pkg/trigger`\): This package is responsible for the trigger and provides a trigger client to be used. It triggers scaling cycles and handles initiation of scaling decisions.

-	**Metric Scraper** (`pkg/metricscraper`\): This package provides a client to interact with the kubernetes metric server. It is responsible for fetching resource utilizations which influence scaling decisions.

-	**Kiali Client** (`pkg/kiali`\): This package provides a client to interact with kiali API. It provides methods to get graphs of the application with varied flexibility of information to be fetched in the graph.

-	**Kubernetes Client** (`pkg/k8s`\): This package provides crucial methods to interact with the kubernetes API server and perform scaling.

-	**Configuration** (`pkg/config`\): This package handles configuration management and is solely responsible for setting thresholds for the trigger client to work on.

How to run
----------

1.	Make sure that `istio` and `kiali` are set up by following the setup manual:

	1.	Download Istio Binary

		```bash
		curl -L https://istio.io/downloadIstio | sh -
		```

	2.	Move into package directory and update `$PATH`

		```bash
		cd istio-1.12.1
		export PATH=$PWD/bin:$PATH
		```

	3.	Install istio and allow envoy sidecar injection

		```bash
		istioctl install --set profile=demo -y
		kubectl label namespace default istio-injection=enabled
		```

	4.	Setup Kiali by installing the addons

		```bash
		kubectl apply -f samples/addons
		kubectl rollout status deployment/kiali -n istio-system
		```

2.	Additionally, metric server should be deployed using the command (make sure to be in the project root before running commands):

	```bash
	kubectl apply -f deploy/metric-server.yaml
	```

3.	Deploy the application onto the cluster using the command (below command deploys tea store):

	```bash
	kubectl apply -f deploy/teastore-clusterip.yaml
	```

4.	Update the host in the configuration file with the endpoint to be hit while generating load. The configuration file also must contain details such as load parameters and CPU/Memory thresholds that should be met for the services that need to be autoscaled in the application. The structure of a config file is as follows:

	```json
	{
	    "kialiHost": {
	        "host": "kiali endpoint IP or domain",
	        "port": 20001
	    },
	    "appHost": {
	        "host": "application endpoint IP or domain",
	        "port": 30080
	    },
	    "thresholds": {
	        "resourceThresholds": {
	            "service 1": {
	                "cpu": 200
	            },
	            "service 2": {
	        "cpu": 200,
	        "memory": 200
	            },
	            "service 3": {
	        "memory": 200
	      }
	        },
	        "throughput": 100000
	    },
	    "loadParameters": {
	        "distributionType": "inc",
	        "steps": 50,
	        "duration": 10,
	        "workers": 50,
	        "minRate": 100,
	        "maxRate": 1000
	    },
	    "namespaces": [
	        "default"
	    ]
	}
	```

5.	Building the binary (requires `go` to be installed).

	```bash
	go build enigma.go
	```

6.	The binary can then be run using the usage defined in the next section. Example:

	-	Run load on application:

		```
		./engima -l -f config.json
		```

	-	Run the dependency aware autoscaler along with load on the application

		```
		./enigma -s -f config.json
		```

Usage
-----

```
Usage of ./enigma:
  -f, --file string      Path to config file or directory (default "config.json")
  -l, --load             Load test application
  -q, --logq             Log queue lengths and create json with threshold queue lengths for each deployment of application (use alongside l)
  -r, --logrc            Log replica counts of application deployments to file (use alongside l or s)
  -p, --logreq           Log request rate from load tester (use alongside l or s)
  -t, --logth            Log e2e throughput of application (use alongside l or s)
  -s, --scale-and-load   Running scaler and simultaneously load test application

```
