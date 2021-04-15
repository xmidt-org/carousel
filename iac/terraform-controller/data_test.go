package terraform_controller

const emptyState = `{
  "version": 4,
  "terraform_version": "0.13.4",
  "serial": 1007,
  "lineage": "9abe4427-8f8c-a697-81e5-0bde5a028c73",
  "outputs": {},
  "resources": []
}`

const cleanState = `{
  "version": 4,
  "terraform_version": "0.13.4",
  "serial": 1002,
  "lineage": "9abe4427-8f8c-a697-81e5-0bde5a028c73",
  "outputs": {
    "blueHostnames": {
      "value": [],
      "type": [
        "tuple",
        []
      ]
    },
    "blueVersion": {
      "value": "0.10.0",
      "type": "string"
    },
    "greenHostnames": {
      "value": [
        "carousel-demo-ffdbb6.example.com",
        "carousel-demo-ea9412.example.com"
      ],
      "type": [
        "tuple",
        [
          "string",
          "string"
        ]
      ]
    },
    "greenVersion": {
      "value": "0.10.0",
      "type": "string"
    }
  },
  "resources": [
    {
      "module": "module.green",
      "mode": "data",
      "type": "null_data_source",
      "name": "name",
      "provider": "provider[\"registry.terraform.io/hashicorp/null\"]",
      "instances": [
        {
          "index_key": 0,
          "schema_version": 0,
          "attributes": {
            "has_computed_default": "default",
            "id": "static",
            "inputs": {
              "hostname": "carousel-demo-ffdbb6.example.com"
            },
            "outputs": {
              "hostname": "carousel-demo-ffdbb6.example.com"
            },
            "random": "8387890289259226922"
          }
        },
        {
          "index_key": 1,
          "schema_version": 0,
          "attributes": {
            "has_computed_default": "default",
            "id": "static",
            "inputs": {
              "hostname": "carousel-demo-ea9412.example.com"
            },
            "outputs": {
              "hostname": "carousel-demo-ea9412.example.com"
            },
            "random": "2518257759071851404"
          }
        }
      ]
    },
    {
      "module": "module.green",
      "mode": "managed",
      "type": "random_id",
      "name": "ID",
      "provider": "provider[\"registry.terraform.io/hashicorp/random\"]",
      "instances": [
        {
          "index_key": 0,
          "schema_version": 0,
          "attributes": {
            "b64_std": "/9u2",
            "b64_url": "_9u2",
            "byte_length": 3,
            "dec": "16767926",
            "hex": "ffdbb6",
            "id": "_9u2",
            "keepers": {
              "name": "0.10.0"
            },
            "prefix": null
          },
          "private": "bnVsbA=="
        },
        {
          "index_key": 1,
          "schema_version": 0,
          "attributes": {
            "b64_std": "6pQS",
            "b64_url": "6pQS",
            "byte_length": 3,
            "dec": "15373330",
            "hex": "ea9412",
            "id": "6pQS",
            "keepers": {
              "name": "0.10.0"
            },
            "prefix": null
          },
          "private": "bnVsbA=="
        }
      ]
    }
  ]
}`

const goodList = `module.green.data.null_data_source.name[0]
module.green.data.null_data_source.name[1]
module.green.random_id.ID[0]
module.green.random_id.ID[1]
`
const countGoodList = `module.green[0].data.null_data_source.name
module.green[1].data.null_data_source.name
module.green[0].random_id.ID
module.green[1].random_id.ID
`
