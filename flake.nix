{
  description = "Todo+ Application";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";

    templ = {
      url = "github:a-h/templ/v0.3.819";
      inputs = {
        nixpkgs.follows = "nixpkgs";
      };
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

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
      templ,
      gitignore,
      gomod2nix,
    }:
    let
      # to work with older version of flakes
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = builtins.substring 0 8 lastModifiedDate;

      # System types to support.
      supportedSystems = [ "x86_64-linux" ];
      # [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      #   goVersion = 22; # Change this to update the whole stack
      #   overlays = [ (final: prev: { go = prev."go_1_${toString goVersion}"; }) ];

      forAllSystems =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            inherit system;
            pkgs = import nixpkgs { inherit system; };
            pkgs-unstable = import nixpkgs-unstable { inherit system; };
          }
        );

      templForSystem = system: templ.packages.${system}.templ;

      goPkgName = "go"; # 1.23
      nodePkgName = "nodejs_22";
    in
    {
      packages = forAllSystems (
        { system, pkgs, ... }:
        let
          buildGoApplication = gomod2nix.legacyPackages.${system}.buildGoApplication;
          templPkg = templForSystem (system);
        in
        {
          todo-server = buildGoApplication {
            inherit version;
            name = "todo-server";
            src = gitignore.lib.gitignoreSource ./.;
            go = pkgs.${goPkgName};
            # Must be added due to bug https://github.com/nix-community/gomod2nix/issues/120
            pwd = ./.;
            CGO_ENABLED = 1;
            # https://stackoverflow.com/a/58441379/359319
            # -trimpath
            #   remove all file system paths from the resulting executable.
            #   Instead of absolute file system paths, the recorded file names
            #   will begin either a module path@version (when using modules),
            #   or a plain import path (when using the standard library, or GOPATH).
            flags = [ "-trimpath" ];
            # go build -ldflags="-help" ./main.go <- will show all options
            ldflags = [
              "-s"
              "-w"
              "-extldflags -static"
            ];

            preBuild = ''
              echo "Generating code with templ ..."
              ${templPkg}/bin/templ generate
            '';

            buildPhase = ''
              runHook preBuild
              echo "Building go binary ..."
              go build -o ./bin/server main.go
              runHook postBuild
            '';

            postBuild = ''
              echo "Running go test ..."
              go test ./...
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

          # dockerImage = pkgs.dockerTools.buildImage {
          #   name = "catscii";
          #   tag = "latest";
          #   copyToRoot = [ todo-server todo-assets ];
          #   config = { Cmd = [ "${todo-server}/bin/todo-server" ]; };
          # };
        }
      );

      devShell = forAllSystems (
        {
          system,
          pkgs,
          pkgs-unstable,
          ...
        }:
        let
          goPkg = pkgs.${goPkgName};
          nodePkg = pkgs.${nodePkgName};
          templPkg = templForSystem (system);
        in
        pkgs.mkShell {
          buildInputs =
            with pkgs;
            [
              esbuild
              tailwindcss
            ]
            ++ [
              goPkg
              nodePkg
              templPkg
            ];

          packages =
            with pkgs;
            [
              air
              atlas
              go-task
              go-tools
              golangci-lint
              gomod2nix.legacyPackages.${system}.gomod2nix
              gotools
              google-cloud-sdk
              sqlite
            ]
            ++ [ pkgs-unstable.gopls ];

          env = {
            CLOUDSDK_ACTIVE_CONFIG_NAME = "todo-plus";
            # GCP_PROJECT_ID = "todo-plus-416720";
            # GCP_IMAGE_BUCKET = "custom-server-images";
            TEMPL_PATH = "${templPkg}/bin/templ";
            CGO_ENABLED = 1;
          };

          shellHook = ''
            echo "Welcome to todo-server!"
            echo "`${goPkg}/bin/go version`"
            echo "templ: `${templPkg}/bin/templ --version`"
            echo "node: `${nodePkg}/bin/node --version`"
            echo "npm: `${nodePkg}/bin/npm --version`"
          '';
        }
      );

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      # defaultPackage =
      #   forAllSystems (system: self.packages.${system}.todo-server);
    };
}
