{
  description = "go-codec — Payload encoding (CBOR / JSON / Raw) for event sourcing";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      treefmt-nix,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      # Inline systems: nixpkgs 26.11 dropped x86_64-darwin, which github:nix-systems/default still lists.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          goPkg = pkgs.go_1_27;

          mkApp = name: runtimeInputs: text: {
            type = "app";
            program = "${pkgs.writeShellApplication { inherit name runtimeInputs text; }}/bin/${name}";
          };

          # Hermetic module source: fetches dependencies through the Nix sandbox
          # (goModules FOD) instead of relying on $HOME/GOMODCACHE, so checks and
          # `nix build` work without network access at build time.
          # `go` is an outer-param of buildGoModule in this nixpkgs — passing it
          # as a call argument is silently ignored (GOTOOLCHAIN=local then runs
          # nixpkgs' default go, e.g. 1.26.8 against a go.mod requiring 1.27.1).
          goModule = (pkgs.buildGoModule.override { go = goPkg; }) {
            pname = "go-codec";
            version = "unstable";
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.gitTracked ./.;
            };
            vendorHash = "sha256-auINCYX4nm63FITg+IWYBFfSsj7uut6KeWai1VLVR9Y=";
          };

          # erraudit runner: fetches the pinned version from source at run time.
          # erraudit is a PRIVATE module (proxy.golang.org 404s), so this needs
          # ambient GitHub credentials and network — it is therefore NOT a flake
          # input (that would force auth for every nix command) and NOT part of
          # any sandboxed check. GOEXPERIMENT=jsonv2 is required because
          # erraudit v0.4.x imports encoding/json/v2, which go 1.26 only builds
          # with the experiment enabled. Version must match the ci.yml pin.
          errauditApp = pkgs.writeShellApplication {
            name = "erraudit";
            runtimeInputs = [ goPkg ];
            text = ''
              export GOPRIVATE="github.com/larsartmann/*"
              export GOEXPERIMENT=jsonv2
              exec go run github.com/larsartmann/erraudit/cmd/erraudit@v0.4.0 "$@"
            '';
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            # The auto-added checks.treefmt is replaced by checks.format below,
            # which pins the go toolchain for goimports.
            flakeCheck = false;
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              nixfmt.enable = true;
            };
          };

          # treefmt's goimports shells out to `go` for module-aware import
          # resolution. The sandbox's ambient go can be older than go.mod's
          # directive; with the default GOTOOLCHAIN=auto it then tries (and,
          # offline, fails) to download the pinned toolchain. Pin the same go
          # the devShell uses and forbid toolchain downloads instead.
          checks.format = (config.treefmt.build.check self).overrideAttrs (old: {
            nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ goPkg ];
            GOTOOLCHAIN = "local";
          });
          devShells.default = pkgs.mkShellNoCC {
            packages = [
              goPkg
              errauditApp
              pkgs.golangci-lint
              pkgs.gopls
              pkgs.trash-cli
            ];

            GOWORK = "off";

            shellHook = ''
              echo "go-codec dev shell — $(go version)"
            '';
          };

          devShells.ci = pkgs.mkShellNoCC {
            packages = [
              goPkg
              pkgs.golangci-lint
            ];

            GOWORK = "off";
          };

          packages.default = goModule.overrideAttrs (_old: {
            doCheck = false;
          });

          checks = {
            build = goModule.overrideAttrs (_old: {
              doCheck = false;
            });

            # Runs `go test ./...` in both JSON modes inside the sandbox against
            # the vendored dependency set.
            test = goModule.overrideAttrs (_old: {
              checkPhase = ''
                runHook preCheck
                go test ./... -count=1
                GOEXPERIMENT=jsonv2 go test ./... -count=1
                runHook postCheck
              '';
            });
          };

          apps = {
            test = mkApp "test" [ goPkg ] ''
              echo "=== Testing json v1 ==="
              go test ./... -count=1 "$@"
              echo "=== Testing json v2 ==="
              GOEXPERIMENT=jsonv2 go test ./... -count=1 "$@"
            '';

            test-race = mkApp "test-race" [ goPkg ] ''
              echo "=== Race testing json v1 ==="
              go test ./... -race -count=1 "$@"
              echo "=== Race testing json v2 ==="
              GOEXPERIMENT=jsonv2 go test ./... -race -count=1 "$@"
            '';

            build = mkApp "build" [ goPkg ] ''
              echo "=== Building json v1 ==="
              go build ./...
              echo "=== Building json v2 ==="
              GOEXPERIMENT=jsonv2 go build ./...
            '';

            lint = mkApp "lint" [ pkgs.golangci-lint ] ''
              echo "=== Linting json v1 ==="
              golangci-lint run ./...
              echo "=== Linting json v2 ==="
              golangci-lint run --build-tags goexperiment.jsonv2 ./...
            '';

            coverage = mkApp "coverage" [ goPkg ] ''
              go test ./... -coverprofile=coverage.out -covermode=atomic "$@"
              go tool cover -func=coverage.out
            '';

            erraudit = {
              type = "app";
              program = "${errauditApp}/bin/erraudit";
            };

            # Parity with the CI test job: run all four drift tripwires in the
            # same order the v1 leg of .github/workflows/ci.yml runs them.
            tripwires = mkApp "tripwires" [ goPkg ] ''
              set -euo pipefail
              bash scripts/check-features-planned.sh
              bash scripts/check-go-version.sh
              bash scripts/check-error-codes.sh
              bash scripts/check-erraudit-version.sh
              echo "tripwires: all four checks passed"
            '';

            clean = mkApp "clean" [ goPkg pkgs.trash-cli ] ''
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };
    };
}
