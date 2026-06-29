{
  description = "praetorian — SSH command restrictor (authorized_keys command= gate)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: rec {
        praetorian = pkgs.callPackage ./nix/package.nix { };
        default = praetorian;
      });

      apps = forAllSystems (pkgs: rec {
        praetorian = {
          type = "app";
          program = "${self.packages.${pkgs.stdenv.hostPlatform.system}.praetorian}/bin/praetorian";
        };
        default = praetorian;
      });

      overlays.default = _final: prev: {
        praetorian = prev.callPackage ./nix/package.nix { };
      };

      nixosModules.praetorian = import ./nix/nixos-module.nix self;
      nixosModules.default = self.nixosModules.praetorian;

      homeManagerModules.praetorian = import ./nix/hm-module.nix self;
      homeManagerModules.default = self.homeManagerModules.praetorian;

      checks = forAllSystems (pkgs: {
        # Build the package (compiles everything).
        package = self.packages.${pkgs.stdenv.hostPlatform.system}.praetorian;

        # Run the Go test suite as a flake check.
        gotest = self.packages.${pkgs.stdenv.hostPlatform.system}.praetorian.overrideAttrs (_: {
          doCheck = true;
        });
      });

      formatter = forAllSystems (pkgs: pkgs.nixfmt-rfc-style);

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.golangci-lint
            pkgs.goreleaser
          ];
        };
      });
    };
}
