# Getting Started
To get started with Nomad Pipeline you need a Nomad cluster up and running. If you don't have one
yet, you can follow the
[Nomad Get Started Guide](https://www.nomadproject.io/intro/getting-started/install) to set up a
local development cluster. The Nomad clients should have Docker installed if using inline flows or
the appropriate drivers for your tasks when running specification flows.

## Build
To build Nomad Pipeline from source, ensure you have Go installed (tested with go1.25.4). Then,
clone the repository and run the following command:
```bash
make
```

This will compile the Nomad Pipeline binaries and place it in the `bin/` directory. You can then run
the Nomad Pipeline CLI using:
```bash
./bin/nomad-pipeline
```

### Nomad Pipeline Server
To start the Nomad Pipeline server, run the following command with the approrite flags. In
particular, if you are running inline flows, you will need to ensure the `rpc-addr` flag is set to
a routable address the nomad-pipeline-runner will be able to reach.
```bash
./bin/nomad-pipeline server run
```

### Nomad Pipeline Runner
If you you will need a base Docker image with the nomad-pipeline-runner binary installed. You can
build this image using the provided Dockerfile and make target:
```bash
make build-docker-all
```

Once built the runner image should be tagged and pushed to a container registry accessible by
your Nomad clients. The
[jrasell/nomad-pipeline-runner](https://hub.docker.com/repository/docker/jrasell/nomad-pipeline-runner/general)
image can be used for convenience but is provided as-is without any guarantees.
