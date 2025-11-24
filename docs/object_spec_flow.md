# Flow Specification

A Flow defines a pipeline of work to be executed in Nomad Pipeline. Flows can be written in HCL
format and support two execution models: **inline** and **specification**.

### Flow Attributes

`id` (string, required): The identifier for the flow which must be unique within the namespace. This
is specified as a label in HCL format.

`namespace` (string, required): Namespace where the flow will be stored. The namespace must exist
prior to creating the flow.

`variable` (block, optional): Input variable definitions for the flow. Each variable block contains:
  - `name` (string): Variable name (specified as label)
  - `type` (type): Variable type (string, number, bool, list, map, object)
  - `required` (bool): Whether the variable must be provided at runtime (default: false)
  - `default` (any): Default value to use when a value is not provided at runtime

`inline` (block, optional): Inline execution configuration. Contains:
  - `id` (string): Identifier for the inline job (specified as label)
  - `runner` (block): Runner configuration defining execution environment
    - `nomad_on_demand` (block): Nomad on-demand runner configuration
      - `namespace` (string): Nomad namespace for job execution
      - `image` (string): Container image to use for execution. This image must have the
      nomad-pipeline-runner  binary installed.
      - `artifact` (block): Artifacts to fetch before execution
        - `source` (string): Source URL (supports git::, http::, etc.)
        - `destination` (string): Local destination path
        - `options` (map): Additional options (e.g., git ref, credentials) for artifact fetching.
        The map values support HCL template expressions for variable interpolation.
      - `resource` (block): Resource requirements
        - `cpu` (number): CPU allocation in MHz
        - `memory` (number): Memory allocation in MB
  - `step` (block): One or more steps to execute
    - `id` (string): Step identifier (specified as label)
    - `condition` (string): Conditional expression to determine if step should run
    - `run` (string): A shell command to execute. This command is run inside the runner container
    and supports multi-line scripts. The definition supports HCL template expressions for variable
    interpolation.

`specification` (block, optional): Specification-based execution configuration. Contains:
  - `id` (string): Specification identifier (specified as label)
  - `condition` (string): Conditional expression to determine if job should run. This is evaluated
  as a HCL boolean expression.
  - `job` (block): Job specification configuration
    - `name_format` (string, optional): An optional override for the Nomad job name and ID. It
    supports interpolation via HCL expression syntax.
    - `path` (string): Path to Nomad job specification file
    - `variables` (map): Variables to pass to the job specification

> **Note:** A flow must contain either `inline` or `specification` blocks, but not both.

### Examples

A simple inline flow in HCL format:
```hcl
flow "hello-world" {
  namespace = "default"
  
  inline "greeting" {
    runner {
      nomad_on_demand {
        namespace = "default"
        image     = "repo/image:version"
        
        resource {
          cpu    = 500
          memory = 256
        }
      }
    }
    
    step "greet" {
      run = "echo 'Hello, World!'"
    }
  }
}
```

A specification flow with multiple jobs in HCL format:
```hcl
flow "etl-pipeline" {
  namespace = "production"
  
  variable "environment" {
    type    = string
    default = "production"
  }
  
  specification "extract" {
    job {
      name_format = "${nomad_pipeline.run_id}-extract"
      path        = "./extract.nomad.hcl"
      
      variables = {
        env = "environment"
      }
    }
  }
  
  specification "transform" {
    job {
      name_format = "${nomad_pipeline.run_id}-transform"
      path        = "./transform.nomad.hcl"
    }
  }
  
  specification "load" {
    job {
      name_format = "${nomad_pipeline.run_id}-load"
      path        = "./load.nomad.hcl"
    }
  }

  specification "handle-failure" {
    condition = "specifications.extract.status == \"failed\" || specifications.transform.status == \"failed\" || specifications.load.status == \"failed\""
    
    job {
      name_format = "${nomad_pipeline.run_id}-cleanup"
      path        = "./cleanup.nomad.hcl"
    }
  }
}
```
