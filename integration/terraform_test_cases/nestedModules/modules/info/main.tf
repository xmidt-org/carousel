module "name" {
  source = "./modules/name"
  servers = var.servers
}


data "null_data_source" "fqdn" {
  count = var.servers
  inputs = {
    hostname = "${var.application}-${element(module.name.name, count.index)}.example.com"
  }
}