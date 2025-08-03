module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "5.1.0"

  name = var.cluster_name
  cidr = var.vpc_cidr

  azs = var.azs
  private_subnets = var.private_subnets
  public_subnets = var.public_subnets

  enable_nat_gateway = true
  single_nat_gateway = true
  enable_dns_hostnames = true
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.21"

  cluster_name = var.cluster_name
  cluster_version = "1.29"
  subnet_ids = module.vpc.private_subnets
  vpc_id = module.vpc.vpc_id
  cluster_endpoint_public_access  = true
  enable_irsa = false

  eks_managed_node_groups = {
    default = {
      instance_types = ["t3.medium"]
      min_size = 1
      max_size = 3
      desired_size = 2
    }
  }
}