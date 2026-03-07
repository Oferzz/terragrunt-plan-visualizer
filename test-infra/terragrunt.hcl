# Root terragrunt configuration
# All child modules inherit from this

generate "backend" {
  path      = "backend.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
terraform {
  backend "local" {
    path = "${get_terragrunt_dir()}/terraform.tfstate"
  }
}
EOF
}
