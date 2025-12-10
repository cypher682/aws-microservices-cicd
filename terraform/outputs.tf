output "ec2_public_ip" {
  description = "Public IP of EC2 instance"
  value       = aws_instance.app.public_ip
}

output "ec2_instance_id" {
  description = "Instance ID"
  value       = aws_instance.app.id
}

output "api_endpoint" {
  description = "API endpoint URL"
  value       = "https://${var.subdomain}.${var.domain_name}"
}

output "nameservers" {
  description = "Route53 nameservers - Update these in GoDaddy"
  value       = aws_route53_zone.main.name_servers
}

output "dynamodb_users_table" {
  description = "DynamoDB users table name"
  value       = aws_dynamodb_table.users.name
}

output "dynamodb_products_table" {
  description = "DynamoDB products table name"
  value       = aws_dynamodb_table.products.name
}

output "ssh_command" {
  description = "SSH command to connect to EC2"
  value       = "ssh -i ~/.ssh/aws-microservices-key ubuntu@${aws_instance.app.public_ip}"
}

output "api_key_ssm_parameter_name" {
  description = "SSM SecureString parameter where the API key is stored"
  value       = aws_ssm_parameter.api_key.name
}
