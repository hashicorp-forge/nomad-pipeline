job "step1" {
  type = "batch"
  group "ubuntu" {

    task "echo" {
      driver = "docker"

      config {
        image   = "ubuntu:24.04"
        command = "echo"
        args    = ["step1"]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}
