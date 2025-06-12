id = "terraform-provider-nomad"

job "test-compile" {

  nomad_namespace = "default"

  resource {
    cpu    = 1000
    memory = 1024
  }

  artifact {
    source      = "git::https://github.com/hashicorp/terraform-provider-nomad"
    destination = "/terraform-provider-nomad"
  }

  step "setup" {
    run = <<EOH
#!/usr/bin/env bash

apt-get -q update
apt-get install -y git make wget

cd /tmp
wget -c https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
tar -C /usr/local -xf go1.24.0.linux-amd64.tar.gz

echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
EOH
  }
  
  step "run_build" {
    run = <<EOH
#!/usr/bin/env bash

source ~/.profile
cd terraform-provider-nomad
go build
EOH
  }
}
