{
  description = "go-codec — Payload encoding (CBOR / JSON / Raw) for event sourcing";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
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
      systems,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

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
          goPkg = pkgs.go_1_26;

          mkApp = name: runtimeInputs: text: {
            type = "app";
            program = "${pkgs.writeShellApplication { inherit name runtimeInputs text; }}/bin/${name}";
          };

          # Hermetic module source: fetches dependencies through the Nix sandbox
          # (goModules FOD) instead of relying on $HOME/GOMODCACHE, so checks and
          # `nix build` work without network access at build time.
          goModule = pkgs.buildGoModule {
            pname = "go-codec";
            version = "unstable";
            src = lib.fileset.toSource {
              root = ./.;
              fileset = lib.fileset.gitTracked ./.;
            };
            vendorHash = "sha256-+JW5EWwTL68usFvVUq32KKyHb+YX0NR1IyDd6K/ShQc=";
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              nixfmt.enable = true;
            };
          };

          checks.format = config.treefmt.build.check self;
          devShells.default = pkgs.mkShellNoCC {
            packages = [
              goPkg
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

          packages.default = goModule.overrideAttrs (old: {
            doCheck = false;
          });

          checks = {
            build = goModule.overrideAttrs (old: {
              doCheck = false;
            });

            # Runs `go test ./...` in both JSON modes inside the sandbox against
            # the vendored dependency set.
            test = goModule.overrideAttrs (old: {
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

            clean = mkApp "clean" [ goPkg pkgs.trash-cli ] ''
              trash-put coverage.out 2>/dev/null || true
              go clean -testcache
            '';
          };
        };
    };
}
