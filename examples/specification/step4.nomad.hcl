job "step4" {
  type = "batch"
  group "ubuntu" {

    task "echo" {
      driver = "docker"

      config {
        image   = "ubuntu:24.04"
        command = "echo"
        args    = ["ran due to failures of previous steps"]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}
