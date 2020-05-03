package main

import (
	// "runtime"
	"fmt"
	"time"
	"math/rand"
	"strings"
	"strconv"
	"sync"
	// "bufio"
	// "os"
)

type Pair struct {
    a, b interface{}
}
type Triple struct {
    a, b, c interface{}
}
type Request struct{
	a string 
	b int
	c int
	d time.Time
	e time.Duration
}

type untypedRequest struct{
	a, b, c interface{} // a for string message, b for data, c for int position
}

func short_task_chan(short chan time.Duration){
	shortamt := 10
	// amt := time.Duration(rand.Intn(1000)+shortamt)
	amt := time.Duration(shortamt)
	time.Sleep(time.Millisecond * amt)
	short <- time.Millisecond * amt
}

func medium_task_chan(med chan time.Duration){
	medamt := 1000
	// amt := time.Duration(rand.Intn(1000)+medamt)
	amt := time.Duration(medamt)
	time.Sleep(time.Millisecond * amt)
	med <- time.Millisecond * amt
}

func long_task_chan(long chan time.Duration){
	longamt := 7000
	// amt := time.Duration(rand.Intn(1000)+longamt)
	amt := time.Duration(longamt)
	time.Sleep(time.Millisecond * amt)
	long <- time.Millisecond * amt
}

func short_task(){
	shortamt := 10
	// amt := time.Duration(rand.Intn(1000)+shortamt)
	amt := time.Duration(shortamt)
	time.Sleep(time.Millisecond * amt)
	// short <- amt
}

func medium_task(){
	medamt := 1000
	// amt := time.Duration(rand.Intn(1000)+medamt)
	amt := time.Duration(medamt)
	time.Sleep(time.Millisecond * amt)
	// med <- amt
}

func long_task(){
	longamt := 7000	
	// amt := time.Duration(rand.Intn(1000)+longamt)
	amt := time.Duration(longamt)
	time.Sleep(time.Millisecond * amt)
	// long <- amt
}

func do_tasks(timings chan map [int]time.Duration,position int){
	clientstr := "Client " + strconv.Itoa(position) + " "
	var nTasks int = 3
	personal_laundry_list := make(map [int]time.Duration)
	var short chan time.Duration = make(chan time.Duration, 10)
	var med chan time.Duration = make(chan time.Duration, 10)
	var long chan time.Duration = make(chan time.Duration, 10)
	// short = append(short,make(chan time.Duration))
	go short_task_chan(short)
	// short = append(short,make(chan time.Duration))
	go medium_task_chan(med)
	// short = append(short,make(chan time.Duration))
	go long_task_chan(long)
	mainloop:
	for{
		select{
		case shortreply := <- short:
			// personal_laundry_list = append(personal_laundry_list, Pair{shortreply,"short"})
			personal_laundry_list[0] = shortreply
			fmt.Println("short reply received")
			// close(short)
		case medreply := <- med:
			// personal_laundry_list = append(personal_laundry_list, Pair{medreply,"med"})
			personal_laundry_list[1] = medreply
			fmt.Println("med reply received")
			// close(med)
		case longreply := <- long:
			// personal_laundry_list = append(personal_laundry_list, Pair{longreply,"long"})
			personal_laundry_list[2] = longreply
			fmt.Println("long reply received")
			// close(long)
		default:
			if len(personal_laundry_list) == nTasks{
				fmt.Println("all internal replies received for " + clientstr + ",breaking loop")
				break mainloop
			}
		}
	}
	timings <- personal_laundry_list
	fmt.Println(clientstr + "list sent")
}

// func collect_rtt_arrays(wchans map [int]chan untypedRequest, position int){
// 	// fmt.Printf("Pinging nodes %d\n", i)
// 	clientstr := "Client " + strconv.Itoa(position) + " "
// 	request_ping := untypedRequest{clientstr + "requesting for ping",time.Now(),position} // 0 means request
// 	for k := range wchans{
// 		wchans[k] <- request_ping
// 	}
// }


