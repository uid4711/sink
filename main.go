package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/ufuchs/itplus/base/zvous"
	"github.com/ufuchs/zeroconf"
)

type readResult struct {
	color []byte
	buf   []byte
	n     int
	err   error
}

const (
	SERVICE_NAME = "_flood._zyx._tcp"
	BUFSIZE      = 8192
	ROUNDS       = 10000
)

//
//
//
func newReadResult(color string) *readResult {
	return &readResult{
		color: []byte(color),
		buf:   make([]byte, BUFSIZE, BUFSIZE*2),
	}
}

//
//
//
func run(conn net.Conn) {
	var (
		//text     = make([]byte, BUFSIZE, BUFSIZE*2)
		//firstRun = true
		wg     sync.WaitGroup
		textOK = false
	)

	var (
		INa      = newReadResult("green")
		INb      = newReadResult("red")
		INin     *readResult
		INout    *readResult
		INtoggle = false
	)

	var (
		i             uint64
		j             int
		OUTa          = make([]byte, ROUNDS*100, ROUNDS*100*2)
		OUTb          = make([]byte, ROUNDS*100, ROUNDS*100*2)
		OUTin, OUTout *[]byte
		OUTtoggle     = false
	)
	OUTb = OUTb[:0]
	OUTin = &OUTb
	OUTout = &OUTa

	//text = text[:0]

	for {

		if INtoggle {
			INin = INb
			INout = INa
		} else {
			INin = INa
			INout = INb
		}
		INtoggle = !INtoggle

		wg.Add(2)

		go func(r *readResult, wg *sync.WaitGroup) {
			r.n, r.err = conn.Read(r.buf)
			//			fmt.Println(r.n, string(r.buf))
			wg.Done()
		}(INin, &wg)

		go func(r *readResult, text *[]byte, wg *sync.WaitGroup) {

			// if textOK {
			// 	text = text[:0]
			// 	textOK = false
			// }

			for i := 0; i < r.n; i++ {

				var c = r.buf[i]

				switch c {
				case 0:
					continue
				case 10:
					*text = append(*text, c)
					if i < r.n {

						continue
					}
					textOK = true
					fmt.Println("...")
					goto end
				default:
					*text = append(*text, c)
				}

			}

		end:

			wg.Done()

		}(INout, OUTin, &wg)

		wg.Wait()

		//fmt.Println(string(text))

		if i == ROUNDS {

			if OUTtoggle {
				OUTb = OUTb[:0]
				OUTin = &OUTb
				OUTout = &OUTa
			} else {
				OUTa = OUTa[:0]
				OUTin = &OUTa
				OUTout = &OUTb
			}
			OUTtoggle = !OUTtoggle

			go func(b *[]byte, j int) {
				s := fmt.Sprintf("aaa-%02d.txt", j)
				ioutil.WriteFile(s, *b, 0644)
			}(OUTout, j)

			j++
			i = 0
		} else {
			i++
		}

	}

}

//
//
//
func main() {

	var (
		hostAndPort string
		wg          sync.WaitGroup
	)

	discovery := zvous.NewZCBrowserService(SERVICE_NAME, zeroconf.IPv4, 4)

	wg.Add(1)

	go func() {

		defer wg.Done()

		for {
			select {
			case conns := <-discovery.Out:

				for _, conn := range conns {

					hostAndPort = conn.ExtractConn()
					return

				}

			}
		}
	}()

	wg.Wait()

	discovery.Close()

	conn, err := net.Dial("tcp", hostAndPort)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()

	run(conn)

}
