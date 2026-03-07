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

resource "null_resource" "app_server" {
  triggers = {
    name        = var.app_name
    environment = var.environment
    version     = var.app_version
  }
}

resource "random_string" "app_id" {
  length  = 16
  special = false
  upper   = false
}

resource "random_password" "db_password" {
  length           = 32
  special          = true
  override_special = "!@#$%"
}

resource "local_file" "app_config" {
  content = jsonencode({
    app_name    = var.app_name
    environment = var.environment
    version     = var.app_version
    app_id      = random_string.app_id.result
    port        = var.port
    replicas    = var.replicas
  })
  filename = "${path.module}/generated/${var.environment}-${var.app_name}-config.json"
}

resource "null_resource" "health_check" {
  depends_on = [null_resource.app_server]

  triggers = {
    endpoint = "http://localhost:${var.port}/health"
    app_id   = random_string.app_id.result
  }
}
