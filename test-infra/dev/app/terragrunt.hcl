include "root" {
  path = find_in_parent_folders()
}

terraform {
  source = "../../modules/app"
}

inputs = {
  app_name    = "web-api"
  environment = "dev"
  app_version = "1.0.0"
  port        = 8080
  replicas    = 2
}