func get_rtts(wchans map [int]chan untypedRequest, position int){
	// fmt.Printf("Pinging nodes %d\n", i)
	clientstr := "Client " + strconv.Itoa(position) + " "
	request_ping := untypedRequest{clientstr + "requesting for ping",time.Now(),position} // 0 means request
	for k := range wchans{
		wchans[k] <- request_ping
	}
}
func send_heartbeat(pingchan chan int, position int) {
	//TODO: add command chan to kill debugger
	if position != 1{
		for {	
			time.Sleep(2000*time.Millisecond)
			pingchan <- 0
			fmt.Println("Pinging my client")
		}
	} else {
		time.Sleep(2000*time.Millisecond)
		fmt.Println("Not pinging client")
	}

}
func receive_heartbeat(pingchan chan int, rchan chan Request, wchans map [int]chan untypedRequest, position int) {
	// clientstr := "Client " + strconv.Itoa(position) + " "
	var nullDuration time.Duration
	counter := 0
	loop:
		for {
			time.Sleep(2000*time.Millisecond)
			select {
			case <- pingchan:
				fmt.Println("Got ping from debugger")
			default:
				fmt.Println("Did not receive ping")
				counter += 1
				if counter > 10 {
					rchan <- Request{"Debugger"+ strconv.Itoa(position) + " is dead", 0,position,time.Now(),nullDuration}
					break loop

				}
			}

		}
		fmt.Println("stopped receiving heartbeats for now; polling for takeover of responsibilities")
}

