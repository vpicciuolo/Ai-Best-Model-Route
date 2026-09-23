{
  description = "Provider-agnostic Bifrost model router for Codex";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    systems.url = "github:nix-systems/default-linux";
    bifrost = {
      # Forked upstream to preserve OpenAI-compatible model display names and
      # reasoning-level metadata through Bifrost's model conversion pipeline.
      url = "github:applyinnovations/bifrost/bc0ff9e2bc56381c668301d35fa0cc378edff4d3";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      systems,
      bifrost,
    }:
    let
      eachSystem = nixpkgs.lib.genAttrs (import systems);
    in
    {
      packages = eachSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
          version = "0.1.0-dev";
          bifrostUI = pkgs.buildNpmPackage {
            pname = "bifrost-ui";
            inherit version;
            src = "${bifrost}/ui";
            npmDepsHash = "sha256-cOswnT4ZahWX66h9oiw4t3r5GZeOH/yjbnTCAsjVgnw=";
            npmBuildScript = "build-enterprise";
            installPhase = ''
              runHook preInstall
              cp -R out $out
              runHook postInstall
            '';
          };
          common = {
            pname = "ai-best-model-route";
            inherit version;
            src = self;
            vendorHash = "sha256-4louYB9FqGkC/FyN6LY/1TtcvGxaZMucvm19lNDRAZI=";
            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ pkgs.sqlite ];
            go = pkgs.go_1_27;
            doCheck = true;
          };
          configCheck = pkgs.buildGo127Module (
            common
            // {
              pname = "ai-best-route-config-check";
              subPackages = [ "cmd/config-check" ];
            }
          );
          router = pkgs.buildGo127Module (
            common
            // {
              pname = "ai-best-route-cli";
              subPackages = [ "cmd/router" ];
              ldflags = [ "-X main.version=${version}" ];
            }
          );
          gateway = pkgs.buildGo127Module (
            common
            // {
              pname = "ai-best-route-gateway";
              subPackages = [ "cmd/gateway" ];
            }
          );
          mockProvider = pkgs.buildGo127Module (
            common
            // {
              pname = "ai-best-route-mock-provider";
              subPackages = [ "cmd/mock-provider" ];
              doCheck = false;
            }
          );
          plugin = pkgs.buildGo127Module (
            common
            // {
              pname = "ai-best-model-route-plugin";
              vendorHash = "sha256-t4CxfKmrKw7pUXX6wEjT3cDTyjMze6UsLr0jrQKaSfY=";
              doCheck = false;
              postPatch = ''
                cp -R ${bifrost}/core bifrost-core
                chmod -R u+w bifrost-core
                go mod edit -replace github.com/maximhq/bifrost/core=./bifrost-core
              '';
              buildPhase = ''
                runHook preBuild
                CGO_ENABLED=1 go build -buildmode=plugin -trimpath \
                  -o ai-best-model-route.so ./plugins/codex-router
                runHook postBuild
              '';
              installPhase = ''
                runHook preInstall
                install -Dm755 ai-best-model-route.so \
                  $out/lib/bifrost/plugins/ai-best-model-route.so
                runHook postInstall
              '';
            }
          );
          bifrostHost = pkgs.buildGo127Module {
            pname = "bifrost-http";
            inherit version;
            src = bifrost;
            modRoot = "transports";
            vendorHash = "sha256-rcqc3JCHdXkLdev1qEYPJAPs+I2iBImlT7TY3x4f9EM=";
            subPackages = [ "bifrost-http" ];
            tags = [ "sqlite_static" ];
            postPatch = ''
              cp -R ${bifrostUI} transports/bifrost-http/ui
            '';
            nativeBuildInputs = [ pkgs.pkg-config ];
            buildInputs = [ pkgs.sqlite ];
            go = pkgs.go_1_27;
            ldflags = [
              "-s"
              "-w"
              "-X main.Version=v2.2.1-router"
            ];
            doCheck = false;
          };
          launcher = pkgs.writeShellApplication {
            name = "ai-best-model-route-server";
            runtimeInputs = [ pkgs.coreutils ];
            text = ''
              app_dir="''${BIFROST_APP_DIR:-/var/lib/bifrost}"
              config_file="''${BIFROST_CONFIG_FILE:-/etc/bifrost/config.json}"
              mkdir -p "$app_dir"
              install -m 0600 "$config_file" "$app_dir/config.json"
              exec ${bifrostHost}/bin/bifrost-http \
                -app-dir "$app_dir" \
                -host "''${BIFROST_HOST:-127.0.0.1}" \
                -port "''${BIFROST_PORT:-8080}"
            '';
          };
          licenseBundle = pkgs.runCommand "ai-best-model-route-licenses" { } ''
            install -Dm644 ${./LICENSE} $out/share/licenses/ai-best-model-route/LICENSE
            install -Dm644 ${./NOTICE} $out/share/licenses/ai-best-model-route/NOTICE
          '';
          imageRoot = pkgs.buildEnv {
            name = "ai-best-model-route-image-root";
            paths = [
              bifrostHost
              plugin
              gateway
              launcher
              licenseBundle
              pkgs.cacert
              pkgs.coreutils
              pkgs.curl
              pkgs.tzdata
            ];
            pathsToLink = [
              "/bin"
              "/etc"
              "/lib"
              "/share"
            ];
          };
          image = pkgs.dockerTools.streamLayeredImage {
            name = "ai-best-model-route";
            tag = "nix";
            contents = [ imageRoot ];
            config = {
              Entrypoint = [ "/bin/ai-best-model-route-server" ];
              Env = [
                "BIFROST_APP_DIR=/var/lib/bifrost"
                "BIFROST_CONFIG_FILE=/etc/bifrost/config.json"
                "BIFROST_HOST=127.0.0.1"
                "BIFROST_PORT=8080"
                "SSL_CERT_FILE=/etc/ssl/certs/ca-bundle.crt"
                "TZDIR=/share/zoneinfo"
              ];
              User = "65532:65532";
              WorkingDir = "/var/lib/bifrost";
              ExposedPorts = {
                "8080/tcp" = { };
              };
              Labels = {
                "org.opencontainers.image.source" = "https://github.com/vpicciuolo/ai-best-model-route";
                "org.opencontainers.image.version" = version;
                "org.opencontainers.image.title" = "AI Best Model Route";
                "org.opencontainers.image.licenses" = "Apache-2.0";
              };
            };
          };
          publishContainer = pkgs.writeShellApplication {
            name = "publish-container";
            runtimeInputs = [
              pkgs.coreutils
              pkgs.skopeo
            ];
            text = ''
              if (( $# < 1 )); then
                echo "usage: publish-container ghcr.io/vpicciuolo/ai-best-model-route:TAG [...]" >&2
                exit 2
              fi
              if [[ -z "''${GHCR_USER:-}" || -z "''${GHCR_TOKEN:-}" ]]; then
                echo "GHCR_USER and GHCR_TOKEN are required" >&2
                exit 1
              fi

              work_dir="$(mktemp -d)"
              trap 'rm -rf -- "$work_dir"' EXIT
              export REGISTRY_AUTH_FILE="$work_dir/auth.json"
              policy_file="$work_dir/policy.json"
              image_tar="$work_dir/image.tar"
              printf '%s\n' '{"default":[{"type":"insecureAcceptAnything"}]}' > "$policy_file"

              printf '%s' "$GHCR_TOKEN" | \
                skopeo login --authfile "$REGISTRY_AUTH_FILE" \
                  --username "$GHCR_USER" --password-stdin ghcr.io >/dev/null
              ${image} > "$image_tar"

              for destination in "$@"; do
                case "$destination" in
                  ghcr.io/vpicciuolo/ai-best-model-route:*) ;;
                  *)
                    echo "refusing unexpected destination: $destination" >&2
                    exit 2
                    ;;
                esac
                skopeo --policy "$policy_file" copy --all \
                  "docker-archive:$image_tar" "docker://$destination"
              done
            '';
          };
        in
        {
          inherit plugin image gateway;
          bifrost = bifrostHost;
          bifrost-ui = bifrostUI;
          ai-best-model-route-image = image;
          config-check = configCheck;
          inherit router;
          publish-container = publishContainer;
          mock-provider = mockProvider;
          default = pkgs.symlinkJoin {
            name = "ai-best-model-route-${version}";
            paths = [
              bifrostHost
              plugin
              configCheck
              router
              gateway
              licenseBundle
            ];
          };
        }
      );

      checks = eachSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          inherit (self.packages.${system}) plugin config-check;
          bifrost = self.packages.${system}.bifrost;
          bifrost-ui = self.packages.${system}.bifrost-ui;
          image = self.packages.${system}.ai-best-model-route-image;
          image-metadata =
            pkgs.runCommand "ai-best-route-image-metadata"
              {
                nativeBuildInputs = [
                  pkgs.gnutar
                  pkgs.jq
                ];
              }
              ''
                work=$TMPDIR/image
                mkdir -p "$work"
                ${self.packages.${system}.ai-best-model-route-image} > "$work/image.tar"
                tar -xf "$work/image.tar" -C "$work" manifest.json
                config_name=$(jq -r '.[0].Config' "$work/manifest.json")
                tar -xf "$work/image.tar" -C "$work" "$config_name"
                jq -e '.config.User == "65532:65532"' "$work/$config_name" >/dev/null
                jq -e '.config.Entrypoint == ["/bin/ai-best-model-route-server"]' "$work/$config_name" >/dev/null
                jq -e '.config.WorkingDir == "/var/lib/bifrost"' "$work/$config_name" >/dev/null
                jq -e '.config.Labels."org.opencontainers.image.licenses" == "Apache-2.0"' "$work/$config_name" >/dev/null
                touch $out
              '';
          ui-content =
            pkgs.runCommand "ai-best-route-ui-content"
              {
                nativeBuildInputs = [ pkgs.ripgrep ];
              }
              ''
                test -s ${self.packages.${system}.bifrost-ui}/index.html
                test -d ${self.packages.${system}.bifrost-ui}/assets
                ! rg -q "API-only build" ${self.packages.${system}.bifrost-ui}/index.html
                touch $out
              '';
          licenses =
            pkgs.runCommand "ai-best-route-licenses"
              {
                nativeBuildInputs = [ pkgs.diffutils ];
              }
              ''
                packaged=${self.packages.${system}.default}/share/licenses/ai-best-model-route
                diff -u ${./LICENSE} "$packaged/LICENSE"
                diff -u ${./NOTICE} "$packaged/NOTICE"
                touch $out
              '';
          format =
            pkgs.runCommand "ai-best-route-format"
              {
                nativeBuildInputs = [
                  pkgs.go_1_27
                  pkgs.jq
                  pkgs.nixfmt
                  pkgs.shellcheck
                  pkgs.shfmt
                ];
              }
              ''
                cp -R ${self} source
                chmod -R u+w source
                cd source
                test -z "$(gofmt -l cmd internal plugins)"
                nixfmt --check flake.nix nix/*.nix
                shfmt -d scripts
                shellcheck scripts/*.sh
                jq empty config/*.json
                touch $out
              '';
          lint =
            pkgs.runCommand "ai-best-route-lint"
              {
                nativeBuildInputs = [
                  pkgs.go_1_27
                  pkgs.pkg-config
                  pkgs.sqlite
                  pkgs.stdenv.cc
                ];
              }
              ''
                export GOCACHE=$TMPDIR/go-cache
                export GOPATH=$TMPDIR/go
                cp -R ${self} source
                chmod -R u+w source
                cp -R ${self.packages.${system}.router.goModules} source/vendor
                cd source
                go vet ./...
                touch $out
              '';
          schema =
            pkgs.runCommand "ai-best-route-schema"
              {
                nativeBuildInputs = [
                  pkgs.check-jsonschema
                  self.packages.${system}.config-check
                ];
              }
              ''
                check-jsonschema --schemafile ${./config/router.schema.json} ${./config/router.example.yaml}
                config-check ${./config/router.example.yaml} >/dev/null
                touch $out
              '';
          tekton =
            pkgs.runCommand "ai-best-route-tekton"
              {
                nativeBuildInputs = [
                  pkgs.ripgrep
                  pkgs.yq-go
                ];
              }
              ''
                pipeline=${./.tekton/ai-best-model-route-images.yaml}
                yq -e '.metadata.annotations."pipelinesascode.tekton.dev/target-namespace" == "tekton-buildkit"' "$pipeline" >/dev/null
                yq -e '.spec.taskRunSpecs[] | select(.pipelineTaskName == "publish-ghcr") | .podTemplate.hostUsers == false' "$pipeline" >/dev/null
                yq -e '.spec.pipelineSpec.tasks[] | select(.name == "publish-ghcr") | .runAfter | select(length == 1) | .[0] == "build-and-push-image"' "$pipeline" >/dev/null
                yq -e '.spec.pipelineSpec.tasks[] | select(.name == "publish-ghcr") | .taskRef.params[] | select(.name == "name" and .value == "nix-run-github-v2")' "$pipeline" >/dev/null
                yq -e '.spec.pipelineSpec.tasks[] | select(.name == "publish-ghcr") | .workspaces[] | select(.name == "github-packages-auth" and .workspace == "github-packages-auth")' "$pipeline" >/dev/null
                yq -e '.spec.workspaces[] | select(.name == "github-packages-auth") | .secret.secretName == "github-packages-credentials"' "$pipeline" >/dev/null
                publish_script="$(yq -r '.spec.pipelineSpec.tasks[] | select(.name == "publish-ghcr") | .params[] | select(.name == "SCRIPT") | .value' "$pipeline")"
                printf '%s\n' "$publish_script" | rg -F 'ghcr.io/vpicciuolo/ai-best-model-route:sha-$(tasks.clone.results.COMMIT_SHA)' >/dev/null
                printf '%s\n' "$publish_script" | rg -F 'ghcr.io/vpicciuolo/ai-best-model-route:main' >/dev/null
                touch $out
              '';
          setup-executor =
            pkgs.runCommand "ai-best-route-setup-executor"
              {
                nativeBuildInputs = [
                  pkgs.bash
                  pkgs.coreutils
                  pkgs.gawk
                  pkgs.gnugrep
                  pkgs.gnused
                  pkgs.jq
                ];
              }
              ''
                cp -R ${self} source
                chmod -R u+w source
                ${pkgs.bash}/bin/bash source/scripts/setup-local-test.sh
                touch $out
              '';
          race =
            pkgs.runCommand "ai-best-route-race"
              {
                nativeBuildInputs = [
                  pkgs.go_1_27
                  pkgs.pkg-config
                  pkgs.sqlite
                  pkgs.stdenv.cc
                ];
              }
              ''
                export GOCACHE=$TMPDIR/go-cache
                export GOPATH=$TMPDIR/go
                cp -R ${self} source
                chmod -R u+w source
                cp -R ${self.packages.${system}.router.goModules} source/vendor
                cd source
                go test -race ./...
                touch $out
              '';
          fuzz-smoke =
            pkgs.runCommand "ai-best-route-fuzz-smoke"
              {
                nativeBuildInputs = [
                  pkgs.go_1_27
                  pkgs.pkg-config
                  pkgs.sqlite
                  pkgs.stdenv.cc
                ];
              }
              ''
                export GOCACHE=$TMPDIR/go-cache
                export GOPATH=$TMPDIR/go
                cp -R ${self} source
                chmod -R u+w source
                cp -R ${self.packages.${system}.router.goModules} source/vendor
                cd source
                go test ./internal/catalog -run '^$' -fuzz FuzzHydrateDeterministic -fuzztime 2s
                touch $out
              '';
          go-test =
            pkgs.runCommand "ai-best-route-go-test"
              {
                nativeBuildInputs = [
                  pkgs.go_1_27
                  pkgs.pkg-config
                  pkgs.sqlite
                  pkgs.stdenv.cc
                ];
              }
              ''
                export GOCACHE=$TMPDIR/go-cache
                export GOPATH=$TMPDIR/go
                cp -R ${self} source
                chmod -R u+w source
                cp -R ${self.packages.${system}.router.goModules} source/vendor
                cd source
                go test ./...
                touch $out
              '';
          plugin-load =
            pkgs.runCommand "ai-best-route-plugin-load"
              {
                nativeBuildInputs = [ pkgs.curl ];
              }
              ''
                mkdir -p $TMPDIR/app
                cat >$TMPDIR/app/pricing.json <<JSON
                {}
                JSON
                cat >$TMPDIR/app/model-parameters.json <<JSON
                {}
                JSON
                cat >$TMPDIR/app/config.json <<JSON
                {
                  "\$schema": "https://www.getbifrost.ai/schema",
                  "version": 2,
                  "config_store": { "enabled": false },
                  "logs_store": { "enabled": false },
                  "framework": {
                    "pricing": {
                      "pricing_url": "file://$TMPDIR/app/pricing.json",
                      "model_parameters_url": "file://$TMPDIR/app/model-parameters.json"
                    }
                  },
                  "client": { "allow_direct_keys": true, "disable_content_logging": true },
                  "providers": { "openai": { "keys": [] } },
                  "plugins": [{
                    "enabled": true,
                    "name": "ai-best-model-route",
                    "path": "${self.packages.${system}.plugin}/lib/bifrost/plugins/ai-best-model-route.so",
                    "config": {
                      "version": 1,
                      "providers": {
                        "openai": { "credential_mode": "request_passthrough", "responses_mode": "native" }
                      },
                      "models": { "openai/test": { "codex": {} } }
                    }
                  }]
                }
                JSON
                ${self.packages.${system}.bifrost}/bin/bifrost-http \
                  -app-dir $TMPDIR/app -host 127.0.0.1 -port 18079 \
                  >$TMPDIR/bifrost.log 2>&1 &
                server_pid=$!
                trap 'kill $server_pid 2>/dev/null || true' EXIT
                ready=0
                for attempt in $(seq 1 60); do
                  if curl --fail --silent http://127.0.0.1:18079/health >/dev/null; then
                    ready=1
                    break
                  fi
                  if ! kill -0 $server_pid 2>/dev/null; then
                    cat $TMPDIR/bifrost.log >&2
                    exit 1
                  fi
                  sleep 0.25
                done
                if [ "$ready" -ne 1 ]; then
                  cat $TMPDIR/bifrost.log >&2
                  exit 1
                fi
                touch $out
              '';
          e2e =
            pkgs.runCommand "ai-best-route-e2e"
              {
                nativeBuildInputs = [
                  pkgs.bash
                  pkgs.curl
                  pkgs.jq
                ];
              }
              ''
                export BIFROST_BIN=${self.packages.${system}.bifrost}/bin/bifrost-http
                export ROUTER_PLUGIN=${self.packages.${system}.plugin}/lib/bifrost/plugins/ai-best-model-route.so
                export MOCK_PROVIDER_BIN=${self.packages.${system}.mock-provider}/bin/mock-provider
                ${pkgs.bash}/bin/bash ${./scripts/e2e.sh}
                touch $out
              '';
        }
      );

      apps = eachSystem (system: {
        config-check = {
          type = "app";
          program = "${self.packages.${system}.config-check}/bin/config-check";
        };
        default = {
          type = "app";
          program = "${self.packages.${system}.router}/bin/router";
        };
        router = {
          type = "app";
          program = "${self.packages.${system}.router}/bin/router";
        };
        gateway = {
          type = "app";
          program = "${self.packages.${system}.gateway}/bin/gateway";
        };
        bifrost = {
          type = "app";
          program = "${self.packages.${system}.bifrost}/bin/bifrost-http";
        };
        publish-container = {
          type = "app";
          program = "${self.packages.${system}.publish-container}/bin/publish-container";
        };
      });

      devShells = eachSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_27
              gopls
              gotools
              gotestsum
              golangci-lint
              govulncheck
              just
              jq
              yq-go
              shellcheck
              shfmt
              nixfmt
              pkg-config
              sqlite
            ];
          };
        }
      );

      formatter = eachSystem (
        system:
        let
          pkgs = import nixpkgs { inherit system; };
        in
        pkgs.writeShellApplication {
          name = "ai-best-route-format-nix";
          runtimeInputs = [ pkgs.nixfmt ];
          text = ''
            if (( $# > 0 )); then
              exec nixfmt "$@"
            fi
            mapfile -d "" files < <(find . -type f -name '*.nix' -not -path './.git/*' -print0)
            nixfmt "''${files[@]}"
          '';
        }
      );

      nixosModules.default = import ./nix/module.nix self;

      _bifrostSource = bifrost;
    };
}
