# BCF Terraform EVPC Examples

## Add/Delete EVPCs

### Step 1: Create a evpc.tf file

### Step 2: Copy & Paste the following content: 

To create an EVPC called "my-evpc" with following properties:
EVPC Name = "my-evpc"
Segment Name = "frontend"
Interface Segment IP = "10.10.10.1/24"

Copy & paste the following code snippet to "evpc.tf"

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
    "10.10.10.1/24"]
}
```

### Step 3: Do a terraform plan to 
terraform plan