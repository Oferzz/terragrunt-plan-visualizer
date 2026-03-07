output "db_identifier" {
  description = "Database instance identifier"
  value       = random_string.db_identifier.result
}

output "connection_info_path" {
  description = "Path to connection info file"
  value       = local_file.db_connection_info.filename
}

output "master_password" {
  description = "Database master password"
  value       = random_password.master_password.result
  sensitive   = true
}
