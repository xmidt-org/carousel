output "fqdn" {
  value = data.null_data_source.name.*.outputs.hostname
  description = "deployed version"
}

output "count" {
  value = var.servers
  description = "deployed version"
}