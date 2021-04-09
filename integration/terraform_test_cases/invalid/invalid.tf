variable "versionBlueCount" {
  description = "Number of instances for version Blue"
  default = 0
}

variable "versionBlue" {
  description = "version for the software of group Blue"
  default = "0.0.0"
}

variable "versionGreenCount" {
  description = "Number of instances for version Green"
  default = 2
}

variable "versionGreen" {
  description = "version for the software of group Green"
  default = "0.1.0"
}

output "green" {
  value = var.versionGreen
}

output "blue" {
  value = var.versionBlue
}