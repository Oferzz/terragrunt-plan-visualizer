variable "db_name" {
  description = "Database name"
  type        = string
}

variable "environment" {
  description = "Deployment environment"
  type        = string
}

variable "engine" {
  description = "Database engine"
  type        = string
  default     = "postgres"
}

variable "engine_version" {
  description = "Database engine version"
  type        = string
  default     = "15"
}

variable "instance_class" {
  description = "Database instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "allocated_storage" {
  description = "Allocated storage in GB"
  type        = number
  default     = 20
}

variable "port" {
  description = "Database port"
  type        = number
  default     = 5432
}

variable "backup_window" {
  description = "Backup window"
  type        = string
  default     = "03:00-04:00"
}

variable "backup_retention_days" {
  description = "Backup retention period in days"
  type        = number
  default     = 7
}
