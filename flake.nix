# TODO: check into to switching to devenv for making the dev shell. Use it inside this flake, not standalone. Project at code/sandbox/temp/devenv-flake.
{
  description = "Todo+ Application";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    # nixpkgs-unstable.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        templ = system: self.inputs.templ.packages.${system}.templ;
        # goVersion = 22; # Change this to update the whole stack
        # overlays =
        #   [ (final: prev: { go = prev."go_1_${toString goVersion}"; }) ];
        # pkgs = import nixpkgs { inherit overlays system; };
        # pkgs = nixpkgs.legacyPackages.${system};
        pkgs = import nixpkgs { inherit system; };
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            air
            esbuild
            go-task
            go_1_22
            gotools
            golangci-lint
            nodejs
            mongosh
            sass
            (templ system)
            tailwindcss
          ];

          shellHook = ''
            echo "Welcome to todo-plus!"
            echo "node `${pkgs.nodejs}/bin/node --version`"
            echo "npm `${pkgs.nodejs}/bin/npm --version`"
            echo "`${pkgs.go}/bin/go version`"
            echo "mongosh `${pkgs.mongosh}/bin/mongosh --version`"
            echo "templ `${(templ system)}/bin/templ --version`"
            echo "`${pkgs.air}/bin/air -v`"
            go get github.com/uptrace/bunrouter
          '';
        };
      });

}
