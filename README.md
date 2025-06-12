# Nomad Pipeline
An experimental (3 day hack project) Nomad Pipeline application. It supports
chaining of Nomad job specifications as well as inline step functions.

The project is unsupported and unofficial. The project intention is to show what
is possible and garner interest.

### Demo
To run any of the three included demos, you will need to have
[Nomad](https://developer.hashicorp.com/nomad) and
[Docker](https://www.docker.com/) available and running.

Clone the repository and start a nomad-pipeline server:
```console
git clone
cd nomad-pipeline
make
./bin/nomad-pipeline server run
```

Create the flow specification you wish to run. There are currently three
available within the [demo](./demo) directory:
```console
./bin/nomad-pipeline flow create <path_to_spec>
```

Run the flow you just created:
```console
./bin/nomad-pipeline flow run <flow_id>
```

You can then explore the `nomad-pipeline run` sub-commands to discover the run
information.

### What Is Bad?
* Run logs are not periodically removed from the data directory
* There is no way to alter the server configuration parameters

### What Could It Do?
* Flow specification input parameters which are provided on run trigger
* Flow specification interpolation, provided by [HIL](https://github.com/hashicorp/hil)
* Flow specification conditional logic; "if this step fails, do this, else continue"
* Flow specification HCL2 functions
* Shared state across flow jobs, provided by Dynamic Host Volumes
* Offer runners via the Nomad [libvirt driver](https://github.com/hashicorp/nomad-driver-virt)
* Provide a UI to front the HTTP API
* A state backend which supports persistence
* Identify and automatically upload end of pipeline artifacts
