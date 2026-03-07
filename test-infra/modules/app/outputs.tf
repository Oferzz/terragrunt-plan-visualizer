output "app_id" {
  description = "Generated application ID"
  value       = random_string.app_id.result
}

output "config_path" {
  description = "Path to generated config file"
  value       = local_file.app_config.filename
}

output "db_password" {
  description = "Generated database password"
  value       = random_password.db_password.result
  sensitive   = true
}
