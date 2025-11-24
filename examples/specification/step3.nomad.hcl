variable "step_3_end_msg" {}

job "step3" {
  type = "batch"
  group "ubuntu" {

    task "echo" {
      driver = "docker"

      config {
        image   = "ubuntu:24.04"
        command = "echo"
        args    = [var.step_3_end_msg]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}
