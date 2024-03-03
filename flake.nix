{
  description = "Todo+ Application";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";

    templ = {
      url = "github:a-h/templ/v0.2.590";
      inputs = { nixpkgs.follows = "nixpkgs"; };
    };

    gitignore = {
      url = "github:hercules-ci/gitignore.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, templ, gitignore }@inputs:
    let
      templ = system: self.inputs.templ.packages.${system}.templ;

      # to work with older version of flakes
      lastModifiedDate =
        self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = builtins.substring 0 8 lastModifiedDate;

      # System types to support.
      supportedSystems = [ "x86_64-linux" ];
      # [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      #   goVersion = 22; # Change this to update the whole stack
      #   overlays = [ (final: prev: { go = prev."go_1_${toString goVersion}"; }) ];
      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

    in {
      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          web = pkgs.buildGo122Module {
            pname = "web";
            inherit version;
            # version = "0.1.0";
            src = gitignore.lib.gitignoreSource ./.;

            # This hash locks the dependencies of this package. It is
            # necessary because of how Go requires network access to resolve
            # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
            # details. Normally one can build with a fake sha256 and rely on native Go
            # mechanisms to tell you what the hash should be or determine what
            # it should be "out-of-band" with other tooling (eg. gomod2nix).
            # To begin with it is recommended to set this, but one must
            # remember to bump this hash when your dependencies change.
            # vendorHash = pkgs.lib.fakeHash;
            vendorHash = "sha256-6Bl7mtoU3GIdfgmTh8JEbOGeyLr/Cz/DVTJjC5824Ic=";

            #
            # go build -o ./bin/server main.go

            # configurePhase = ''
            #   runHook preConfigure

            #   ${templ system}/bin/templ generate

            #   runHook postConfigure
            # '';

            buildPhase = ''
              runHook preBuild

              echo "Building todo-plus..."
              # can't run npm here because nix does not have internet access
              # ${pkgs.nodejs}/bin/npm ci
              # ${pkgs.tailwindcss}/bin/tailwindcss -i ./web/assets/css/main.css -o dist/assets/css/main.css

              # module lookup disabled by GOPROXY=off
              # go get github.com/a-h/templ

              # cannot find module providing package github.com/a-h/templ: import lookup disabled by -mod=vendor
              echo "Generating code with templ ..."
              ${templ system}/bin/templ generate
              echo "Building go binary ..."
              go build -o ./bin/server -mod=mod main.go

              runHook postBuild
            '';

            installPhase = ''
              runHook preInstall

              # mkdir -p $out/bin
              # cp -r node_modules $out/node_modules
              # cp package.json $out/package.json
              # cp -r dist $out/dist

              runHook postInstall
            '';
          };
        });

      # Add dependencies that are only needed for development
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              gotools
              go-tools
              (templ system)
              go-task
              nodejs
              tailwindcss
            ];

            packages = with pkgs; [
              air
              atlas
              esbuild
              golangci-lint
              mongosh
              sass
            ];

            shellHook = ''
              echo "Welcome to todo-plus!"
              echo "`${pkgs.go}/bin/go version`"
              echo "templ: `${(templ system)}/bin/templ --version`"
              echo "node: `${pkgs.nodejs}/bin/node --version`"
              echo "npm: `${pkgs.nodejs}/bin/npm --version`"
            '';
          };
        });

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      defaultPackage = forAllSystems (system: self.packages.${system}.web);

      # apps = forAllSystems (system: {
      #   default = {
      #     type = "app";
      #     program = "${self.packages.${system}.default}/bin/server";
      #   };
      # });
    };
}

