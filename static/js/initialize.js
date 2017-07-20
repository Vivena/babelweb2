/* ----       BabelWeb.2          ----*/

function BabelWebV2() {
    /* ----    Data base babel    ----*/
    var Routes = {};
    var Xroutes = {};
    var Neighbours = {};
    var Interfaces = {};

    /*----   Graph    ----*/
    var nodes = [];   //Nodes of the graph
    var links = [];
		var Idnodes =	{};
		var Idlinks =	{};

    /* ----    The structure of the data base    ----*/

    nodes.push(new Node("center"));
    //nodes.push(new Node("test"));
    //links.push(new Link("center", "test"));

    function NeighbourEntry(address, cost, iff, reach, rtt,
			    rttcost, rxcost, txcost) {
	this.address = address;
	this.cost = cost;
	this.iff = iff;
	this.reach = reach;
	this.rtt = rtt;
	this.rttcost = rttcost;
	this.rxcost = rxcost;
	this.txcost = txcost;
    }

    function RouteEntry(from, id, iff, installed, metric,
			prefix, refmetric, via) {
	this.from = from;
	this.id = id;
	this.iff = iff;
	this.installed = installed;
	this.metric = metric;
	this.prefix = prefix;
	this.refmetric = refmetric;
	this.via = via;
    }

    function XrouteEntry(from, metric, prefix) {
	this.from = from;
	this.metric = metric;
	this.prefix = prefix;
    }

    function InterfaceEntry(ipv4, ipv6, up) {
	this.ipv4 = ipv4;
	this.ipv6 = ipv6;
	this.up = up;
    }

    /*----   The structure of the graph    ----*/
    function Node(id ) {
	this.id = id;
    }

    function Link(source,target) {
        this.source = source;
        this.target = target;
    }

    function connect() {
	var socket = null;
	try {
            socket = new WebSocket("ws://localhost:8080/ws");
	} catch (exception) {
            console.error(exception);
	}

	socket.onerror = function(error) {
            console.error(error);
	};

	socket.onopen = function(event) {
            d3.selectAll("body").select("#state")
		.text("Connected").style("background-color", "green");

            this.onclose = function(event) {
		d3.selectAll("body").select("#state")
		    .text("Disconnected").style("background-color", "red");
            };

            this.onmessage = function(event) {
		ConvertJSON(event);
            };
	};
    }

    function ConvertJSON(event) {
	var data = JSON.parse(event.data);
	//console.log(data);

	switch (data.action) {
	case "add":
	    add(data);
            break;
	case "change":
	    change(data);
            break;
	case "flush":
	    flush(data);
            break;
	default:
	}
    }

    function add(data) {
	var entry = data.data;
	switch (data.table) {
	case "neighbour":
            Neighbours[data.id] =
		new NeighbourEntry(entry.address, entry.cost, entry.if,
				   entry.reach, entry.rtt, entry.rttcost,
				   entry.rxcost, entry.txcost);
					 addNode(entry.address);
					 addLink("center",entry.address);
	    		 restart();

	    insertNeighbour_html(entry.address, entry.cost, entry.if,
				 entry.reach, entry.rtt, entry.rttcost,
				 entry.rxcost, entry.txcost, data.id);
	    break;
	case "route":
            Routes[data.id] =
		new RouteEntry(entry.from.IP, entry.id, entry.if,
                               entry.installed, entry.metric, entry.prefix.IP,
			       entry.refmetric, entry.via);
            insertRoute_html(entry.prefix.IP, entry.metric, entry.refmetric,
			     entry.id, entry.via, entry.if,
			     entry.installed, data.id);

             if(entry.refmetric != 0){
							// addNode(entry.id);
          		 //addLink(entry.via,entry.id);
        		}
            break;
	case "xroute":
	    insertXroute_html(entry.metric, entry.prefix.IP, data.id);
            Xroutes[data.id] =
		new XrouteEntry(entry.from.IP, entry.metric, entry.prefix.IP);
            break;
	case "interface":
            Interfaces[data.id] =
		new InterfaceEntry(entry.ipv4, entry.ipv6, entry.up);
            break;
	default:
	}
    }

		function addNode(id) {
			nodes.push(new Node(id));
		 	Idnodes[id]= nodes.length-1;
		}

		function  addLink(id_source, id_target) {
			links.push(new Link(id_source, id_target));
			if(id_source in Idlinks){
				Idlinks[id_source].push(links.length-1);
			}
			else {
				Idlinks[id_source] = [links.length-1]
			}
			if(id_target in Idlinks){
				Idlinks[id_target].push(links.length-1);
			}
			else {
				Idlinks[id_target] = [links.length-1]
			}

		}

    function change(data){
	var entry = data.data;

	switch (data.table) {
	case "neighbour":
	    updateRowNeighbour(entry.address, entry.cost, entry.if,
			       entry.reach, entry.rtt, entry.rttcost,
			       entry.rxcost, entry.txcost,data.id);
            break;
	case "route":
	    updateRowRoute(entry.prefix.IP, entry.metric, entry.refmetric,
			   entry.id, entry.via, entry.if,
			   entry.installed, data.id);
            break;
	case "xroute":
	    updateRowXroute(entry.prefix.IP, entry.metric, data.id) ;
            break;
	case "interface":
            break;
	default:
	}
    }

    function flush(data){
			var entry = data.data;
			switch (data.table) {
				case "neighbour":		  deteteNeighbour(data.id);
					break;
				case "route":	    		deleteRow(data.id);
					break;
				case "xroute":	    	deleteRow(data.id);
					break;
				case "interface":	    deleteRow(data.id);
					break;
				default:
			}
		}
/*-----------------------------------------------------------*/
		function deteteNeighbour(id) {
			deleteRow(id);
			var index = Idnodes[id];
			nodes.splice(index,1);
			console.log(Idnodes);
			//Idnodes.delete(id);
			delete Idnodes[id];
			console.log(Idnodes);

			var indexlink = Idlinks[id];
			for(var x in   indexlink){
				links.splice(indexlink[x],1);
			}
			console.log(Idlinks);
			//Idlinks.delete(id);
			delete Idlinks[id];
			console.log(Idlinks);

			restart();
		}


    function deleteRow(row_id) {
	var row = document.getElementById(row_id);
	row.parentNode.removeChild(row);
    }

    /* Colors */
    var palette = {
	"gray" : "#777"
	, "lightGray" : "#ddd"
	, "blue" : "#03f"
	, "violet" : "#c0f"
	, "pink" : "#f69"
	, "green" : "#4d4"
	, "lightGreen" : "#8e8"
	, "yellow" : "#ff0"
	, "orange" : "#f18973"
	, "red" : "#f30"
    };

    var colors = {
	installed: palette.green
        , uninstalled: palette.lightGreen
        , unreachable: palette.lightGray
        , wiredLink: palette.yellow
        , losslessWireless: palette.orange
        , unreachableNeighbour: palette.red
        , current: palette.pink
        , neighbour: palette.violet
        , other: palette.blue
        , selected: palette.blue
        , route: palette.gray
    };

    function updateRowNeighbour(address, cost, iff, reach, rtt,
				rttcost, rxcost, txcost, id_row) {
	var row = document.getElementById(id_row);
	console.log(parseInt(cost));

	if(parseInt(cost) <= 96)
	    row.style.background = colors.wiredLink;
	else if(parseInt(cost) <= 256)
            row.style.background = colors.losslessWireless;
        else
            row.style.background = colors.unreachable;

	console.log(row);
	row.cells[0].innerHTML = address;
	row.cells[1].innerHTML = iff;
	row.cells[2].innerHTML = reach.toString(16);
	row.cells[3].innerHTML = rxcost;
	row.cells[4].innerHTML = txcost;
	row.cells[5].innerHTML = cost;
	row.cells[6].innerHTML = rtt;
    }

    function updateRowRoute(prefix, metric, refmetric, id, via,
			    iff, installed, id_row) {
	var row = document.getElementById(id_row);
	if(parseInt(metric) >= 65535)
	    row.style.background = colors.unreachable;
	else if(installed == true)
	    row.style.background = colors.installed;
        else if(installed == false)
            row.style.background = colors.uninstalled;

	console.log(row);
	row.cells[0].innerHTML = prefix;
	row.cells[1].innerHTML = metric;
	row.cells[2].innerHTML = refmetric;
	row.cells[3].innerHTML = id;
	row.cells[4].innerHTML = via;
	row.cells[5].innerHTML = iff;
	row.cells[6].innerHTML = installed;
    }

    function updateRowXroute(prefix, metric, id_row) {
	var row = document.getElementById(id_row);
	console.log(row);
	row.cells[0].innerHTML = prefix;
	row.cells[1].innerHTML = metric;
    }

    function insertNeighbour_html(address, cost, iff, reach, rtt, rttcost,
				  rxcost, txcost, id_row){
	if(document.getElementById("loading") != null)
	    deleteRow("loading");
	var arrayLignes = document.getElementById("neighbour");
	var row = arrayLignes.insertRow(-1);

	if(parseInt(cost) <= 96)
	    row.style.background = colors.wiredLink;
	else if(parseInt(cost) <= 256)
            row.style.background = colors.losslessWireless;
        else
            row.style.background = colors.unreachable;

	row.id = id_row;
	var colonne1 = row.insertCell(0);
	colonne1.innerHTML += address;
	var colonne2 = row.insertCell(1);
	colonne2.innerHTML += iff;
	var colonne3 = row.insertCell(2);
	colonne3.innerHTML += reach.toString(16);
	var colonne4 = row.insertCell(3);
	colonne4.innerHTML += rxcost;
	var colonne5 = row.insertCell(4);
	colonne5.innerHTML +=txcost ;
	var colonne6 = row.insertCell(5);
	colonne6.innerHTML +=cost;
	var colonne7 = row.insertCell(6);
	colonne7.innerHTML +=rtt;
    }

    function insertRoute_html(prefix, metric, refmetric, id, via,
			      iff, installed, id_row){
	if(document.getElementById("loading") != null)
	    deleteRow("loading");

	var arrayLignes = document.getElementById("route");
	var ligne = arrayLignes.insertRow(-1);

	if(parseInt(metric) >= 65535)
	    ligne.style.background = colors.unreachable;
	else if(installed == true)
	    ligne.style.background = colors.installed;
        else if(installed == false)
            ligne.style.background = colors.uninstalled;

	ligne.id = id_row;
	var colonne1 = ligne.insertCell(0);
	colonne1.innerHTML += prefix;
	var colonne2 = ligne.insertCell(1);
	colonne2.innerHTML += metric;
	var colonne3 = ligne.insertCell(2);
	colonne3.innerHTML += refmetric;
	var colonne4 = ligne.insertCell(3);
	colonne4.innerHTML += id;
	var colonne5 = ligne.insertCell(4);
	colonne5.innerHTML += via ;
	var colonne6 = ligne.insertCell(5);
	colonne6.innerHTML += iff;
	var colonne7 = ligne.insertCell(6);
	colonne7.innerHTML += installed;
    }

    function insertXroute_html(metric, prefix, id_row){
	if(document.getElementById("loading") != null)
	    deleteRow("loading");

	var arrayLignes = document.getElementById("xroute");
	var ligne = arrayLignes.insertRow(-1);
	ligne.id = id_row;
	var colonne1 = ligne.insertCell(0);
	colonne1.innerHTML += prefix;
	var colonne2 = ligne.insertCell(1);
	colonne2.innerHTML += metric;
    }

    var svg, color, width, height, link, node, simulation;

    function initGraph() {
	width = 600;
	height = 300;
	vis = d3.select("#fig")
	    .insert("svg:svg", ".legend")
	    .attr("width", width)
	    .attr("height", height)
	    .attr("stroke-width", "1.5px");

	svg = d3.select("svg");
        width = +svg.attr("width");
        height = +svg.attr("height");
	color = d3.scaleOrdinal(d3.schemeCategory20);
	simulation = d3.forceSimulation()
            .force("link", d3.forceLink().id(function(d) { return d.id; }))
            .force("charge", d3.forceManyBody())
            .force("center", d3.forceCenter(width / 2, height / 2))
	    .on("tick", ticked);

	node = svg.append("g")
            .attr("class", "nodes")
            .selectAll("circle");

	link = svg.append("g")
            .attr("class", "links")
            .selectAll("line");

	restart();
    }

    function restart() {
	node = node
            .data(nodes, function(d) { return d.id;})
            .enter().append("circle")
            .attr("r", 5)
            .attr("fill", function(d) { return color(d.group); })
            .call(d3.drag()
                  .on("start", dragstarted)
                  .on("drag", dragged)
                  .on("end", dragended))
	    .merge(node);

	link = link
            .data(links, function(d) {
		return d.source.id + "-" + d.target.id;});

	link.exit().transition()
	    .attr("stroke-opacity", 0)
	    .attrTween("x1", function(d) {
		return function() { return d.source.x; }; })
	    .attrTween("x2", function(d) {
		return function() { return d.target.x; }; })
	    .attrTween("y1", function(d) {
		return function() { return d.source.y; }; })
	    .attrTween("y2", function(d) {
		return function() { return d.target.y; }; })
	    .remove();

	link = link.enter().append("line").merge(link);

	simulation.force("link").links(links);
	simulation.alpha(1).restart();
	simulation.nodes(nodes);
    }

    function ticked() {
	link
	    .attr("x1", function(d) { return d.source.x; })
	    .attr("y1", function(d) { return d.source.y; })
	    .attr("x2", function(d) { return d.target.x; })
	    .attr("y2", function(d) { return d.target.y; });

	node
	    .attr("cx", function(d) { return d.x; })
	    .attr("cy", function(d) { return d.y; });
    }

    function dragstarted(d) {
	if (!d3.event.active)
	    simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
    }

    function dragged(d) {
	d.fx = d3.event.x;
	d.fy = d3.event.y;
    }

    function dragended(d) {
	if (!d3.event.active)
	    simulation.alphaTarget(0);
	d.fx = null;
	d.fy = null;
    }

    BBabelWebV2= {};
    BabelWebV2.connect = connect;
    //BabelWebV2.init = init;
    BabelWebV2.initGraph = initGraph;

    return BabelWebV2;
}
