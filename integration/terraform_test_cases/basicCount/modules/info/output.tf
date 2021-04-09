output "fqdn" {
  value = data.null_data_source.name.outputs.hostname
  description = "deployed version"
}