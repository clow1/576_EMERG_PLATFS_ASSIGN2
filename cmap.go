package main

type ReduceResponse struct {
  s string
  i int
}

type ReduceMessage struct {
  astr string
  aint int
  funct ReduceFunc
  resp chan ReduceResponse
}

type Message struct {
  msg  string
  resp chan int
}

type ChannelMap struct {
  cm map[string]int
  stop chan int
  askchan chan Message
  addchan chan string
  reducechan chan ReduceMessage
}

func NewChannelMap() *ChannelMap {

  chanmap := new(ChannelMap)
  chanmap.cm = make(map[string]int,ASK_BUFFER_SIZE)
  chanmap.askchan = make(chan Message, ASK_BUFFER_SIZE)
  chanmap.addchan = make(chan string, ADD_BUFFER_SIZE)
  chanmap.reducechan = make(chan ReduceMessage)
  chanmap.stop = make(chan int)
  return chanmap
}

func NewLockingMap() *ChannelMap {
  return NewChannelMap()
}


//==============================================================================


// write the word to addchan
// when the word is obtained in Listen() it will be added
// without causing race conditions
func (chanmap ChannelMap) AddWord(word string) {
  chanmap.addchan <- word
}


func(chanmap ChannelMap) Reduce(f ReduceFunc, s string , acc int) (string, int) {
  resp := make(chan ReduceResponse)
  msg := ReduceMessage{s, acc, f, resp}
  chanmap.reducechan <- msg
  ret := <- msg.resp
  return ret.s, ret.i
}


// send the word to askchan along with a response channel
// wait for a value to be written into response chan
// return the response
func (chanmap ChannelMap) GetCount(word string) int {
  m := Message{msg: word, resp: make(chan int)}
  chanmap.askchan <- m
  r := <-m.resp
  return r
}


// just write a value to stop
func(chanmap ChannelMap) Stop() {
  chanmap.stop <- 0
}




func (chanmap ChannelMap) Listen() {
  for {
  select {
    case msg_ask := <- chanmap.askchan:
      var word = msg_ask.msg
      msg_ask.resp <- chanmap.cm[word]
    case msg_add := <- chanmap.addchan:
      if chanmap.cm[msg_add] != 0 {
        chanmap.cm[msg_add]++
      } else {
          chanmap.cm[msg_add] = 1
      }
    case msg_red := <- chanmap.reducechan:
      var redfunc = msg_red.funct
      var accum_int = msg_red.aint
      var accum_str = msg_red.astr
      for k,v := range chanmap.cm {
        accum_str, accum_int = redfunc(accum_str, accum_int, k, v)
      }
      resp_struct := ReduceResponse{accum_str, accum_int}
      msg_red.resp <- resp_struct
    case <-chanmap.stop:
      return
  }
  }

}
