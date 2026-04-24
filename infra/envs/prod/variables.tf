variable "gemini_api_key" {
  type      = string
  sensitive = true
}

variable "image_tag" {
  type    = string
  default = "bootstrap"
}
