variable "db_cluster_name" {}
variable "target_sns_topic_arn" {}
variable "lambda_schedule" {
  default = "cron(0 * * * ? *)"
}
variable "region" {}
variable "target_account" {}