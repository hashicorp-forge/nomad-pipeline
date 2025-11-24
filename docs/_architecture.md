# Architecture
Nomad Pipeline application architecture consists of four core conponents, two of which are optional
depending on the chosen execution model and deployment.

1. **Nomad Pipeline Controller** (required): The central orchestrator that manages flows, runs, and
  triggers. It provides the API for users to interact with the system, handles run logic, and
  coordinates with Nomad to execute jobs. It is packaged as the `nomad-pipeline` binary.

2. **Nomad Cluster** (required): Nomad is used as the underlying job scheduler and distribured
  orchestrator.

3. **Nomad Pipeline Runner** (optional): A lightweight agent that runs inside a scheduled Nomad
  allocation to facilitate the execution of inline flow steps. It provides an environment for
  executing commands defined in inline flows and provides the controller with status and log
  information about the steps via RPC. This component is only required when using inline flow
  execution model and is packaged as the `nomad-pipeline-runner` binary.

4. **Nomad Pipeline UI** (optional): A web-based user interface for visualizing and managing flows,
  runs, and triggers. It interacts with the Nomad Pipeline Controller API to provide a user-friendly
  experience. This component is optional as users can interact with the system via the API or CLI
  and the code is available [here](https://github.com/jrasell/nomad-pipeline-ui).

## Concepts
- **Namespaces**: Logical partitions within Nomad Pipeline to isolate flows, runs, and triggers.

- **Flows**: Definitions of multi-step workflows that specify the sequence of tasks to be executed.
  Flows can be defined using either inline commands or Nomad job specifications.
  
- **Runs**: Instances of flow executions. Each time a flow is triggered, a new run is created to
  track its progress, status, and results.
  
- **Triggers**: Mechanisms to automatically start flow executions based on events such as webhooks
  from git events, or scheduled times.

## Data Storage
Nomad Pipeline has two data storage concepts. The first is the object backend which is used to store
flows, runs, triggers, and namespaces. The second are execution logs which are stored separately due
to their size and access patterns.

### Object Backend
Nomad Pipeline has two supported object backends:
1. **In-Memory Dev Backend**: A simple, non-persistent backend primarily intended for development
  and testing purposes. It stores all data in memory, meaning that all state is lost when the
  scontroller restarts. This is configured using the `--state-backend=dev` server flag.
   
2. **Nomad Variables Backend**: A persistent backend that leverages Nomad's Variables feature to
  store state. This backend provides durability, high availability, and the ability to share state
  across multiple controller instances. This is configured using the `--state-backend=nomad-vars`
  server flag and supports an in-memeory cached mode for improved read performance. This is not
  suitable for high throughput environments as Nomad Variable write operations are serialized and
  replicated via Raft. This means Nomad Pipeline could impact the performance of the Nomad cluster
  if under heavy load. 

### Log Backend
Execution logs are persisted to the Nomad Pipeline Controller's local filesystem. The root path is 
configured by the `--data-dir` server flag. Logs are organized in a hierarchical directory structure:

```
<data-dir>/
    └── logs/
        ├── runs/
        │   └── <namespace>/
        │       └── <run-id>/
        │           └── <step-id>/
        │               └── logs/
        |                  ├── stdout.log
        |                  ├── stderr.log
        └── ...
```

In order to persist logs outside the host filesystem, log shippers can be used to forward logs to
external systems.
