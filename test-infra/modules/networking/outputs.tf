output "vpc_id" {
  description = "VPC identifier"
  value       = random_id.vpc_id.hex
}

output "network_config_path" {
  description = "Path to network config file"
  value       = local_file.network_config.filename
}