func debugger(pingchan chan int, rchan chan untypedRequest, wchans map [int]chan untypedRequest, dcchan chan untypedRequest, otherclientchans map [int] chan Request, position int){
	clientstr := "Client " + strconv.Itoa(position) + " "
	dstr := "Debugger for " + clientstr 
	var personal_laundry_list map [int]time.Duration
	// var other_laundry_lists []Pair
	// fmt.Println(clientstr + "initialised; ")
	duration := make(chan map [int]time.Duration,2)
	rtt := make(map [int]time.Duration)
	rtt_map := make(map [int]map[int]time.Duration)
	laundrylists := make(map [int]map[int]time.Duration)
	laundrylistmsg := "laundry list"
	rttrep := "rtt reply"
	rttreq := "requesting for ping"

	//for network flow
	var in_net_flow []uint64 
	in_net_flow = make([]uint64,len(wchans))
	var imtx = &sync.Mutex{}
	var out_net_flow []uint64
	out_net_flow = make([]uint64,len(wchans))
	var omtx = &sync.Mutex{}
	
	var nullDuration time.Duration
	// specifieddelay := 4000
	go do_tasks(duration,position)
	go get_rtts(wchans,position)
	fmt.Printf("INITIALISING on node %d with wchans %d\n", position,wchans)
	initloop:
		for{
			select{
			case timings := <- duration:
				personal_laundry_list = timings
				for  k := range wchans{
					// fmt.Printf("%T,%T\n",k,v)
					wchans[k] <- untypedRequest{laundrylistmsg,personal_laundry_list,position}
				}
				fmt.Println("personal list for " + clientstr, personal_laundry_list)
				laundrylists[position] = personal_laundry_list
			case msg := <- rchan:
				// fmt.Println(clientstr + "msg received, msg : ", msg)
				if strings.Contains(msg.a.(string),laundrylistmsg){
					some_other_laundry_list := msg.b.(map [int]time.Duration)
					// other_laundry_lists = append(other_laundry_lists,msg.b)
					laundrylists[msg.c.(int)] = some_other_laundry_list
					fmt.Println("other_laundry_lists received by client " + strconv.Itoa(position) + " : ", laundrylists)
				} else if strings.Contains(msg.a.(string), rttrep){
					fmt.Printf("Updating rtt for node %d\n", msg.c.(int))
					now := time.Now()
					this_rtt := now.Sub(msg.b.(time.Time)) // Sub function is simply subtraction
					rtt[msg.c.(int)] = this_rtt
					//fmt.Printf("Node %d has updated rtt: %d\n", id, rtt)
					if len(rtt) == len(wchans) {
						fmt.Printf("DONE on node %d with rtts %d\n", position, rtt)
						rtt_map[position] = rtt
						my_rtt := untypedRequest{"rtt map reply",rtt,position}
						for k := range wchans{
							wchans[k] <- my_rtt
							// fmt.Println("SENT THE RTTTAYEEEEEE")
						}
					}
				} else if strings.Contains(msg.a.(string),"rtt map reply"){
					fmt.Println("GOT THE RTTTAYEEEEEE")
					rtt_map[msg.c.(int)] = msg.b.(map [int]time.Duration)
					if len(rtt_map) == len(wchans) {
						// fmt.Println("DONE WITH THE RTTATYTEEEEEEE")
						fmt.Println(rtt_map)
					}
				} else if strings.Contains(msg.a.(string), rttreq){
					fmt.Printf("RTT Replying to node %d, %T\n", msg.c.(int),wchans[msg.c.(int)])
					// fmt.Printf("%T\n",])
					reply_ping := untypedRequest{clientstr + rttrep, msg.b, position} // 1 means reply					
					// wchans[msg.c.(int)] <- reply_ping
					for k := range wchans{
						if k == msg.c.(int){
							wchans[k] <- reply_ping
						}
					}
					// fmt.Println(clientstr + "rtt reply sent")
					// if msg.c.(int) > position{
					// 	wchans[msg.c.(int)-1] <- reply_ping
					// } else{
					// 	wchans[msg.c.(int)] <- reply_ping
					// }
				} 
				// else if strings.Contains(msg.a.(strings),"DIE"){
				// 	fmt.Println(clientstr + "die command received; simulating debugger death")
				// }
			default:
				if len(rtt) == len(wchans) && len(laundrylists) == len(wchans)+1{
					nrtts := "number of rtts: " + strconv.Itoa(len(rtt)) + " "
					nlls := "number of laundry lists: " + strconv.Itoa(len(laundrylists)) + " "
					fmt.Println(clientstr + "DONE WITH INITIALISATION : " + nrtts + nlls)
					dcchan <- untypedRequest{"ready",0,0}
					break initloop
				}	else{
					nrtts := "number of rtts: " + strconv.Itoa(len(rtt)) + " "
					nlls := "number of laundry lists: " + strconv.Itoa(len(laundrylists)) + " "
					// fmt.Println(clientstr + "INCOMPLETE : " + nrtts + nlls)
					fmt.Printf("INCOMPLETE on node %d with %s rtts %d, %s laundrylists %d\n", position, nrtts, rtt, nlls, laundrylists)
					amt := time.Duration(500)
					time.Sleep(time.Millisecond * amt)
				}
			}
		}
	// runloop:
	go send_heartbeat(pingchan,position)

	for{
			select{
			case clientmsg := <- rchan:
				fmt.Println(dstr + "received msg :", clientmsg)
				if strings.Contains(clientmsg.a.(string),"COMPLETE"){
					fmt.Println(dstr + "COMPLETE :", clientmsg)
					task := clientmsg.c.(Triple).a.(int)
					ctaskrtt := clientmsg.c.(Triple).b.(time.Duration)
					ctasktime := clientmsg.c.(Triple).c.(time.Duration)
					processed_by := clientmsg.b.(Pair).a.(int)
					tasknum := clientmsg.b.(Pair).b.(int)
					okmsg := "Everything within expected parameters, both less than 5ms gap"
					notok := "Task "+ strconv.Itoa(tasknum) +" "
					rttdifference := ctaskrtt- rtt_map[position][processed_by]
					taskdifference := ctasktime - laundrylists[processed_by][task]
					if rttdifference>5000000{	//some arbitrary number of seconds
						notok = notok + "RTT taking too long, "+ strconv.Itoa(int(rttdifference.Milliseconds())) + " milliseconds longer than expected. "
					}
					if taskdifference>5000000{
						notok = notok + "Task taking too long, "+ strconv.Itoa(int(taskdifference.Milliseconds())) + " milliseconds longer than expected. "
					}
					if notok != "" {
						dcchan <- untypedRequest{notok,ctaskrtt,ctasktime}	
					} else{
						dcchan <- untypedRequest{okmsg,rttdifference,taskdifference}
					}
					fmt.Println("CHECKING",ctaskrtt.Milliseconds(),rtt_map[position][processed_by].Milliseconds())
				} else if strings.Contains(clientmsg.a.(string),"re-pinging for stability check"){
					for k := range otherclientchans{
						if k == clientmsg.c.(int){
							otherclientchans[k] <- Request{"reply to ping " + strconv.Itoa(clientmsg.b.(int)) + " for stability check",clientmsg.b.(int),position,time.Now(),nullDuration}
						}
					}


				} else if clientmsg.a.(string) == "count in" {
					imtx.Lock()
					in_net_flow[clientmsg.c.(int)]++
					imtx.Unlock()
					fmt.Println("incremented in")


					go func () {
						time.Sleep(time.Millisecond*1000) // get flow rate for 1min
						imtx.Lock()
						in_net_flow[clientmsg.c.(int)]--
						imtx.Unlock()
						fmt.Println("decremented in")

					} ()
					fmt.Println(in_net_flow)
				} else if clientmsg.a.(string) == "count out" {
					omtx.Lock()
					out_net_flow[clientmsg.c.(int)]++
					omtx.Unlock()
					fmt.Println("incremented out")
					go func () {
						time.Sleep(time.Millisecond*1000) // get flow rate for 1min
						omtx.Lock()
						out_net_flow[clientmsg.c.(int)]--
						omtx.Unlock()
						fmt.Println("decremented out")

					} ()
					fmt.Println(out_net_flow)
				}
			// case msg := <- rchan:
			// 	fmt.Println(dstr + "received msg :", msg.a.(string))
			default:
				// nrtts := "number of rtts: " + strconv.Itoa(len(rtt)) + " "
				// nlls := "number of laundry lists: " + strconv.Itoa(len(laundrylists)) + " "
				// fmt.Println(dstr + nrtts + nlls + "running..........")
				// amt := time.Duration(50000)
				// time.Sleep(time.Millisecond * amt)
			}
		}
	// rtt := make(map [int]time.Duration)
	// n := len(ping_channels)
}

