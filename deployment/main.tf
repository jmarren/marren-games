terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-west-1"
}

data "local_file" "ask_away_user_data" {
  filename = "${path.module}/user_data.sh"
}

resource "aws_eip" "two" {
  instance = aws_instance.ask_away_instance.id
  domain   = "vpc"
}

resource "aws_eip_association" "eip_assoc" {
  instance_id   = aws_instance.ask_away_instance.id
  allocation_id = aws_eip.two.id
}

resource "aws_security_group" "ask_away_security_group" {
  name        = "allow_web_and_ssh_ask_away"
  description = "Allow SSH inbound traffic"

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "ask_away_instance" {
  ami           = "ami-04c8c3693f870f90f"
  instance_type = "t4g.micro"

  user_data = <<-EOF
  #!/bin/bash
  ${data.local_file.ask_away_user_data.content}
  EOF

  security_groups = [aws_security_group.ask_away_security_group.name]

  iam_instance_profile = aws_iam_instance_profile.ask_away_ec2_profile.name

  tags = {
    Name = "ask_away_instance"
  }
}

resource "aws_s3_bucket_policy" "make_ask_away_s3_read_public" {
  bucket = aws_s3_bucket.ask-away-s3-bucket.id
  policy = data.aws_iam_policy_document.make_ask_away_s3_read_public.json
}

data "aws_iam_policy_document" "make_ask_away_s3_read_public" {
  statement {
    principals {
      type        = "*"
      identifiers = ["*"]
    }

    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion"
    ]

    resources = [
      "${aws_s3_bucket.ask-away-s3-bucket.arn}/public/*"
    ]
  }
}


resource "aws_iam_role" "ask_away_ec2_role" {
  name = "ask_away_ec2_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Sid" : ""
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "ask-away-s3-bucket" {
  bucket = "ask-away-s3-bucket"
}

resource "aws_iam_policy" "allow_access_to_ask_away_s3" {
  name        = "ask_away_S3BucketAccessPolicy"
  description = "Allows access to the ask away s3 bucket"
  policy      = data.aws_iam_policy_document.allow_access_to_ask_away_s3.json
}

data "aws_iam_policy_document" "allow_access_to_ask_away_s3" {
  statement {
    actions = [
      "s3:*",
      "s3-object-lamda:*"
    ]

    resources = [
      aws_s3_bucket.ask-away-s3-bucket.arn,
      "${aws_s3_bucket.ask-away-s3-bucket.arn}/*",
    ]
  }
}


resource "aws_iam_role_policy_attachment" "allow_access_to_ask_away_s3" {
  role       = aws_iam_role.ask_away_ec2_role.name
  policy_arn = aws_iam_policy.allow_access_to_ask_away_s3.arn
}

resource "aws_iam_instance_profile" "ask_away_ec2_profile" {
  name = "ask_away_EC2Profile"
  role = aws_iam_role.ask_away_ec2_role.name
}




resource "aws_iam_policy" "ask_away_secretsmanager_policy" {
  name        = "Ask_Away_SecretsManagerAccessPolicy"
  description = "Allows access to the Secrets Manager service"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect   = "Allow",
        Action   = "secretsmanager:GetSecretValue",
        Resource = var.ask_away_secret_arns
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "ask_away_secrets_policy_attach" {
  role       = aws_iam_role.ask_away_ec2_role.name
  policy_arn = aws_iam_policy.ask_away_secretsmanager_policy.arn
}



variable "ask_away_secret_arns" {
  description = "List of ARNs for required secrets"
  type        = list(string)
}



