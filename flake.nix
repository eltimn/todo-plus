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

    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, templ, gitignore, gomod2nix }@inputs:
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
        let
          pkgs = nixpkgsFor.${system};
          buildGoApplication =
            gomod2nix.legacyPackages.${system}.buildGoApplication;
        in {
          todo-server = buildGoApplication {
            inherit version;
            name = "todo-server";
            src = gitignore.lib.gitignoreSource ./.;
            go = pkgs.go_1_22;
            # Must be added due to bug https://github.com/nix-community/gomod2nix/issues/120
            pwd = ./.;
            CGO_ENABLED = 0;
            # https://stackoverflow.com/a/58441379/359319
            # -trimpath
            #   remove all file system paths from the resulting executable.
            #   Instead of absolute file system paths, the recorded file names
            #   will begin either a module path@version (when using modules),
            #   or a plain import path (when using the standard library, or GOPATH).
            flags = [ "-trimpath" ];
            # go build -ldflags="-help" ./main.go <- will show all options
            ldflags = [ "-s" "-w" "-extldflags -static" ];

            preBuild = ''
              echo "Generating code with templ ..."
              ${templ system}/bin/templ generate
            '';

            buildPhase = ''
              runHook preBuild
              echo "Building go binary ..."
              go build -o ./bin/server main.go
              runHook postBuild
            '';

            installPhase = ''
              runHook preInstall
              mkdir -p $out/bin
              cp ./bin/server $out/bin/todo-server
              runHook postInstall
            '';
          };

          todo-assets = pkgs.buildNpmPackage {
            name = "todo-assets";
            src = gitignore.lib.gitignoreSource ./.;
            npmDepsHash = "sha256-QNpo1s9zL8V3Wab2fk9Ef2iAYkb/Vd1DmdOeyc38XU8=";
            dontNpmBuild = true;

            buildPhase = ''
              runHook preBuild
              echo "Building todo-assets ..."
              ${pkgs.tailwindcss}/bin/tailwindcss -i ./web/assets/css/main.css -o dist/assets/css/main.css --minify
              ${pkgs.esbuild}/bin/esbuild web/assets/js/main.js --outdir=dist/assets/js --bundle --target='esnext' --format=esm --minify
              runHook postBuild
            '';

            installPhase = ''
              cp web/assets/js/htmx*.min.js dist/assets/js/
              mkdir -p $out/assets
              cp -r dist/assets $out
            '';
          };
        });

      # Add dependencies that are only needed for development
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              esbuild
              go_1_22
              nodejs
              tailwindcss
              (templ system)
            ];

            packages = with pkgs; [
              air
              atlas
              go-task
              go-tools
              golangci-lint
              gomod2nix.legacyPackages.${system}.gomod2nix
              gopls
              gotools
            ];

            shellHook = ''
              echo "Welcome to todo-server!"
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
      # defaultPackage =
      #   forAllSystems (system: self.packages.${system}.todo-server);
    };
}

