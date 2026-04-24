output "bucket_name" {
  value = aws_s3_bucket.frontend.id
}

output "cloudfront_url" {
  value = "http://${aws_s3_bucket_website_configuration.frontend.website_endpoint}"
}

output "cloudfront_distribution_id" {
  value = ""
}

output "website_endpoint" {
  value = aws_s3_bucket_website_configuration.frontend.website_endpoint
}
