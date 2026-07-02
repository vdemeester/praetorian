{
  lib,
  buildGoModule,
  # Build metadata, normally threaded in from the flake so tagged builds
  # report accurate values. Defaults keep `nix-build`/`callPackage` working.
  commit ? "unknown",
  date ? "1970-01-01T00:00:00Z",
}:
let
  fs = lib.fileset;
in
buildGoModule (finalAttrs: {
  pname = "praetorian";
  # Single source of truth: the VERSION file at the repo root.
  version = lib.strings.trim (builtins.readFile ../VERSION);

  # Only the files that affect the build — keeps the store path stable when docs
  # or CI config change.
  src = fs.toSource {
    root = ../.;
    fileset = fs.unions [
      ../go.mod
      ../go.sum
      ../main.go
      ../internal
      ../version
    ];
  };

  vendorHash = "sha256-V+Bz7oS9KHtI2w7Zes9ndZhVXquGEKxVIcGmd+Up2lY=";

  env.CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X github.com/vdemeester/praetorian/version.Version=${finalAttrs.version}"
    "-X github.com/vdemeester/praetorian/version.Commit=${commit}"
    "-X github.com/vdemeester/praetorian/version.Date=${date}"
  ];

  meta = {
    description = "SSH command restrictor (authorized_keys command= gate)";
    homepage = "https://github.com/vdemeester/praetorian";
    license = lib.licenses.gpl3Only;
    mainProgram = "praetorian";
    platforms = lib.platforms.unix;
  };
})