// func find_new_debugger(){

// }
func findByValue(m map[int]Pair, value time.Time) int {
	for key, val := range m {
		if val.b == value {
			return key
		}
	}
	return -1
}

func find_most_stable_debugger(timemap map [int][]Triple) int{
	averages := make(map [int]int)
	variance := make(map [int]int)
	// var variance map [int]int
	for i:= range timemap{
		sum:= 0
		for j:= range timemap[i]{
			sum += int(timemap[i][j].c.(time.Duration).Nanoseconds())
		}
		average := sum/len(timemap[i])
		averages[i] = average
	}
	for q:= range averages{
		var sum_sqdifferences int
		for w := range timemap[q]{
			sum_sqdifferences += (int(timemap[q][w].c.(time.Duration).Nanoseconds()) - averages[q])*(int(timemap[q][w].c.(time.Duration).Nanoseconds()) - averages[q])
		}
		variance[q] = sum_sqdifferences/(len(timemap[q])-1)
	}
	lowest_value := variance[0]
	most_stable := 0
	for e,r:= range variance{
		if lowest_value<r{
			lowest_value = r
			most_stable = e
		}
	}
	return most_stable
}

func client(pingchans []chan int, rchan chan Request, wchan chan Request, position int, drchan chan untypedRequest, dwchans map [int]chan untypedRequest, otherclientchans map [int]chan Request) {
	var nullTime time.Time
	fmt.Println("Initialised: " + strconv.Itoa(position), len(dwchans))
	clientstr := "Client " + strconv.Itoa(position) + " "
	// duration = debugger_initialisation(rchan,wchan,position)
	preparedstr := "heylo client, this is server"
	// var tasks []Pair
	// cdchan := make(chan untypedRequest,10)
	dcchan := make(chan untypedRequest,10)
	debugger_id := position
	go debugger(pingchans[debugger_id], drchan, dwchans,dcchan,otherclientchans, position)
	completed_tasks := make(map [int]Triple)
	requested_tasks := make(map [int]Pair)
	specifieddelay := 50000
	var tasklist []string
	tasklist = append(tasklist,"short")
	tasklist = append(tasklist,"medium")
	tasklist = append(tasklist,"long")
	// var nullTime time.Time
	var nullDuration time.Duration
	// delay := time.Duration(1000)
	// var dead bool
	// var shifted_responsibility int
	timer := 0
	initloop:
		for{
			select{
			case rmsg:= <- rchan:
				// fmt.Println(strconv.Itoa(position) + " INITIALISATION message received! Message: " + rmsg.a)
				if rmsg.a == preparedstr{
					preparedreply := "hihi, this is client " +strconv.Itoa(position) + "! :D"
					initint := 0
					drchan <- untypedRequest{clientstr + "can talk to " +strconv.Itoa(rmsg.c), rmsg.c, position}
					wchan <- Request{preparedreply,initint,position,nullTime,nullDuration}
				}
				case dmsg:= <- dcchan:
					if dmsg.a == "ready"{
						break initloop
					}
				}
			}
		task_count := 0
		go receive_heartbeat(pingchans[debugger_id],rchan,dwchans,position)

	// runloop:

		for{
			select{
			case rmsg:= <- rchan:
				if strings.Contains(rmsg.a,"task done"){
					fmt.Println("Reply for task " + strconv.Itoa(rmsg.b) + " received", rmsg)
					endtime := time.Now()
					tasknumber := findByValue(requested_tasks,rmsg.d)
					taskcomplete := clientstr + " REQUESTED TASK " + strconv.Itoa(tasknumber) + " COMPLETE: duration: " + strconv.Itoa(int(rmsg.e.Nanoseconds()))
					taskrtt := endtime.Sub(requested_tasks[tasknumber].b.(time.Time))-rmsg.e   // endtime minus start time minus time taken to run task
					completed_tasks[tasknumber] = Triple{rmsg.b,taskrtt,rmsg.e}  //Pair of two durations: rtt for this specific task (minus task duration, so purely as a measure of network speed), and task timing
					drchan <- untypedRequest{taskcomplete,Pair{rmsg.c,tasknumber},completed_tasks[tasknumber]}
					//report to debugger about network in
					drchan <- untypedRequest{"count in",0,rmsg.c,}

				} else if strings.Contains(rmsg.a,"dead") {
					fmt.Println(clientstr + "Detected dead debugger")
					//tell all clients that this debugger is down
					for i,_ := range dwchans {
						if i != position {
							dwchans[i] <- untypedRequest{"debugger " + strconv.Itoa(rmsg.c) + " is dead",0,position}
						}
					}
					nPings := 5
					timemap := make(map [int][]Triple)   // map of position to nPings of Triple of time.Time,time.Time,time.Duration {start,end,rtt}
					var nullTime time.Time
					var nullDuration time.Duration
					for j :=0; j<nPings; j++{
						for  k := range dwchans{
							// fmt.Printf("%T,%T\n",k,v)
							reping := clientstr + "re-pinging for stability check"
							selected_int_map := timemap[k]
							selected_int_map = append(selected_int_map,Triple{time.Now(),nullTime,nullDuration})
							timemap[k] = selected_int_map
							dwchans[k] <- untypedRequest{reping,j,position}
						}
					}
					tempchan := make(chan Request,10)
					
					// var checks [len(dwchans)]bool
					for{ 		
					select{			// BLOCKING; CLIENT WILL NOT PROCESS OTHER MESSAGES UNTIL NEW DEBUGGER IS ASSIGNED
					case rmsg:= <- rchan:
						if strings.Contains(rmsg.a,"reply to ping"){
							endTime := time.Now()
							rttcount := endTime.Sub(rmsg.d)
							fmt.Println("CURRENT TIMEMAP ", timemap)
							fmt.Println(timemap[rmsg.c][rmsg.b].b.(time.Time))
							timemap[rmsg.c][rmsg.b] = Triple{timemap[rmsg.c][rmsg.b].a.(time.Time),endTime,rttcount}
							// timemap[rmsg.c][rmsg.b].b.(time.Time) = endTime
							// timemap[rmsg.c][rmsg.b].c.(time.Duration) = rttcount
							// var checks []bool 
							fmt.Println("#######################START OF NEW CHECKLOOP#######")
							var check bool = true
							checkloop:
								for k := range timemap{
									for f := range timemap[k]{
										if timemap[k][f].b.(time.Time) == nullTime{
											fmt.Println("NOT DONE YET, " , timemap[k][f], timemap[k][f].b.(time.Time).String(), nullTime.String())
											check = false
											break checkloop
										}
										fmt.Println(check)
									}
								}
							if check == true{
								fmt.Println("ALL CHECKS RECEIVED : ", timemap)
								most_stable_debugger := find_most_stable_debugger(timemap)
								fmt.Println("MOST STABLE DEBUGGER : ", most_stable_debugger)
							} else{
								fmt.Println("CHECKS INCOMPLETE : ", timemap)
							}
						} else{
							tempchan <- rmsg
							}
						}
					}
				}
			case dmsg := <- dcchan :
					fmt.Println("Task report received",dmsg)
			default:
				if len(requested_tasks)<3{
					if timer == specifieddelay{
						fmt.Println(clientstr + "ready",timer)
						randomtask := rand.Intn(len(tasklist))
						// randomtask := current_task
						request := "request pls do a "+ tasklist[randomtask] +" task"
						reqstart := time.Now()
						wchan <- Request{request,randomtask,position,reqstart,nullDuration}
						//report to debugger about network flow out
						drchan <- untypedRequest{"count out",0,0}


						requested_tasks[task_count] = Pair{randomtask,reqstart}
						// amt := time.Duration(rand.Intn(9000))
						// time.Sleep(time.Millisecond * amt)
						timer = 0
						task_count +=1
						fmt.Println(strconv.Itoa(position) + " wants task " +tasklist[randomtask], timer)
					} else{
						// fmt.Println(clientstr + "running..........")
						timer += 1
					}
				}else{
				fmt.Println(clientstr + "done with requests, ", requested_tasks,completed_tasks)
				amt := time.Duration(5000)
				time.Sleep(time.Millisecond * amt)
				}
		}
	}
}

