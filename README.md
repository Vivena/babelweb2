# BabelWeb2
Yet Another Monitoring Tool for the Babel Routing Daemon.

## Installation

    go get github.com/Vivena/babelweb2
    go install github.com/Vivena/babelweb2

## Get Started

Launch Babel on your local host:

    babeld -g 33123 ...

By default, Babwelweb attempts to connect to a Babel node at [::1]:33123:

    babelweb2

You may specify a list of Babel nodes to monitor:

    babelweb2 -node=[::1]:33123 -node=[2001:660:3301:9208::88]:33123

The web interface is on port 8080 (type "http://localhost:8080" in your
browser). You may specify a different port with the "-http" flag.

From the routing daemon BabelWeb2 must obtain at least:
- Announced metric
- Computed metric
- Next hop
- Router-id

BabelWeb2 was written by Belynda Hamaz, Edward Guyot and Fedor Ryabinin
based on Gabriel Kerneis's [BabelWeb](https://github.com/kerneis/babelweb).
