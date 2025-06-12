job "step2" {
  type = "batch"
  group "ubuntu" {

    task "echo" {
      driver = "docker"

      config {
        image   = "ubuntu:24.04"
        command = "echo"
        args    = ["step2"]
      }

      resources {
        cpu    = 500
        memory = 256
      }
    }
  }
}