// func randomly_kill_debuggers(dchans []untypedRequest){
// 	randomlööp := rand.Intn(len(tasks))
// 	dchan[randomlööp] <- "DIE"
// }

func sreceiver(wchan chan Request,initchan chan Request, requestchan chan Request) {
	for {
	  	msg := <- wchan
		fmt.Println("Server Receiving: " +  msg.a + ", at position  " + strconv.Itoa(msg.c))
		initmsg := "hihi, this is client "
		requestmsg := "task"
		if strings.Contains(msg.a,initmsg){
			initreceived := "server received init message from client " + strconv.Itoa(msg.c)
			fmt.Println(initreceived)
			initchan <- msg
		}else if strings.Contains(msg.a,requestmsg){
			requestreceived := "server received REQUEST from client " + strconv.Itoa(msg.c)
			fmt.Println(requestreceived)
			requestchan <- msg
		}
	}
}

// main is the central server
func main() {
	// min := 10 //lower limit, as given, "at least 10"
	// max := rand.Intn(20) + min // randomising max number of clients to create, setting upper limit for range
	// nClients := rand.Intn(max - min + 1) + min //random number of clients
	nClients := 3
	// var nullTime time.Time
	var nullDuration time.Duration
	var froots []int
	var rchans []chan Request //read-only channel 
	var wchans []chan Request //write-only channel
	var dchans []chan untypedRequest
	var pingchans []chan int
	initchan := make(chan Request, 10)
	requestchan := make(chan Request, 10)
	// messywords <- go scompiler(messywords) //start your string compiler so the instant any receiver catches it, you can start adding to the static array, you want only ONE async process doing this
	for i := 0; i<nClients+1; i++ {
		fmt.Println(i)
		rchan := make(chan Request,10)
		rchans = append(rchans,rchan)
		wchan := make(chan Request,10)	   //buffer of size 1 to allow server to pick this up at its convenience and not lock everything up with buffer size 0
		wchans = append(wchans,wchan)
		dchans = append(dchans,make(chan untypedRequest, 10))
		froots = append(froots,i)
		pingchans = append(pingchans, make(chan int,10))
		go sreceiver(wchans[i],initchan,requestchan)
		// go mdispatcher(rchans[i],messywords)
	}
	for j := 1; j<nClients+1; j++ {
		// fmt.Println(j)
		dwchans := make(map [int]chan untypedRequest)
		drchan := dchans[j]
		rchan := rchans[j]
		wchan := wchans[j]
		otherclientchans := make(map [int]chan Request)
		// var dwchans []chan untypedRequest
		for k :=0; k<nClients+1;k++{
			if k != j{
				dwchans[k] = dchans[k]
				otherclientchans[k] = rchans[k]
				// dwchans = append(dwchans,dchans[k])
			} else{
				// fmt.Println("SKIPPING channel " + strconv.Itoa(k) + " for client " + strconv.Itoa(j))
			}
		}
		// fmt.Println(strconv.Itoa(j)+ " " + strconv.Itoa(len(wchans))+ " " + strconv.Itoa(nClients))
		otherclientchans[0] = wchan
		go client(pingchans, rchan, wchan,j,drchan,dwchans,otherclientchans)
	}
	cswchans := make(map [int]chan untypedRequest)
	csotherclientchans := make(map [int]chan Request)
	for k:=1; k<nClients+1;k++{
		cswchans[k] = dchans[k]
		csotherclientchans[k] = rchans[k]
	}
	// csdchan := make(chan untypedRequest,10)
	dcschan := make(chan untypedRequest,10)
	// fmt.Print(cswchans)
	go debugger(pingchans[0],dchans[0],cswchans,dcschan,csotherclientchans,0)
	// Init done
	fmt.Println("CENTRAL SERVER PROTOCOL START")
	fmt.Println(froots)
	for j :=0; j<nClients;j++{
		msg := "heylo client, this is server"
		rchans[j] <- Request{msg,0,0,time.Now(),nullDuration}
	}
	for{
		select{
		case request := <- requestchan:
			fmt.Println("MAINSERVER REQUEST : Request" + strconv.Itoa(request.b) + " from " + strconv.Itoa(request.c) + " : " + request.a)
			if strings.Contains(request.a,"short"){
				taskstart := time.Now()
				short_task()
				taskend := time.Now()
				timetaken := taskend.Sub(taskstart)
				rchans[request.c] <- Request{"short task done ",request.b,0,request.d,timetaken} // (str,request number,position,current time, timetaken)
				fmt.Println("MAINSERVER : COMPLETED REQUEST " + strconv.Itoa(request.b) + " from " + strconv.Itoa(request.c) + " COMPLETE in " + strconv.Itoa(int(timetaken.Nanoseconds())))
			} else if strings.Contains(request.a,"medium"){
				taskstart := time.Now()
				medium_task()
				taskend := time.Now()
				timetaken := taskend.Sub(taskstart)
				rchans[request.c] <- Request{"medium task done ",request.b,0,request.d,timetaken}
				fmt.Println("MAINSERVER : COMPLETED REQUEST " + strconv.Itoa(request.b) + " from " + strconv.Itoa(request.c) + " COMPLETE in " + strconv.Itoa(int(timetaken.Nanoseconds())))
			} else if strings.Contains(request.a,"long"){
				taskstart := time.Now()
				long_task()
				taskend := time.Now()
				timetaken := taskend.Sub(taskstart)
				rchans[request.c] <- Request{"long task done ",request.b,0,request.d,timetaken}
				fmt.Println("MAINSERVER : COMPLETED REQUEST " + strconv.Itoa(request.b) + " from " + strconv.Itoa(request.c) + " COMPLETE in " + strconv.Itoa(int(timetaken.Nanoseconds())))
			}
		case initmsg := <- initchan:
			fmt.Println(initmsg.a)
		}
	}
}