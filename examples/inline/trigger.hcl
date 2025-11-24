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
