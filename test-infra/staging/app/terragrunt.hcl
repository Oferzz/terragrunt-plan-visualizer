include "root" {
  path = find_in_parent_folders()
}

terraform {
  source = "../../modules/app"
}

inputs = {
  app_name    = "web-api"
  environment = "staging"
  app_version = "2.0.0"
  port        = 8080
  replicas    = 3
}
