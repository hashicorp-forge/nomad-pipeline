trigger "job-cascade-scheduled" {

  namespace = "default"
  flow      = "job-cascade"

  source "cron_every_minute" {
    provider = "cron"

    config {
      crons = ["*/1 * * * *"]
    }
  }
}
