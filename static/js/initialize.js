/* ----       BabelWebV2          ----*/

function BabelWebV2() {
  /* ----    A propos de babel    ----*/
  var Routes = {};
  var Xroutes = {};
  var Neighbours ={};
  var Interfaces ={};

  /*----   graphe    ----*/
  var nodes = [];   //liste des noeuds du graphe
  var links = [];   //les liens

/* ----    A propos de babel    ----*/

  function NeighbourEntry(address, cost, iff, reach, rtt, rttcost, rxcost, txcost ) {
    this.address = address;
    this.cost = cost;
    this.iff = iff;
    this.reach = reach;
    this.rtt = rtt;
    this.rttcost = rttcost;
    this.rxcost = rxcost;
    this.txcost = txcost;
  }

  function RouteEntry(from, id, iff, installed, metric, prefix, refmetric, via){
    this.from = from; // pour le moment contient juste ip  sans le mask
    this.id = id;
    this.iff = iff;
    this.installed = installed;
    this.metric = metric;
    this.prefix = prefix;
    this.refmetric = refmetric;
    this.via = via;
  }

  function XrouteEntry(prefix, from,metric) {
    this.prefix = prefix;
    this.from = from;
    this.metric = metric;
  }

  function InterfaceEntry(up, ipv4, ipv6) {
    this.up = up;
    this.ipv4 = ipv4;
    this.ipv6 = ipv6;
  }

  /*----   graphe    ----*/
  function Node(id ) {
    this.id = id;
  }

  function Link(source,target) {
          this.source = source;
          this.target = target;
  }

  /*----   test graphe  ----*/
  nodes.push(new Node("center"));
  nodes.push(new Node("test"));
  links.push(new Link("center","test"));

  function init(){
    //TODO
  }

  function connect(){
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
        var elem = document.getElementById('state'); // À changer
        elem.innerHTML = "Connected";
        elem.style.backgroundColor = "green";

        this.onclose = function(event) {
          var elem = document.getElementById('state'); // À changer
          elem.innerHTML = "Disconnected.";
          elem.style.backgroundColor = "red";
        };

        this.onmessage = function(event) {
            ConvertJSON(event);

        };
        //this.send("Hello world!");
    };
  }

  function ConvertJSON(event) {
    var data = JSON.parse(event.data);
    console.log(data);

    switch (data.action) {
      case "add": add(data);
        break;
      case "change":add(data);
        break;
      case "flush": flush(data);
        break;
      default:
    }
  }

  function add(data){
    switch (data.table) {
      case "neighbour":
        var entry = data.data;
        Neighbours[data.id]= new NeighbourEntry(entry.address,
                                                     entry.cost,
                                                     entry.if,
                                                     entry.reach,
                                                     entry.rtt,
                                                     entry.rttcost,
                                                     entry.rxcost,
                                                     entry.txcost);
        nodes.push(new Node(entry.address));

        insertNeighbour_html(entry.address,
                            entry.cost,
                            entry.if,
                            entry.reach,
                            entry.rtt,
                            entry.rttcost,
                            entry.rxcost,
                            entry.txcost);

      console.log("test 1 : ");
      console.log(Neighbours[data.id]);
      break;

      case "route":
        var entry = data.data;
        if(entry.refmetric == 0)// est un voisin
        Routes[data.id]= new RouteEntry(entry.from.IP,
                                             entry.id,
                                             entry.if,
                                             entry.installed,
                                             entry.metric,
                                             entry.prefix.IP,
                                             entry.refmetric,
                                             entry.via);
        // if(nodes.includes(entry.via)== false)
        //   nodes.push(new Node(entry.from.via));
        //
        // if(entry.refmetric == 0)// est un voisin
        // {
        //   links.push(new Link("center",entry.from.via));
        // }
        console.log("test 2 : ");
        console.log(Routes[data.id]);
        break;

      case "xroute": //TODO
        break;

      case "interface": //TODO
        break;

      default:
    }
  }

  function insertNeighbour_html(address, cost, iff, reach, rtt, rttcost, rxcost, txcost){
    var arrayLignes = document.getElementById("neighbour");
    var ligne = arrayLignes.insertRow(-1);
    var colonne1 = ligne.insertCell(0);
    colonne1.innerHTML += address;
    var colonne2 = ligne.insertCell(1);
    colonne2.innerHTML += iff;
    var colonne3 = ligne.insertCell(2);
    colonne3.innerHTML +=reach;
    var colonne4 = ligne.insertCell(3);
    colonne4.innerHTML += rxcost;
    var colonne5 = ligne.insertCell(4);
    colonne5.innerHTML +=txcost ;
    var colonne6 = ligne.insertCell(5);
    colonne6.innerHTML +=cost;
    var colonne7 = ligne.insertCell(6);
    colonne7.innerHTML +=rtt;
  }

  function change(message){//TODO
  }
  function flush(message){//TODO
  }
  function initGraph() {
    /* Setup svg graph */
    width = 600;
    height = 400; /* display size */
    vis = d3.select("#fig")
      .insert("svg:svg", ".legend")
      .attr("width", width)
      .attr("height", height)
      .attr("stroke-width", "1.5px");
    // force = d3.layout.force(); /* force to coerce nodes */
    // force.charge(-1000); /* stronger repulsion enhances graph */
    // force.on("tick", onTick);
  }


  function initGraph2() {
    var svg = d3.select("svg"),
            width = +svg.attr("width"),
            height = +svg.attr("height");

    var color = d3.scaleOrdinal(d3.schemeCategory20);

    var simulation = d3.forceSimulation()
            .force("link", d3.forceLink().id(function(d) { return d.id; }))
            .force("charge", d3.forceManyBody())
            .force("center", d3.forceCenter(width / 2, height / 2));


    var link = svg.append("g")
            .attr("class", "links")
            .selectAll("line")
            .data(links)
            .enter().append("line")
            //.attr("stroke-width", function(d) { return Math.sqrt(d.value); });

    var node = svg.append("g")
            .attr("class", "nodes")
            .selectAll("circle")
            .data(nodes)
            .enter().append("circle")
            .attr("r", 5)
            .attr("fill", function(d) { return color(d.group); })
            .call(d3.drag()
                  .on("start", dragstarted)
                  .on("drag", dragged)
                  .on("end", dragended));

    node.append("title")
              .text(function(d) { return d.id; });

    simulation
        .nodes(nodes)
        .on("tick", ticked);

    simulation.force("link")
        .links(links);

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
     if (!d3.event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
   }

   function dragged(d) {
      d.fx = d3.event.x;
      d.fy = d3.event.y;
   }

   function dragended(d) {
      if (!d3.event.active) simulation.alphaTarget(0);
       d.fx = null;
       d.fy = null;
  }
 }

  BBabelWebV2= {};
  BabelWebV2.connect = connect;
  BabelWebV2.init = init;
  BabelWebV2.initGraph = initGraph;

  return BabelWebV2;
}


