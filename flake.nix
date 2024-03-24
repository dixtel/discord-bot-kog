{
  description = "c9 tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
        # twgpu-tools = pkgs.rustPlatform.buildRustPackage rec {
        #   pname = "twgpu-tools";
        #   version = "0.2.0";

        #   src = pkgs.fetchFromGitLab {
        #     owner = "Patiga";
        #     repo = "twgpu";
        #     rev = version;
        #     hash = "sha256-+s5RBC3XSgb8omTbUNLywZnP6jSxZBKSS1BmXOjRF8M=";
        #   };
        #   # sourceRoot = "${src}/${pname}";
        #   cargoHash = "sha256-BwRhHJcmf+ldneeAMoQN+Fv2f8BPQCrZ9kTIkkSeKaI=";
        # };
      in
      {
        # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-run#flake-output-attributes
        # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-run#apps
        apps = {
          # default = {
          #   type = "app";
          #   program = "${self.packages.${system}.default}/bin/site";
          # };
        };

        packages = {
          default = pkgs.rustPlatform.buildRustPackage rec {
            pname = "site";
            version = "0.1";
            src = ./.;
            cargoHash = "sha256-+Hlcu7YzTWJeKFbZlYgEYXKK+ZbNhZcubdGMCfRs9YM=ls";
          };
        };

        devShells = {
          default = pkgs.mkShell {
            nativeBuildInputs = with pkgs; [
              go
              rustc
              cargo
              delve
            ];

            hardeningDisable = [ "fortify" ];

            shellHook = ''
              cargo install twgpu-tools

              LD_LIBRARY_PATH="''${LD_LIBRARY_PATH:+$LD_LIBRARY_PATH:}${
                with pkgs;
                  lib.makeLibraryPath [
                    vulkan-loader
                    xorg.libX11
                    xorg.libXcursor
                    xorg.libXi
                    xorg.libXrandr
                  ]
              }"
              export LD_LIBRARY_PATH
            '';
          };
        };
      });
}
