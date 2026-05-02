{
    description = "session - A TUI for managing sessions";

    inputs = {
        nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
        flake-parts.url = "github:hercules-ci/flake-parts";
    };

    outputs = { flake-parts, ... }@inputs: flake-parts.lib.mkFlake { inherit inputs; } {
        systems = [ "x86_64-linux" "aarch64-linux" ];

        perSystem = { pkgs, ... }: {
            packages.default = pkgs.buildGoModule {
                pname = "session";
                version = "1.0.1";

                src = ./.;

                vendorHash = "sha256-DopXgsK+bO9T24yvbpxi5slVxx+0pgKOQf0datvFg+0=";

                subPackages = [ "." ];

                ldflags = [ "-s" "-w" ];
            };

            devShells.default = pkgs.mkShell {
                buildInputs = with pkgs; [
                    go
                    gopls
                ];

                shellHook = ''
                    export GOPATH=$HOME/go
                    export PATH=$GOPATH/bin:$PATH
                    if [ ! -d ./test/zero-depth-worktree/project-one ]; then
                        cd test/zero-depth-worktree
                        mkdir project-one
                        cd project-one
                        git init --bare -b main
                        git worktree add main
                        cd ../../../
                    fi
                '';
            };
        };
    };
}
