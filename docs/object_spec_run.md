# Run Specification

A Run represents a single execution instance of a Flow in Nomad Pipeline. Runs track the lifecycle,
status, and results of flow executions, providing observability and management capabilities for
pipeline workflows.

The run object is generated and controlled by the Nomad Pipeline server when a flow is executed and
not directly created or modified by users.

### Run Attributes

- `id` (ULID, required): Unique identifier for the run in ULID format.

- `flow_id` (string, required): ID of the flow being executed.

- `namespace` (string, required): Namespace containing the flow.

- `status` (string, required): Current status of the run. Possible values are:
  - `pending` - Run is created but not yet started
  - `running` - Run is currently executing
  - `success` - Run completed successfully
  - `failed` - Run failed during execution
  - `cancelled` - Run was cancelled by user
  - `skipped` - Run was skipped due to conditions

- `trigger` (string, required): What triggered the run (e.g., "manual", "webhook", "schedule").

- `create_time` (timestamp, required): When the run was created.

- `start_time` (timestamp, optional): When the run started executing.

- `end_time` (timestamp, optional): When the run completed.

- `variables` (map, optional): Variables provided for this run execution.

- `inline_run` (object, optional): Inline execution details. Present if the flow is an inline flow.
  Contains:
  - `job_id` (string): Nomad job ID for the runner job
  - `steps` (array): Array of step execution details, each containing:
    - `id` (string): Step identifier from flow definition
    - `status` (string): Step execution status
    - `exit_code` (number): Exit code of the step execution
    - `start_time` (timestamp): When the step started
    - `end_time` (timestamp): When the step completed

- `spec_run` (object, optional): Specification execution details. Present if the flow is a
  specification flow. Contains:
  - `specs` (array): Array of specification execution details, each containing:
    - `id` (string): Specification identifier from flow definition
    - `nomad_job_id` (string): ID of the Nomad job created
    - `nomad_job_namespace` (string): Nomad namespace where job was created
    - `status` (string): Specification execution status
    - `start_time` (timestamp): When the specification started
    - `end_time` (timestamp): When the specification completed
