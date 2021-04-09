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

module "green" {
  source = "./modules/info"
  count = var.versionGreenCount

  application = "carousel-demo"
  serverVersion = var.versionGreen
  group = "green"
}
module "blue" {
  source = "./modules/info"
  count = var.versionBlueCount

  application = "carousel-demo"
  serverVersion = var.versionBlue
  group = "blue"
}

output "blueHostnames" {
  value = module.blue.*.fqdn
}
output "greenHostnames" {
  value = module.green.*.fqdn
}

output "blueVersion" {
  value = var.versionBlue
}
output "greenVersion" {
  value = var.versionGreen
}