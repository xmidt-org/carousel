# The UUID for the server name in group
resource "random_id" "ID" {
  keepers = {
    # Generate a new id each time we switch to a new AMI id
    name = var.serverVersion
  }

  byte_length = 3
}

data "null_data_source" "name" {
  inputs = {
    hostname = "${var.application}-${random_id.ID.hex}.example.com"
  }
}