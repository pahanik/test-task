module "kubernetes" {
  source  = "ptsgr/kubernetes/openstack"
  version = "0.1.1"
  cloud_user_name = "your username"
  cloud_user_pass = "your password"
  cloud_auth_url    = "openstack url"
  cloud_tenant_id ="openstack project name"
  external_network_name = "external network name"
  cloud_network_id = "internal network id"
  cluster_VIP = "true" # default = false
  #cluster_VIP is cluster virtual ip, by default use keepalived.

  cluster_name = "cluster name"
  vms_image_id = "image id"
  vms_ssh_key = "ssh public key" # use for create key-pair

  master_count = "3" # default = "1"
  master_flavor_id = "master flavor id"

  workers_count = "worker nodes count"
  worker_flavor_id = "worker flavor id"
}
