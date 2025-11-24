flow "terraform-provider-nomad" {
  namespace = "jrasell"

  variable "go_version" {
    type    = string
    default = "1.25.4"
  }

  variable "trigger.git_sha" {
    type     = string
    required = true
  }

  inline "test-compile" {

    runner {
      nomad_on_demand {
        namespace = "default"
        image     = "jrasell/nomad-pipeline-runner:latest"

        artifact {
          source      = "git::https://github.com/jrasell/terraform-provider-nomad"
          destination = "terraform-provider-nomad"

          options = {
            ref = "${var.trigger.git_sha}"
          }
        }

        resource {
          cpu    = 1000
          memory = 1024
        }
      }
    }

    step "setup" {
      run = <<EOH
apt-get -q update
apt-get install -y git make wget

cd /tmp
wget -c https://go.dev/dl/go${var.go_version}.linux-amd64.tar.gz
tar -C /usr/local -xf go${var.go_version}.linux-amd64.tar.gz

echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
EOH
    }

    step "run_build" {
      run = <<EOH
source ~/.profile
cd terraform-provider-nomad
go build
EOH
    }

    step "failure_notification" {
      condition = "inline.steps.run_build.status == \"failed\""
      run       = <<EOH
echo "Build failed for commit ${var.trigger.git_sha}"
EOH
    }
  }
}
