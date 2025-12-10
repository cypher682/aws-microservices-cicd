variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "aws-microservices-cicd"
}

variable "domain_name" {
  description = "Domain name"
  type        = string
  default     = "cipherpol.xyz"
}

variable "subdomain" {
  description = "Subdomain for API"
  type        = string
  default     = "api"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "ssh_public_key_path" {
  description = "Path to SSH public key"
  type        = string
  default     = "~/.ssh/aws-microservices-key.pub"
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access EC2"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}
