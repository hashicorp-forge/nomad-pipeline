id = "job-cascade"

job "step1" {
  specification {
    path = "./demo/specification/step1.nomad.hcl"
  }
}

job "step2" {
  specification {
    path = "./demo/specification/step2.nomad.hcl"
  }
}

job "step3" {
  specification {
    path = "./demo/specification/step3.nomad.hcl"
  }
}
