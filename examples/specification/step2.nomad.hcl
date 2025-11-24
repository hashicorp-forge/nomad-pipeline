variable "step_2_end_msg" {}

job "step2" {
  type = "batch"
  group "ubuntu" {

    task "echo" {
      driver = "docker"

      config {
        image   = "ubuntu:24.04"
        command = "echo"
        args    = [var.step_2_end_msg]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}
