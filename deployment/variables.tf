variable "client_id" {
  type        = "string"
  description = "Client ID"
}

variable "client_secret" {
  type        = "string"
  description = "Client secret."
}

variable "resource_group_name" {
  description = "Resource group name"
  type        = "string"
}

variable "resource_group_location" {
  description = "Resource group location"
  type        = "string"
}

variable "docker_root" {
  description = "The root url for docker images 'docker_root/ion-management'"
  default     = "dotjson"
}

variable "docker_tag" {
  description = "The docker tag to use for image, this is normally the ion version"
  default     = "latest"
}

variable "batch_dedicated_node_count" {
  description = "Number of dedicated nodes to provision in the batch pool"
  default     = "1"
}

variable "batch_low_priority_node_count" {
  description = "Number of low priority nodes to provision in the batch pool"
  default     = "3"
}

variable "aks_node_count" {
  description = "Number of kubernetes nodes to provision in the AKS cluster"
  default     = "3"
}
