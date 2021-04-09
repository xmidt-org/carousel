# The UUID for the server name in group
resource "random_id" "ID" {
  count = var.servers
  keepers = {
    # Generate a new id each time we switch to a new AMI id
    name = var.serverVersion
  }

  byte_length = 3
}

data "null_data_source" "name" {
  count = var.servers
  inputs = {
    hostname = "${var.application}-${element(random_id.ID.*.hex, count.index)}"
  }
}