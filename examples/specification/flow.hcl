flow "job-cascade" {
  namespace = "default"

  variable "end_msg" {
    type    = string
    default = "step completed successfully"
  }

  specification "step1" {
    job {
      name_format = "${nomad_pipeline.run_id}-step1"
      path        = "./examples/specification/step1.nomad.hcl"

      variables = {
        step_1_end_msg = "end_msg"
      }
    }
  }

  specification "step2" {
    job {
      name_format = "${nomad_pipeline.run_id}-step2"
      path        = "./examples/specification/step2.nomad.hcl"

      variables = {
        step_2_end_msg = "end_msg"
      }
    }
  }

  specification "step3" {
    job {
      name_format = "${nomad_pipeline.run_id}-step3"
      path        = "./examples/specification/step3.nomad.hcl"

      variables = {
        step_3_end_msg = "end_msg"
      }
    }
  }

  specification "step4-conditional" {
    condition = "specifications.step1.status == \"failed\" || specifications.step2.status == \"failed\" || specifications.step3.status == \"failed\""

    job {
      path = "./examples/specification/step4.nomad.hcl"
    }
  }
}
