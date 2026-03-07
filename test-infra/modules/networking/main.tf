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

resource "random_id" "vpc_id" {
  byte_length = 8
}

resource "null_resource" "vpc" {
  triggers = {
    cidr_block   = var.vpc_cidr
    environment  = var.environment
    vpc_id       = random_id.vpc_id.hex
  }
}

resource "null_resource" "public_subnets" {
  count = length(var.public_subnet_cidrs)

  triggers = {
    cidr_block = var.public_subnet_cidrs[count.index]
    az         = var.availability_zones[count.index % length(var.availability_zones)]
    vpc_id     = random_id.vpc_id.hex
  }
}

resource "null_resource" "private_subnets" {
  count = length(var.private_subnet_cidrs)

  triggers = {
    cidr_block = var.private_subnet_cidrs[count.index]
    az         = var.availability_zones[count.index % length(var.availability_zones)]
    vpc_id     = random_id.vpc_id.hex
  }
}

resource "null_resource" "internet_gateway" {
  triggers = {
    vpc_id      = random_id.vpc_id.hex
    environment = var.environment
  }
}

resource "null_resource" "nat_gateway" {
  triggers = {
    vpc_id      = random_id.vpc_id.hex
    environment = var.environment
  }
}

resource "null_resource" "security_group_app" {
  triggers = {
    name        = "${var.environment}-app-sg"
    vpc_id      = random_id.vpc_id.hex
    ingress_port = "8080"
  }
}

resource "null_resource" "security_group_db" {
  triggers = {
    name        = "${var.environment}-db-sg"
    vpc_id      = random_id.vpc_id.hex
    ingress_port = "5432"
  }
}

resource "local_file" "network_config" {
  content = jsonencode({
    vpc_id              = random_id.vpc_id.hex
    vpc_cidr            = var.vpc_cidr
    public_subnet_cidrs = var.public_subnet_cidrs
    private_subnet_cidrs = var.private_subnet_cidrs
    environment         = var.environment
  })
  filename = "${path.module}/generated/${var.environment}-network-config.json"
}
