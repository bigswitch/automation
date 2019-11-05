# BCF Terraform EVPC Examples

## Add/Delete EVPCs

```terraform
resource "bcf_evpc" "my-evpc" {
  name = "terraform"
  description = "evpc terraform"
}

resource "bcf_segment" "my-evpc-segment-1" {
  name = "frontend"
  evpc = "${bcf_evpc.my-evpc.name}"
  description = "segment for frontend apps"
  subnets = [
    "10.10.10.1/24",
    "10.10.11.1/24"]
}
```