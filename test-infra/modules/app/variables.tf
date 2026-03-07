variable "app_name" {
  description = "Name of the application"
  type        = string
}

variable "environment" {
  description = "Deployment environment"
  type        = string
}

variable "app_version" {
  description = "Application version"
  type        = string
  default     = "1.0.0"
}

variable "port" {
  description = "Application port"
  type        = number
  default     = 8080
}

variable "replicas" {
  description = "Number of replicas"
  type        = number
  default     = 1
}
