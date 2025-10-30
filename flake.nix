{
  inputs = {
    nixpkgs.url = "github:numtide/nixpkgs-unfree?ref=nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs =
    inputs@{ flake-parts, self, ... }:
    flake-parts.lib.mkFlake { inherit inputs self; } {
      systems = [
        "aarch64-darwin"
        "x86_64-linux"
      ];

      perSystem =
        { pkgs, ... }:
        {
          devShells.default = pkgs.mkShell {
            buildInputs = with pkgs; [
              go
              gopls
              fyne
              (if pkgs.stdenvNoCC.hostPlatform.isDarwin then apple-sdk_14 else glfw)
            ];
          };

          packages =
            let
              attrs = rec {
                pname = "cave-assistant";
                version = "0.1.0";
                src = ./.;
                vendorHash = "sha256-7OhM7t0BRnmpEtvuEt/AuxklB20NT+ivvrTYe5KibEs=";
                ldflags = [
                  "-s" # omits symbol table
                  "-w" # omits DWARF debug info
                  "-H=windowsgui"
                ];
              };
            in
            {
              default = pkgs.buildGoModule attrs;
              windows = pkgs.pkgsCross.mingwW64.buildGoModule attrs;
            };
        };
    };
}
