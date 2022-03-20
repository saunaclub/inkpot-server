{ lib, buildGoModule, fetchFromGitHub, stdenv, glibc }:

buildGoModule rec {
  pname = "inkpot-cli";
  version = "0.0.1";

  src = ./.;
  vendorSha256 = "sha256-+NV7tsw76n4PIYjeSU9SqyzMdv7OoRJA1TANqhJcR9I=";

  buildInputs = [
    stdenv
    glibc.static
  ];
  ldflags = "-linkmode external -extldflags -static";

  meta = with lib; {
    description = "Command-line tool to generate 4-bit grayscale images";
    homepage = "https://github.com/saunaclub/inkpot-cli/";
    license = licenses.gpl3Plus;
    maintainers = with maintainers; [];
  };
}
