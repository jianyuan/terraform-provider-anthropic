# Unrestricted networking — sessions can reach any public host.
resource "anthropic_environment" "default" {
  name = "default"
  config = {
    networking = {
      type = "unrestricted"
    }
  }
}

# Locked-down environment with explicit allow-list and pre-installed packages.
resource "anthropic_environment" "data_analysis" {
  name = "data-analysis"
  config = {
    networking = {
      type                   = "limited"
      allowed_hosts          = ["api.example.com"]
      allow_package_managers = true
    }
    packages = {
      pip = ["pandas", "numpy", "scikit-learn"]
      npm = ["express"]
    }
  }
}
