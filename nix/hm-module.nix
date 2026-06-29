# home-manager module for praetorian (user-level install + config).
#
# Used on non-NixOS hosts (e.g. Fedora via standalone home-manager) or to place
# config at ~/.config/praetorian/, praetorian's second config search location.
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
    enable = lib.mkEnableOption "praetorian SSH command restrictor (user)";

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
        Alias -> allow-rule map rendered into ~/.config/praetorian/config.json.
        Each alias maps `command="praetorian run <alias>"` to its allow rules.
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];
    xdg.configFile."praetorian/config.json".source = format.generate "praetorian-config.json" {
      alias = cfg.aliases;
    };
  };
}
