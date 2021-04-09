output "name" {
  value = data.null_data_source.name.*.outputs.hostname
  description = "deployed name"
}

output "count" {
  value = var.servers
  description = "deployed version"
}