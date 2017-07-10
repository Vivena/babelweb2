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
  function NeighbourEntry(address, iff, reach, rxcost, txcost, cost, rtt, rttcost) {
    this.address = address;
    this.iff = iff;
    this.reach = reach;
    this.rxcost = rxcost;
    this.txcost = txcost;
    this.cost = cost;
    this.rtt = rtt;
    this.rttcost = rttcost;
  }

  function RouteEntry(prefix, from, installed, id, metric, refmetric, via, iff){
    this.prefix = prefix;
    this.from = from;
    this.installed = installed;
    this.id = id;
    this.metric = metric;
    this.refmetric = refmetric;
    this.via = via;
    this.iff = iff;
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
  function Node(id ,group) {
    this.id = id;
    this.group = group;
  }

  function Link(source,target,value) {
          this.source = source;
          this.target = target;
          this.value = value;
  }

  /*----   test graphe  ----*/
  nodes.push(new Node("Myriel",1));
  nodes.push(new Node("Napoleon",1));
  nodes.push(new Node("Mlle.Baptistine",1));
  nodes.push(new Node("Mme.Magloire", 1));

  links.push(new Link("Napoleon","Myriel",1));
  links.push(new Link("Mlle.Baptistine","Myriel",8));
  links.push(new Link("Mme.Magloire","Myriel",10));
  links.push(new Link("Mme.Magloire", "Mlle.Baptistine", 6));


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
      console.log("Connected.");
    var t = d3.select("body")
              .append("p")
              .text("Connected");
    };

    this.onclose = function(event) {
      console.log("Disconnected.");
    };

    this.onmessage = function(event) {
      console.log(event)
    //  ConvertJSON(event)
    };

    // this.send("Hello world!");
  }

  function init(){

  }

  function ConvertJSON(message) {
    var data = JSON.parse(message.data);

    switch (data.action) {
      case "add": add(data.action);
        break;
      case "change":change(data.action);
        break;
      case "flush": flush(data.action);
        break;
      default:
    }
  }

  function add(message){
//{action : add , tableId : route , entryId : 12354 ,
// entry : {prefix : bla , reach : bla , ...}  }

    switch (message.tableId) {
      case "Route":
        Routes.push(new RouteEntry(message.tableId.entryId.));
        break;
      case "Xroute": Xroutes.push(new XrouteEntry(message.tableId.entryId));
        break;
      case "Interface":Interfaces.push(new InterfaceEntry(message.tableId.entryId));
        break;
      case "Neighbour":
        break;
      default:
    }
  }

  function change(message){

  }

  function flush(message){

  }

  function initGraph() {
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
            .attr("stroke-width", function(d) { return Math.sqrt(d.value); });

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
