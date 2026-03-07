include "root" {
  path = find_in_parent_folders()
}

terraform {
  source = "../../modules/database"
}

inputs = {
  db_name               = "myapp"
  environment           = "dev"
  engine                = "postgres"
  engine_version        = "15"
  instance_class        = "db.t3.medium"
  allocated_storage     = 20
  backup_retention_days = 7
}
