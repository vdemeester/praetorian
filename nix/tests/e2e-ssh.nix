# End-to-end SSH test for praetorian (issue #24, OpenSSH path).
#
# Uses the NixOS test framework: a `server` node runs sshd + the praetorian
# NixOS module; a `client` node generates an ephemeral keypair at runtime, the
# server authorizes its public key with a forced command="praetorian run ci"
# restriction, and the client drives real git operations plus adversarial
# commands, asserting allowed-vs-denied. This also integration-tests the NixOS
# module and config.json auto-discovery (the command= passes no --config).
#
# No private key is committed: the keypair is created inside the test VM.
{ self, pkgs }:
pkgs.testers.runNixOSTest {
  name = "praetorian-e2e-ssh";

  nodes.server =
    { pkgs, ... }:
    {
      imports = [ self.nixosModules.praetorian ];

      services.openssh.enable = true;
      environment.systemPackages = [ pkgs.git ];

      services.praetorian = {
        enable = true;
        aliases.ci.allow = {
          "git-receive-pack".arg = {
            pos = 1;
            glob = "/srv/git/*";
          };
          "git-upload-pack".arg = {
            pos = 1;
            glob = "/srv/git/*";
          };
        };
      };

      # The git user's key is authorized at runtime (ephemeral key), so no key
      # material lives in the Nix store. sshd reads ~/.ssh/authorized_keys.
      users.users.git = {
        isNormalUser = true;
        home = "/home/git";
      };

      # Bare-repo parent directory, owned by the git user.
      systemd.tmpfiles.rules = [ "d /srv/git 0755 git users -" ];
    };

  nodes.client =
    { pkgs, ... }:
    {
      environment.systemPackages = [
        pkgs.git
        pkgs.openssh
      ];
    };

  testScript = ''
    start_all()
    server.wait_for_unit("sshd.service")

    # Initialise a bare repo as the git user.
    server.succeed("runuser -u git -- git init --bare /srv/git/test.git")

    # Generate an ephemeral client keypair inside the VM (nothing committed).
    client.succeed("mkdir -p /root/.ssh")
    client.succeed("ssh-keygen -t ed25519 -N ''' -f /root/.ssh/id_ed25519")
    pubkey = client.succeed("cat /root/.ssh/id_ed25519.pub").strip()

    # Authorize it on the server with the forced praetorian command.
    opts = "no-pty,no-agent-forwarding,no-port-forwarding,no-X11-forwarding"
    server.succeed("install -d -m700 -o git -g users /home/git/.ssh")
    server.succeed(
        f"printf 'command=\"praetorian run ci\",{opts} %s\\n' '{pubkey}' "
        "> /home/git/.ssh/authorized_keys"
    )
    server.succeed(
        "chown git:users /home/git/.ssh/authorized_keys && "
        "chmod 600 /home/git/.ssh/authorized_keys"
    )

    client.succeed("ssh-keyscan server > /root/.ssh/known_hosts 2>/dev/null")

    ssh = "ssh -i /root/.ssh/id_ed25519 -o IdentitiesOnly=yes"
    git_ssh = f"GIT_SSH_COMMAND='{ssh}'"

    # ALLOWED: clone (git-upload-pack on /srv/git/*).
    client.succeed(f"{git_ssh} git clone ssh://git@server/srv/git/test.git /tmp/clone")

    # ALLOWED: push (git-receive-pack on /srv/git/*).
    client.succeed(
        "cd /tmp/clone && "
        "git config user.email a@b.c && git config user.name t && "
        "git commit --allow-empty -m e2e && "
        f"{git_ssh} git push origin HEAD:master"
    )

    # DENIED: an arbitrary command is rejected by the gate.
    client.fail(f"{ssh} git@server id")

    # DENIED: a path outside the glob is rejected.
    client.fail(f"{ssh} git@server git-upload-pack /etc/passwd")

    # DENIED (no-shell invariant): shell metacharacters are inert bytes, not
    # interpreted. "id; ls" tokenizes to ["id;", "ls"]; "id;" is not an allowed
    # command, so it is denied rather than a shell running `id` then `ls`.
    client.fail(f"{ssh} git@server 'id; ls'")
  '';
}
