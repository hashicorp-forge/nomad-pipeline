# Nomad Pipeline
An experimental workflow orchestration system built on top of
[HashiCorp Nomad](https://www.nomadproject.io/). Nomad Pipeline allows users to define, schedule,
and monitor multi-step workflows (flows) that are executed as Nomad jobs.

> **Note**
This project is the result of two hack week projects and is unsupported and unofficial.

### Docs
The documentation for Nomad Pipeline can be found within the [docs](./docs) directory. The
[examples](./examples) directory contains example flow definitions and Nomad job files.

### What Is Bad?
* Run logs are not periodically removed from the data directory
* Run state objects are not periodically garbage collected

### What Could It Do?
* Shared state across flow jobs, provided by Dynamic Host Volumes or suchlike
* Offer runners via the Nomad [libvirt driver](https://github.com/hashicorp/nomad-driver-virt)
* Identify and automatically upload end of pipeline artifacts
* Persistent pipeline runners that accept flow runs over a long-lived connection
* Storage backend HA locking
