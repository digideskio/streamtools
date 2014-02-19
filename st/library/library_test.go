package library

import (
	"github.com/nytlabs/streamtools/st/blocks" // blocks
	"log"
	"testing"
	"time"
)

func newBlock(id, kind string) (blocks.BlockInterface, blocks.BlockChans) {

	library := map[string]func() blocks.BlockInterface{
		"count":   NewCount,
		"toFile":  NewToFile,
		"fromNSQ": NewFromNSQ,
		"toNSQ":   NewToNSQ,
		"fromSQS": NewFromSQS,
	}

	chans := blocks.BlockChans{
		InChan:    make(chan *blocks.Msg),
		QueryChan: make(chan *blocks.QueryMsg),
		AddChan:   make(chan *blocks.AddChanMsg),
		DelChan:   make(chan *blocks.Msg),
		ErrChan:   make(chan error),
		QuitChan:  make(chan bool),
	}

	// actual block
	b := library[kind]()
	b.Build(chans)

	return b, chans

}

func TestToFromNSQ(t *testing.T) {
	log.Println("testing toNSQ")

	toB, toC := newBlock("testingToNSQ", "toNSQ")
	go blocks.BlockRoutine(toB)

	ruleMsg := map[string]interface{}{"Topic": "librarytest", "NsqdTCPAddrs": "127.0.0.1:4150"}
	toRule := &blocks.Msg{Msg: ruleMsg, Route: "rule"}
	toC.InChan <- toRule

	nsqMsg := map[string]interface{}{"Foo": "Bar"}
	postData := &blocks.Msg{Msg: nsqMsg, Route: "in"}
	toC.InChan <- postData

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		log.Println("quitting chan")
		toC.QuitChan <- true
	})

	log.Println("testing fromNSQ")

	fromB, fromC := newBlock("testingfromNSQ", "fromNSQ")
	go blocks.BlockRoutine(fromB)

	outChan := make(chan *blocks.Msg)
	fromC.AddChan <- &blocks.AddChanMsg{Route: "1", Channel: outChan}

	nsqSetup := map[string]interface{}{"ReadTopic": "librarytest", "LookupdAddr": "127.0.0.1:4161", "ReadChannel": "libtestchannel", "MaxInFlight": 100}
	fromRule := &blocks.Msg{Msg: nsqSetup, Route: "rule"}
	fromC.InChan <- fromRule

	time.AfterFunc(time.Duration(5)*time.Second, func() {
		fromC.QuitChan <- true
	})

	for {
		select {
		case message := <-outChan:
			log.Println("caught message on outChan")
			log.Println(message)

		case err := <-fromC.ErrChan:
			if err != nil {
				t.Errorf(err.Error())
			} else {
				return
			}
		}
	}
}

func TestCount(t *testing.T) {
	log.Println("testing Count")
	b, c := newBlock("testingCount", "count")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestToFile(t *testing.T) {
	log.Println("testing toFile")
	b, c := newBlock("testingToFile", "toFile")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestFromSQS(t *testing.T) {
	log.Println("testing FromSQS")
	b, c := newBlock("testingFromSQS", "fromSQS")
	go blocks.BlockRoutine(b)
	time.AfterFunc(time.Duration(5)*time.Second, func() {
		c.QuitChan <- true
	})
	err := <-c.ErrChan
	if err != nil {
		t.Errorf(err.Error())
	}
}
