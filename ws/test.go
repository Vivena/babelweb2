// func ConnectGroup(updates chan parser.BabelUpdate, node string, wg *sync.WaitGroup) {
import (
	"babelweb2/parser"
	"log"
	"sync"
)

//MCUpdates multicast updates sent by the routine comminicating with the routers
func MCUpdates(updates chan parser.BabelUpdate, g *Listenergroupe,
	wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		update, quit := <-updates
		if !quit {
			log.Println("closing all channels")
			g.Iter(func(l *Listener) {
				close(l.conduct)
			})
			wg.Done()
			return
		}
		if !(Db.Bd.CheckUpdate(update)) {
			continue
		}
		Db.Lock()
		err := Db.Bd.Update(update)
		if err != nil {
			log.Println(err)
		}
		Db.Unlock()
		t := update.ToS()
		g.Iter(func(l *Listener) {
			l.conduct <- t
		})
		//TODO unlock()
		if (t.action == "add") && (t.table == "interface") {
			newInterface := t.entry["ipv6"]
			go connectGroup(newInterface, wg)
		}
	}
}
var ConductGroup := make([]parser.BabelUpdate, 3)

func connectGroup(node string, wg *sync.WaitGroup) {
	conduct := make(chan parser.BabelUpdate, ws.ChanelSize)
	ConductGroup = append(ConductGroup, conduct)
	go Connection(conduct, node, wg)
}

func merge( updates chan parser.BabelUpdate) {
    var wg sync.WaitGroup
    output := func(c chan parser.BabelUpdate) {
        for n := range c {
            updates <- n
        }
        wg.Done()
    }
    wg.Add(len(ConductGroup))
    for _, c := range ConductGroup {
        go output(c)
    }
    go func() {
        wg.Wait()
        close(out)
    }()
}
