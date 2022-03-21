{ lib, buildGoModule, fetchFromGitHub, stdenv, glibc }:

buildGoModule rec {
  pname = "inkpot-server";
  version = "0.0.1";

  src = ./.;
  vendorSha256 = "sha256-Dhfx0+NziS4NSzT4r4AWzlGj5vHYihJNJH6dH0xbyjw=";

  buildInputs = [
    stdenv
    glibc.static
  ];
  ldflags = "-linkmode external -extldflags -static";

  meta = with lib; {
    description = "A smol social network spilling onto your e-paper display";
    homepage = "https://github.com/saunaclub/inkpot-server/";
    license = licenses.agpl3Plus;
    maintainers = with maintainers; [];
  };
}
