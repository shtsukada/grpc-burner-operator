output "cluster_name" {
  description = "EKS cluster name"
  value = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the cluster"
  value = module.eks.cluster_security_group_id
}

output "vpc_id" {
  description = "VPC ID"
  value = module.vpc.vpc_id
}
