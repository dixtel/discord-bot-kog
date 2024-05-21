{ pkgs ? import (fetchTarball "https://github.com/NixOS/nixpkgs/archive/refs/tags/23.11.zip") { } }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    # backend
    go
    gopls
    rustc
    cargo
    delve
    pkg-config
    openssl
    # others
    direnv
    vscode.fhs
    go-task
    _1password
  ];

  hardeningDisable = [ "fortify" ];

  shellHook = ''
    if [ ! -f ./twgpu/bin/twgpu-map-photography ]; then
      ${pkgs.cargo}/bin/cargo install --root=./twgpu twgpu-tools
    fi

    export PATH="$PWD/twgpu/bin:$PATH"

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
}
