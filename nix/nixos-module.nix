# NixOS module for praetorian.
#
# Installs the praetorian binary and renders /etc/praetorian/config.json from a
# structured Nix attribute set. praetorian parses JSON natively (hclsimple
# dispatches on the .json extension), so the Nix attrset below *is* the config:
# its shape mirrors the HCL labeled-block schema exactly.
#
#   alias.<name>.allow.<command> = { arg = {pos; glob;}; num_args = N; any_arg; no_arg; };
#
# This module only handles installation + config generation. Wiring the
# `command="praetorian run <alias>"` restriction into a user's authorized_keys
# is the consumer's responsibility.
self:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.praetorian;
  format = pkgs.formats.json { };
in
{
  options.services.praetorian = {
    enable = lib.mkEnableOption "praetorian SSH command restrictor";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.praetorian;
      defaultText = lib.literalExpression "praetorian.packages.\${system}.praetorian";
      description = "The praetorian package to use.";
    };

    aliases = lib.mkOption {
      type = lib.types.attrsOf format.type;
      default = { };
      example = lib.literalExpression ''
        {
          aomi-git.allow = {
            "git-receive-pack" = { arg = { pos = 1; glob = "/home/vincent/git/*"; }; num_args = 1; };
            "git-upload-pack"  = { arg = { pos = 1; glob = "/home/vincent/git/*"; }; num_args = 1; };
          };
        }
      '';
      description = ''
        Alias -> allow-rule map rendered into /etc/praetorian/config.json.
        Each alias maps `command="praetorian run <alias>"` to its allow rules.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];

    environment.etc."praetorian/config.json".source = format.generate "praetorian-config.json" {
      alias = cfg.aliases;
    };
  };
}
