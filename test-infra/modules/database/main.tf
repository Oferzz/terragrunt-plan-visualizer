terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

resource "random_password" "master_password" {
  length           = 40
  special          = true
  override_special = "!@#$%^&*"
}

resource "random_string" "db_identifier" {
  length  = 8
  special = false
  upper   = false
}

resource "null_resource" "database_instance" {
  triggers = {
    db_name     = var.db_name
    environment = var.environment
    engine      = var.engine
    instance    = var.instance_class
    storage     = var.allocated_storage
    identifier  = random_string.db_identifier.result
  }
}

resource "null_resource" "database_subnet_group" {
  triggers = {
    name        = "${var.environment}-${var.db_name}-subnet-group"
    environment = var.environment
  }
}

resource "null_resource" "database_parameter_group" {
  triggers = {
    family      = "${var.engine}${var.engine_version}"
    environment = var.environment
  }
}

resource "local_file" "db_connection_info" {
  content = jsonencode({
    host          = "db-${random_string.db_identifier.result}.local"
    port          = var.port
    database      = var.db_name
    engine        = var.engine
    instance_class = var.instance_class
  })
  filename = "${path.module}/generated/${var.environment}-db-connection.json"
}

resource "null_resource" "backup_schedule" {
  triggers = {
    window    = var.backup_window
    retention = var.backup_retention_days
    db_id     = random_string.db_identifier.result
  }
}
