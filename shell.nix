{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [
    go
    rustc
    cargo
    delve
  ];

  hardeningDisable = [ "fortify" ];

  shellHook = ''
    if [ ! -f ./twgpu/bin/twgpu-map-photography ]; then
      cargo install --root=./twgpu twgpu-tools
      export PATH="$PWD/twgpu/bin:$PATH"
    fi

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
