# BabelWeb2
Yet Another Monitoring Tool for the Babel Routing Daemon

## Installation

    go get github.com/Vivena/babelweb2
    go install github.com/Vivena/babelweb2

## Get Started

Launch Babel on your local host:

    babeld -G 33123 ...

Start BabelWeb2:

    babelweb2

Or you can precise routers to monitor:

    babelweb2 -hp=[::1]:33123 -hp=[2001:660:3301:9208::88]:33123

By default, babelweb2 interface is located at: http://localhost:8080/
It's possible to change this behavior by editing `static/js/config.js` file.

From the routing daemon BabelWeb2 must obtain at least:

    - announced metric

    - computed metric

    - next hop

    - router-id

BabelWeb2 is created by Belynda Hamaz, Edward Guyot and Fedor Ryabinin.
It is highly inspired by Gabriel Kerneis's BabelWeb.