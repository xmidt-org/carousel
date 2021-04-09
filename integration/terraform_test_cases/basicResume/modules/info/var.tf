variable "servers" {
  description = "a way around terraform calc error"
  default = "0"
}

variable "application" {
  type = string
  description = "The application of the machine aka Svalinn"
  default = "application"
}
variable "serverVersion" {
  description = "version for the software"
  default = "0.1.1"
}

variable "group" {
  description = "version for the software of group"
  default = "blue"
}