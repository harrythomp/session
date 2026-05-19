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
                version = "2.0.2";

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
                    if [ ! -d ./test/zero-depth-worktree ]; then
                        mkdir test/zero-depth-worktree
                        cd test/zero-depth-worktree
                        mkdir project-one.git
                        cd project-one.git
                        git init --bare -b main
                        git worktree add ../project-one-main -b main
                        cd ../../../
                    fi
                    if [ ! -d ./test/extra-git-project ]; then
                        mkdir test/extra-git-project
                        cd test/extra-git-project
                        git init -b main
                        cd ../../
                    fi
                '';
            };
        };
    };
}
