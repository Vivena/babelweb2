var Terminal = (function(){

  var constructor = function(id,socketWarper){
    var waitingServData = false
    var that = this
    var focus = false
    var socketWarper = socketWarper
    var fireCursorInterval = function() {
      setTimeout(function () {
        if (focus) {
          that.cursor.style.visibility = that.cursor.style.visibility === 'visible' ? 'hidden' : 'visible'
          fireCursorInterval()
        }
        else{
          that.cursor.style.visibility = 'hidden'
        }

      }, 500)
    }

    this.div = document.createElement('div');

    this.terminal = document.createElement('div')
    this.inside = document.createElement('div')
    this.output = document.createElement('p')
    this.inputLine = document.createElement('p')
    this.inputSpan = document.createElement('span')
    this.inputShow = document.createElement('p')
    this.inputElem = document.createElement('input')
    this.cursor = document.createElement('span')


    this.connect = function(ip){
      setTimeout(function(){
        console.log("connection")
        socketWarper.socket.send("connect "+ip)
      }, 1000)
      this.inputSpan.textContent = "notConnected > "
    }

    this.inputElem.onkeydown = function(e){
      if (e.which === 37 || e.which === 39 || e.which === 38 || e.which === 40 || e.which === 9) {
        e.preventDefault()
      }
      else if ( e.which !== 13) {
				setTimeout(function(){
          that.inputShow.textContent = that.inputElem.value
        }, 1)
}
    }

    this.inputElem.onkeyup = function(e){
      inval = that.inputElem.value
    if (e.which === 13) {
        if (waitingServData && inval !== "stop") {
          that.print("waiting for server's responce")
          that.inputElem.value = ''
          that.inputShow.textContent = ''
        }
        else{
          div = document.getElementById(id)
          var inputResult = that.inputSpan.textContent + inval
          that.print(inputResult)
          waitingServData = false
          that.inputElem.value = ''
          that.inputShow.textContent = ''

          console.log(socketWarper.socket)
          socketWarper.socket.send(inval)
        }
      }
    }

    this.getServData = function(data,done){
      this.print(data)
      if(done){
        waitingServData = false
      }
    }

    this.print = function(message){
      var toAdd = document.createTextNode(message)

      this.output.appendChild(document.createElement("br"));
      this.output.appendChild(toAdd)
      this.div.scrollTop = this.div.scrollHeight;

    }

/******************************************************************************/

  this.inputElem.onblur = function(){
    focus = false
  }

    this.inputElem.onfocus = function(){
          focus = true
          fireCursorInterval()
    }

    this.div.onclick = function(){
      that.inputElem.focus()
    }

    this.setTextColor = function (col) {
			 this.div.style.color = col
       this.inputElem.style.color = col
     }

		this.setBackgroundColor = function (col) {
			this.div.style.background = col
      this.inputElem.style.backgroundColor = col
    }

    this.setHeight = function(height){
      this.div.style.height = height
    }
    this.setWidth = function(width){
      this.div.style.width = width
    }
    this.setIp = function(ip){
      this.inputSpan.textContent = ip + " > "
    }

    this.inputLine.appendChild(this.inputSpan)
    this.inputLine.appendChild(this.inputShow)
    this.inputLine.appendChild(this.cursor)
    this.inside.appendChild(this.output)
    this.inside.appendChild(this.inputLine)
    this.div.appendChild(this.inside)

    document.body.appendChild(this.inputElem)

    this.setBackgroundColor('black')
    this.setTextColor('white')
    this.setWidth('100%')
    this.setHeight('100%')

    this.inputElem.style.position = 'absolute'
		this.inputElem.style.zIndex = '-100'
		this.inputElem.style.outline = 'none'
		this.inputElem.style.border = 'none'
		this.inputElem.style.opacity = '0'
    this.inputElem.style.fontSize = '0.2em'

    this.inputSpan.textContent = "notConnected > "

    this.cursor.style.background = 'white'
    this.cursor.innerHTML = 'I'
    this.cursor.style.visibility = 'hidden'
    this.inputShow.style.display = 'inline'
    this.inputShow.style.margin = '0'
    this.inputLine.style.margin = '0'
    this.output.style.margin = '0'
    this.inside.style.padding = '5px'
    this.div.style.overflow = 'auto'
    this.div.style.wordWrap = "break-word"
  }
  return constructor
}())
