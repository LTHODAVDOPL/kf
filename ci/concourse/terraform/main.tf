variable "project" {
  type = string
}

variable "k8s_network_selflink" {
  type = string
}

variable "gke_version" {
  type = string
}

provider "google" {
  project     = var.project
  region      = "us-central1"
}

provider "google-beta" {
  project     = var.project
  region      = "us-central1"
}

resource "random_pet" "kf_test" {
}

resource "google_service_account" "kf_test" {
  account_id   = "${random_pet.kf_test.id}"
  display_name = "Managed by Terraform in Concourse"
}

resource "google_project_iam_member" "kf_test" {
  role    = "roles/storage.admin"
  member = "serviceAccount:${google_service_account.kf_test.email}"
}

resource "google_container_cluster" "kf_test" {
  provider = "google-beta"
  name     = "kf-test-${random_pet.kf_test.id}"
  location = "us-central1"

  min_master_version = var.gke_version

  initial_node_count = 1

  master_auth {
    username = ""
    password = ""

    client_certificate_config {
      issue_client_certificate = false
    }
  }

  ip_allocation_policy {
    use_ip_aliases = true
  }

  addons_config {
    istio_config {
      disabled = false
    }
    cloudrun_config {
      disabled = false
    }
    http_load_balancing {
      disabled = false
    }
  }

  node_config {
    machine_type = "n1-standard-4"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    service_account = "${google_service_account.kf_test.email}"

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/userinfo.email",
    ]
  }

  network = var.k8s_network_selflink
}

output "cluster_name" {
  value = google_container_cluster.kf_test.name
}

output "cluster_region" {
  value = google_container_cluster.kf_test.location
}

output "cluster_project" {
  value = var.project
}

output "cluster_version" {
  value = google_container_cluster.kf_test.master_version
}
