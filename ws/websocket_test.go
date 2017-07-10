package ws

import(
  "time"
  "net/http"
  "testing"
  "sync"
  "log"
)


func CMess(updates chan interface{}) {
	for {
		time.Sleep(1000000000)
		updates <- "salut"
	}

}

func TestWS(t *testing.T){
  var wg sync.WaitGroup
	wg.Add(2)

  updates := make(chan interface{},1024)

  log.Println("test")
	go CMess(updates)
	bcastGrp := NewListenerGroupe()
	go MCUpdates(updates, bcastGrp)
	ws := Handler(bcastGrp)

	http.Handle("/", http.FileServer(http.Dir(root)))
	http.Handle("/style.css", http.FileServer(http.Dir(cssPage)))
	http.Handle("/initialize.js", http.FileServer(http.Dir(jsPage)))
	http.Handle("/d3/d3.js", http.FileServer(http.Dir(d3)))
	http.Handle("/ws", ws)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
