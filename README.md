# BabelWeb2
Yet Another Monitoring Tool for the [Babel routing daemon][babeld].

## Installation

    go get github.com/Vivena/babelweb2
    go install github.com/Vivena/babelweb2

## Get Started

Launch Babel on your local host:

    babeld -g 33123 ...

By default, BabelWeb2 attempts to connect to a Babel node at `[::1]:33123`:

    babelweb2

You may specify a list of Babel nodes to monitor:

    babelweb2 -node=[::1]:33123 -node=[2001:660:3301:9208::88]:33123

The web interface is on port 8080 (type "http://localhost:8080" in your
browser). You may specify a different port with the `-http` flag.

From the routing daemon BabelWeb2 must obtain at least:
- Announced metric
- Computed metric
- Next hop
- Router-id

BabelWeb2 was written by Belynda Hamaz, Edward Guyot and Fedor Ryabinin
based on Gabriel Kerneis’s [BabelWeb][babelweb].

Special thanks to Antonin, Zeinab, Gwendoline and Boris for sharing the
office. Not less special thanks to Athénaïs for being somewhere around.
Authors wish to express their most sincere gratitude to [Juliusz
Chroboczek][jch], neither uncle nor cousin but most decidedly teacher and
friend.

[babeld]: https://github.com/jech/babeld
[babelweb]: https://github.com/kerneis/babelweb
[jch]: https://www.irif.fr/~jch/