// /*----   test graphe  ----*/
//  nodes.push(new Node("Myriel",1));
//  nodes.push(new Node("Napoleon",1));
//  nodes.push(new Node("Mlle.Baptistine",1));
//  nodes.push(new Node("Mme.Magloire", 1));
//
//  links.push(new Link("Napoleon","Myriel"));
//  links.push(new Link("Mlle.Baptistine","Myriel"));
//  links.push(new Link("Mme.Magloire","Myriel"));
//  links.push(new Link("Mme.Magloire", "Mlle.Baptistine"));


// var nodes = [
//           {"id": "Myriel", "group": 1},
//           {"id": "Napoleon", "group": 1},
//           // {"id": "Mlle.Baptistine", "group": 1},
//           // {"id": "Mme.Magloire", "group": 1},
//           // {"id": "CountessdeLo", "group": 1},
//           // {"id": "Geborand", "group": 1},
//           // {"id": "Champtercier", "group": 1},
//           // {"id": "Cravatte", "group": 1},
//           // {"id": "Count", "group": 1},
//           // {"id": "OldMan", "group": 1},
//           // {"id": "Labarre", "group": 2},
//           // {"id": "Valjean", "group": 2},
//           // {"id": "Marguerite", "group": 3},
//           // {"id": "Mme.deR", "group": 2},
//           // {"id": "Isabeau", "group": 2},
//           // {"id": "Gervais", "group": 2},
//       ]
//
//       var links = [
//           {"source": "Napoleon", "target": "Myriel", "value": 1},
//         // {"source": "Mlle.Baptistine", "target": "Myriel", "value": 8},
//         // {"source": "Mme.Magloire", "target": "Myriel", "value": 10},
//         // {"source": "Mme.Magloire", "target": "Mlle.Baptistine", "value": 6},
//         // {"source": "CountessdeLo", "target": "Myriel", "value": 1},
//         // {"source": "Geborand", "target": "Myriel", "value": 1},
//         // {"source": "Champtercier", "target": "Myriel", "value": 1},
//         // {"source": "Cravatte", "target": "Myriel", "value": 1},
//         // {"source": "Valjean", "target": "Labarre", "value": 1},
//         // {"source": "Valjean", "target": "Mme.Magloire", "value": 3},
//         // {"source": "Valjean", "target": "Mlle.Baptistine", "value": 3},
//       ]
