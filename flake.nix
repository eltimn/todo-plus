{
  description = "Todo+ Application";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-23.11";
    # nixpkgs-unstable.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    templ.url = "github:a-h/templ";
  };

  outputs = { self, nixpkgs, flake-utils, templ }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        templ = system: self.inputs.templ.packages.${system}.templ;
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go nodejs mongosh air (templ system) ];

          shellHook = ''
            export DEBUG=1
          '';
        };
      });

}
