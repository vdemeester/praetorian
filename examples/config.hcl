# Example praetorian configuration.
#
# Used as: command="praetorian run <alias>" in authorized_keys.
# Humans write HCL; Nix can generate the equivalent JSON (same parser).

alias "okinawa-tpm" {
  # Required token prefix, any trailing args (the tool self-restricts).
  allow "borg serve" {}
  allow "nix-store --serve" {}

  # Command with a positional-arg constraint and an exact arg count.
  allow "git-receive-pack" {
    arg {
      pos  = 1
      glob = "/srv/git/*"
    }
    num_args = 1
  }
  allow "git-upload-pack" {
    arg {
      pos  = 1
      glob = "/srv/git/*"
    }
    num_args = 1
  }

  # rsync has dynamic flags, so match on any-arg rather than a fixed position.
  allow "rsync" {
    any_arg = "/srv/backup/*"
    no_arg  = "/srv/backup/.secret/*"
  }
}

# Jump-host-only key.
alias "aomi-tpm" {
  allow "nc" {
    any_arg = "*.internal"
    arg {
      pos  = -1
      glob = "22"
    }
    num_args = 2
  }
}

# Very restricted backup service key.
alias "backup-key" {
  allow "borg serve" {}
}
