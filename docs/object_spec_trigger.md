# Trigger Specification

A Trigger defines automated execution of a Flow based on external events or schedules. Triggers
allow flows to be executed automatically in response to webhooks, git events, or time-based
schedules.

### Trigger Attributes

- `id` (string, required): Unique identifier for the trigger. This is specified as a label in HCL
  format.

- `namespace` (string, required): Namespace containing the flow to be triggered.

- `flow` (string, required): ID of the flow to execute when the trigger fires.

- `source` (block, required): Source configuration defining what triggers the flow execution.
  Contains:
  - `id` (string): Source identifier (specified as label)
  - `provider` (string): Provider type (e.g., "git-webhook", "cron")
  - `config` (block): Provider-specific configuration

### Source Providers

#### git-webhook

Used for Git webhook events.

Configuration attributes:
- `provider` (string): Git provider name which currently only supports "github".
- `repository` (string): Repository identifier (e.g., "owner/repo")
- `events` (list): List of events to trigger on which currently only supports ["push"].

#### cron

Used for time-based scheduled execution.

Configuration attributes:
- `crons` (list): List of cron expressions for scheduling

### Examples

A trigger for GitHub push events in HCL format:
```hcl
trigger "terraform-provider-nomad" {
  namespace = "jrasell"
  flow      = "terraform-provider-nomad"

  source "github-push" {
    provider = "git-webhook"

    config {
      provider   = "github"
      repository = "jrasell/terraform-provider-nomad"
      events     = ["push"]
    }
  }
}
```

A scheduled trigger using cron in HCL format:
```hcl
trigger "cron" {
  namespace = "default"
  flow      = "job-cascade"

  source "cron_every_minute" {
    provider = "cron"

    config {
      crons = ["*/1 * * * *"]
    }
  }
}
```

A trigger definition in JSON format:
```json
{
  "id": "terraform-provider-nomad",
  "namespace": "jrasell",
  "flow": "terraform-provider-nomad",
  "source": {
    "id": "github-push",
    "provider": "git-webhook",
    "config": "provider = \"github\"\nrepository = \"jrasell/terraform-provider-nomad\"\nevents = [\"push\"]"
  }
}
```
