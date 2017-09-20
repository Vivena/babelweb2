function babelWebV2() {
    var babelDesc = {};
    var current = "unknown";

    var routers = {};
    var nodes = [];
    var links = [];
    var metrics = [];

    var addrToRouterId;

    var initEnd = false;
    var lostUpdate = false;

    var svg, color, width, height, simulation, vis, koef = 1;

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

    function connect(socketWarper) {
        socketWarper.socket = new WebSocket(`ws${location.protocol == 'https:' ? 's' : ''}://${location.host}/ws`)

	socketWarper.socket.onerror = console.error;

	socketWarper.socket.onopen = function(event) {
	    d3.selectAll("body").select("#state")
		.text("Connected").style("background-color", "green");
	}

	socketWarper.socket.onclose = function(event) {
	    d3.selectAll("body").select("#state")
		.text("Disconnected").style("background-color", "red");
	}

	socketWarper.socket.onmessage = convertJSON;
	redraw();
    }

    function convertJSON(event) {
	var data = JSON.parse(event.data);

        if(!'router' in data) {
            throw "No router in update";
        }

	if(current === "unknown")
	    current = data.router;

	if(!(data.router in babelDesc))
	    babelDesc[data.router] = {
		"self": {"name": data.name, "id": data.router},
		"interface": {},
		"neighbour": {},
		"route": {},
		"xroute": {},
	    };
	babelDesc[data.router].self.name = data.name;
	babelDesc[data.router][data.table][data.id] = {};
	if(data.action === "flush")
	    delete babelDesc[data.router][data.table][data.id];
	else {
	    for(var key in data.data) {
		babelDesc[data.router][data.table][data.id][key] =
		    data.data[key];
	    }
	}

	updateSwitch();
	if(data.router === current)
	    updateCurrent(current);
    }

    function isEmpty() {
	for(t in babelDesc[current]) {
	    if(t == "self")
		continue;
	    for(i in babelDesc[current][t])
		return metrics.length <= 1;
	}
	return true;
    }

    function updateSwitch() {
	var options = d3.select("#nodes").selectAll("option")
	    .data(d3.keys(babelDesc), function(d) { return d;});
	options.enter().append("option")
	    .attr("value", function(d) { return d; })
	    .text(function (d) { return babelDesc[d].self.name; });
	options.exit().remove();
	d3.selectAll("#nodes option").sort(function(x, y) {
	    if(x === "unknown")
		return -1;
	    else
		return babelDesc[x].self.name
		.localeCompare(babelDesc[y].self.name);
	});

	var sel = document.getElementById("nodes");
	for(var i, j = 0; i = sel.options[j]; j++) {
	    if(i.value == current) {
		sel.selectedIndex = j;
		break;
	    }
	}
    }

    function updateCurrent(newCurrent) {
	if(!initEnd) {
	    lostUpdate = true;
	    return;
	}

	if(current != newCurrent) {
	    oneToOne();
	    routers = {};
	}

	current = newCurrent;
	updateTable("neighbour");
	updateTable("route");
	updateTable("xroute");
	recomputeNetwork();
	redraw();

    }

    function zoomOut(factor) {
	koef /= factor;

	if(koef == 1)
	    d3.select("#oto").attr("disabled", true);
	else
	    d3.select("#oto").attr("disabled", null);

	redraw();
    }

    function zoomIn(factor) {
	zoomOut(1/factor);
    }

    function oneToOne() {
	zoomOut(koef);
    }

    function initLegend() {
       for(id in colors) {
         d3.selectAll(".legend-"+id)
           .append("svg:svg")
           .attr("width", 10)
           .attr("height", 10)
           .attr("class", "legend-dot")
           .append("svg:circle")
           .attr("cx", 5).attr("cy", 5).attr("r", 5)
           .attr("stroke-width", 0)
           .attr("fill",colors[id]);
       }
   }

    function initGraph() {
	width = 600;
	height = 400;

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
            .force("link", d3.forceLink())
            .force("charge", d3.forceManyBody().strength(-500))
	    .on("tick", ticked);

	initEnd = true;
	if(lostUpdate) {
	    updateCurrent(current);
	    lostUpdate = false;
	}
    }

    function recomputeNetwork() {
	if(typeof routers[current] == 'undefined') {
	    routers[current] = {
		id: current,
		metric: 0,
		fx: width/2,
		fy: height/2,
		fixed: true
	    };
	}

	function first(array, f) {
	    var i = 0, n = array.length, a = array[0], b;
	    while (++i < n) {
		if (f.call(array, a, b = array[i]) > 0) {
		    a = b;
		}
	    }
	    return a;
	}

	function insertKey(arr, obj) {
	    for(var i=0; i<arr.length; i++) {
		if(arr[i].key == obj.key) return arr;
	    }
	    arr.push(obj);
	    return arr;
	}

	for (var r in routers) {
	    routers[r].metric = undefined;
	}
	routers[current].metric = 0;

	var neighToRouterMetric = {};
	for(var route in babelDesc[current].route) {
	    var r = babelDesc[current].route[route];
	    var metric = r.metric;
	    var refmetric = r.refmetric;

	    if(!routers[r.id]) {
		routers[r.id] = {
		    id:r.id,
		    metric:metric,
		    via:r.via
		};
	    } else {
		if(routers[r.id].metric == undefined ||
		   metric < routers[r.id].metric) {
		    routers[r.id].metric = metric;
		    routers[r.id].via = r.via;
		}
	    }

	    if(!neighToRouterMetric[r.via])
		neighToRouterMetric[r.via] = {};
	    if(!neighToRouterMetric[r.via][r.id])
		neighToRouterMetric[r.via][r.id] = { refmetric: refmetric };
	    else
		neighToRouterMetric[r.via][r.id].refmetric =
		Math.min(neighToRouterMetric[r.via][r.id].refmetric, refmetric);
	}

	addrToRouterId = {};
	for(var n in neighToRouterMetric) {
	    addrToRouterId[n] =
		first(d3.entries(neighToRouterMetric[n]),
		      function(a, b) {
			  return a.value.refmetric -
			      b.value.refmetric;
		      }).key;
	}

	nodes = []; metrics = [];
	for (var r in routers) {
	    if(routers[r].metric == undefined)
		delete routers[r];
	    else {
		nodes.push(routers[r]);
		metrics.push({source:routers[current],
			      target:routers[r],
			      metric:routers[r].metric,
			     });
	    }
	}
	for (var n in neighToRouterMetric)
	    for(var id in neighToRouterMetric[n])
		metrics.push({source:routers[addrToRouterId[n]],
			      target:routers[id],
			      metric:neighToRouterMetric[n][id].refmetric});

	links = [];
	for(var r_key in babelDesc[current].route) {
	    var r = babelDesc[current].route[r_key];
	    if(r.metric == 65535)
		continue;

	    insertKey(links, {
		key: normalizeId(r.id + r.via + r.installed),
		path: [routers[current],
		       routers[addrToRouterId[r.via]],
		       routers[r.id]
		      ],
		installed: r.installed }
		     );
	}
    }

    function redraw() {
	function isNeighbour(id) {
	    for(var n in babelDesc[current].neighbour)
		if(addrToRouterId[babelDesc[current].neighbour[n].address] ==
		   id)
		    return true;
	    return false;
	}

	if(isEmpty()) {
	    d3.select("#oto").attr("disabled", true);
	    d3.select("#zin").attr("disabled", true);
	    d3.select("#zout").attr("disabled", true);
	}
	else {
	    // It's not an error that there is no #oto (I hope so)
	    d3.select("#zin").attr("disabled", null);
	    d3.select("#zout").attr("disabled", null);
	}

	simulation.force("link")
	    .links(metrics)
	    .strength(1)
	    .distance(function(d) {return d.metric * koef;});

	var node = vis.selectAll("circle.node")
	    .data(nodes);
	node.enter().append("svg:circle")
	    .attr("class", "node")
	    .attr("r", 5)
	    .attr("stroke", "white")
	    .attr("id", function(d) {return "node-"+normalizeId(d.id);})
	    .call(d3.drag()
		  .on("start", dragstarted)
		  .on("drag", dragged)
		  .on("end", dragended))
	    .append("svg:title");
	node.exit().remove();
	vis.selectAll("circle.node")
	    .style("fill", function(d) {
		return (d.id == current) ?
		    colors.current : (isNeighbour(d.id) ?
				      colors.neighbour : colors.other);
	    })
	    .each(function(d) {
		d3.select(this).select("title")
		    .text(
			nodeName(d.id)
			    + " ["+d.id+"]"
			    + " (metric: "+d.metric+")");

	    });

	function nodeName(id) {
	    var name =
		(babelDesc[id] && babelDesc[id].self.name) ||
		"unknown";
	    return name;
	}

	var route_path = d3.line()
	    .x(function(d) {
		if(typeof d == 'undefined') return null;
		else return d.x;
	    })
	    .y(function(d) {
		if(typeof d == 'undefined') return null;
		else return d.y;
	    })
	    .curve();

	var link = vis.selectAll("path.route")
	    .data(links);
	link.enter().insert("svg:path", "circle.node")
	    .attr("class", "route")
	    .attr("stroke", colors.route)
	    .attr("fill", "none")
	    .attr("id", function(d) { return "link-"+d.key; })
	    .attr("d", function(d) { return route_path(d.path); });
	link.exit().remove();

	simulation.alpha(1).restart();
	simulation.nodes(nodes);
    }

    function ticked() {

	vis.selectAll("circle.node")
	    .attr("cx", function(d) {return d.x; })
	    .attr("cy", function(d) {return d.y; });


	var route_path = d3.line()
	    .x(function(d) {
		if(typeof d == 'undefined') return null;
		else return d.x;
	    })
	    .y(function(d) {
		if(typeof d == 'undefined') return null;
		else return d.y;
	    })
	    .curve(d3.curveLinear);

	var show_all = d3.select("#show_all").property("checked");
	vis.selectAll("path.route")
	    .attr("display", function(d) {
		return (d.installed && d.metric != 65535) || show_all ?
		    "inline" : "none"; })
	    .attr("opacity", function(d) {return d.installed ? "1" : "0.3";})
	    .attr("stroke", colors.route)
	    .attr("stroke-dasharray", function(d) {
		return d.installed ? "none" : "5,2"; })
	    .attr("d", function(d) { return route_path(d.path); });
    }

    function dragstarted(d) {
	if(d.id == current)
	    return;
	if (!d3.event.active)
	    simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
    }

    function dragged(d) {
	if(d.id == current)
	    return;
	d.fx = d3.event.x;
	d.fy = d3.event.y;
    }

    function dragended(d) {
	if(d.id == current)
	    return;
	if (!d3.event.active)
	    simulation.alphaTarget(0);
	d.fx = null;
	d.fy = null;
    }

    function normalizeId(s) {
	var allowedChars = "0123456789abcdef";
	var res = "";
	for(var i = 0; i < s.length; i++) {
	    var c = s.charAt(i);
	    if (allowedChars.indexOf(c) != -1)
		res += c;
	}
	return res;
    }

    function updateRow(d, name, headers) {
	var tr = d3.select(this);

	var costColor = d3.scaleLog()
	    .domain([0, 96, 256, 65535])
	    .range([colors.wiredLink,
		    colors.wiredLink,
		    colors.losslessWireless,
		    colors.unreachableNeighbour])
	    .interpolate(d3.interpolateHcl);

	if(name == "route")
	    tr.style("background-color",
		     (d.value.metric == 65535 ? colors.unreachable :
		      d.value.installed ? colors.installed :
		      colors.uninstalled));
	else if(name == "neighbour")
	    tr.style("background-color", costColor(d.value.rxcost));

	var row = tr.selectAll("td")
	    .data(headers.map(function(h){
		if(h == "reach") {
		    s = d.value[h].toString(16);
		    for(; s.length < 4;)
			s = "0" + s;
		    return s;
		}
		return d.value[h];
	    }));
	row.text(function(d){return d;});
	row.enter().append("td").text(function(d){return d;});
    }

    function updateTable(name) {
	var table = d3.select("#"+name);
	table.select("tr.loading").remove();
	var headers = [];
	table.selectAll("th").each(function() {
	    headers.push(d3.select(this).text());
	});
	var rows = table.select("tbody").selectAll("tr")
	    .data(d3.entries(babelDesc[current][name]), function(d){
		if( typeof d == 'undefined' ) return null;
		else return d.key;
	    });
	rows.enter().append("tr")
	    .attr("id", function(d) {
		return name + "-" + normalizeId(d.key); });
	rows.exit().remove();
	table.select("tbody")
	    .selectAll("tr")
	    .each(function(d){updateRow.call(this, d, name, headers); });
    }

    babelWebV2.connect = connect;
    babelWebV2.initLegend =  initLegend;
    babelWebV2.initGraph = initGraph;
    babelWebV2.updateCurrent = updateCurrent;
    babelWebV2.redraw = redraw;
    babelWebV2.zoomIn = zoomIn;
    babelWebV2.zoomOut = zoomOut;
    babelWebV2.oneToOne = oneToOne;

    return babelWebV2;
}
