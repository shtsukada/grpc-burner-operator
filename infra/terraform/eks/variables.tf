variable "region" {
  description = "AWS region"
  default     = "ap-northeast-1"
}

variable "cluster_name" {
  description = "EKS Cluster name"
  default     = "grpc-observability-cluster"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  default     = "10.0.0.0/16"
}

variable "azs" {
  description = "Availability Zones"
  default     = ["ap-northeast-1a", "ap-northeast-1c"]
}

variable "public_subnets" {
  description = "Public Subnets"
  default     = ["10.0.101.0/24", "10.0.102.0/24"]
}

variable "private_subnets" {
  description = "Private Subnets"
  default     = ["10.0.1.0/24", "10.0.2.0/24"]
}

variable "aws_auth_users" {
  description = "List of IAM users to add to aws-auth configmap"
  type = list(object({
    userarn  = string
    username = string
    groups   = list(string)
  }))
  default = []
}




