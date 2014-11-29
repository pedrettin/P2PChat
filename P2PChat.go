package main

 //worked with Matt Pozderac

import (
        "encoding/json"
        "fmt"
        "log"
        "net"
        "gopkg.in/qml.v0"
        "os"
        "strings"
        "sync"
)
 
const PORT = ":1500"
 
var (
        output chan string = make(chan string) //channel waitin on the user to type something
        listIPs map[string]string = make(map[string]string)//list of users IPS connected to me
        listConnections map[string]net.Conn = make(map[string]net.Conn)//list of users connections connected to me
        myName string //name of the client
        testing bool = true
        ctrl Control
        mutex = new(sync.Mutex)
)
 
type Control struct {
        Root        qml.Object
        convstring  string
        userlist    string
        inputString string
}
 
//message sent out to the server
type Message struct {
        Kind      string //type of message ("CONNECT","PRIVATE","PUBLIC","DISCONNECT","ADD")
        Username  string //my username
        IP        string //Ip address of my computer
        MSG       string //message
        Usernames []string //usernames of people connected
        IPs               []string //IP addresses of all the users connected
}
 
//start the connection, introduces the user to the chat and creates graphical interface.
func main() {
        //adding myself to the list
        myName= os.Args[2]
       
        //starting graphics
        qml.Init(nil)
        engine := qml.NewEngine()
        ctrl = Control{convstring: ""}
        ctrl.convstring = ""
        context := engine.Context()
        context.SetVar("ctrl", &ctrl)
        component, err := engine.LoadFile("chat.qml")
        if err != nil {
                fmt.Println("no file to load for ui")
                fmt.Println(err.Error())
                os.Exit(0)
        }
       
        win := component.CreateWindow(nil)
        ctrl.Root = win.Root()
       
        win.Show() //show window
        ctrl.updateText("Hello "+myName+".\nFor private messages, type the message followed by * and the name of the receiver.\n To leave the conversation type disconnect")
       
        go server()//starting server
        if os.Args[1]!="127.0.0.1"{go introduceMyself(os.Args[1])} //connect to the first peer
        go userInput()
       
        win.Wait()
        closing:=createMessage("DISCONNECT",myName,getMyIp(),"",make([]string,0),make([]string,0))
        closing.send()
}
 
//part of the peer that acts like a server
 
//waits for possible peers to connect
func server(){
        if testing {log.Println("server")}
        tcpAddr, err := net.ResolveTCPAddr("tcp4", PORT)
        checkError(err)
        listener, err := net.ListenTCP("tcp", tcpAddr)
        checkError(err)
        for {
                conn, err := listener.Accept()
                if err != nil {
                        continue
                }
               
                go receive(conn)
        }
}
 
//receives message from peer
func receive(conn net.Conn){
        if testing {log.Println("receive")}
        defer conn.Close()
        dec:=json.NewDecoder(conn)
        msg:= new(Message)
        for {
                if err := dec.Decode(msg);err != nil {
                        return
                }
                switch msg.Kind{
                        case "CONNECT":
                                if testing {log.Println("Kind = CONNECT")}
                                if !handleConnect(*msg, conn){return}
                        case "PRIVATE":
                                if testing {log.Println("Kind = PRIVATE")}
                                ctrl.updateText("(private) from "+msg.Username+": "+msg.MSG)
                        case "PUBLIC":
                                if testing {log.Println("Kind = PUBLIC")}
                                ctrl.updateText(msg.Username+": "+msg.MSG)
                        case "DISCONNECT":
                                if testing {log.Println("Kind = DISCONNECT")}
                                disconnect(*msg)
                                return
                        case "HEARTBEAT"://ask about it in the morning
                                log.Println("HEARTBEAT")
                        case "LIST":
                                if testing {log.Println("Kind = LIST")}
                                connectToPeers(*msg)
                                return
                        case "ADD":
                                if testing {log.Println("Kind = ADD")}
                                addPeer(*msg)
                }
        }
}
 
//introduces peer to the chat
func introduceMyself(IP string){
        if testing {log.Println("introduceMyself")}
        conn:=createConnection(IP)
        enc:= json.NewEncoder(conn)
        introMessage:= createMessage("CONNECT", myName , getMyIp(), "", make([]string, 0), make([]string, 0))
        enc.Encode(introMessage)
        go receive(conn)
}
 
//handle a connection with a new peer
func handleConnect(msg Message, conn net.Conn) bool{
        if testing {log.Println("handleConnect")}
        Users,IPs:=getFromMap(listIPs)
        Users = append(Users, myName) //add my name to the list
        IPs = append(IPs, getMyIp()) //add my ip to the list
        response:=createMessage("LIST","","","",Users,IPs)
        if alreadyAUser(msg.Username){
                response.MSG="Username already taken, choose another one that is not in the list"
                response.send()
                return false
        }
        mutex.Lock()
        listIPs[msg.Username]=msg.IP
        listConnections[msg.Username]=conn
        mutex.Unlock()
        log.Println(listConnections)
        response.sendPrivate(msg.Username)
        return true
}
 
//connects with everyone in the chat. The message passed in contains a list of users and ips
func connectToPeers(msg Message) {
        for index, ip := range msg.IPs {
                conn:=createConnection(ip)
                mutex.Lock()
                listIPs[msg.Usernames[index]]=ip
                listConnections[msg.Usernames[index]]=conn
                mutex.Unlock()
        }
        users,_:=getFromMap(listIPs)
        ctrl.updateList(users)
        addMessage := createMessage("ADD", myName, getMyIp(), "", make([]string, 0), make([]string, 0))
        addMessage.send()
}
 
//adds a peer to everyone list
func addPeer(msg Message){
        mutex.Lock()
        listIPs[msg.Username]=msg.IP
        conn:=createConnection(msg.IP)
        listConnections[msg.Username]=conn
        mutex.Unlock()
        userNames,_:=getFromMap(listIPs)
        ctrl.updateList(userNames)
        ctrl.updateText(msg.Username+" just joined the chat")  
}
 
//sends message to all peers
func (msg *Message) send(){
        if testing {log.Println("send")}
        if testing {log.Println(listConnections)}
        for _,peerConnection := range listConnections{
                enc:=json.NewEncoder(peerConnection)
                enc.Encode(msg)
        }
}
 
//sends message to a peer
func (msg *Message) sendPrivate(receiver string){
        if testing {log.Println("sendPrivate")}
        if alreadyAUser(receiver){
                peerConnection:=listConnections[receiver]
                enc:=json.NewEncoder(peerConnection)
                enc.Encode(msg)
        }else{
                ctrl.updateText(receiver+" is not a real user")
        }      
}
 
 
//disconnect user by deleting him/her from list
func disconnect(msg Message){
        mutex.Lock()
        delete(listIPs, msg.Username)
        delete(listConnections, msg.Username)
        mutex.Unlock()
        newUserList, _ := getFromMap(listIPs)
        ctrl.updateList(newUserList)
        ctrl.updateText(msg.Username + " left the chat")       
}

//returns two slices, the first one with the keys of the map and the second on with the values
func getFromMap(mappa map[string]string) ([]string, []string){
        var keys []string
        var values []string
        for key,value := range mappa{
                keys = append(keys,key)
                values = append(values,value)
        }
        return keys,values
}
 
//creates a new connection, given the IP address, and returns it
func createConnection(IP string) (conn net.Conn){
        service:= IP+PORT
        tcpAddr, err := net.ResolveTCPAddr("tcp", service)
        handleErr(err)
        conn, err = net.DialTCP("tcp", nil, tcpAddr)
        handleErr(err)
        return
}

//returns my ip
func getMyIp() (IP string){
        name, err := os.Hostname()
        handleErr(err)
        addr, err := net.ResolveIPAddr("ip", name)
        handleErr(err)
        IP = addr.String()
        return
}

//creates a new message using the parameters passed in and returns it
func createMessage(Kind string, Username string, IP string, MSG string, Usernames []string, IPs []string) (msg *Message) {
        msg = new(Message)
        msg.Kind = Kind
        msg.Username = Username
        msg.IP = IP
        msg.MSG = MSG
        msg.Usernames = Usernames
        msg.IPs = IPs
        return
}
 
//sends message to the server
func userInput(){
        if testing {log.Println("userInput")}
        msg:=new(Message)
        for {
                message:= <-output
                whatever:=strings.Split(message,"*")
                if message=="disconnect"{
                        msg=createMessage("DISCONNECT",myName,"","", make([]string, 0), make([]string, 0))
                        msg.send()
                        break
                } else if len(whatever)>1 {
                        msg=createMessage("PRIVATE",myName,"",whatever[0], make([]string, 0), make([]string, 0))
                        msg.sendPrivate(whatever[1])
                        ctrl.updateText("(private) from "+myName+": "+msg.MSG)
                } else {
                        msg=createMessage("PUBLIC",myName,"",whatever[0], make([]string, 0), make([]string, 0))
                        msg.send()
                        ctrl.updateText(myName+": "+msg.MSG)
                }
        }
        os.Exit(1)
}
 
//handles errors
func handleErr(err error) {
        if err != nil {
                log.Println("No one in the chat yet")
        }
}
 
//check errors
func checkError(err error) {
        if err != nil {
                fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
                os.Exit(1)
        }
}
 
//checks to see if a userName is already in the list
func alreadyAUser(user string) bool {
        for userName,_:= range listIPs {
                if userName == user {return true}
        }
        return false
}
 
 
//Graphics methods
 
func (ctrl *Control) TextEntered(text qml.Object) {
        //this method is called whenever a return key is typed in the text entry field.  The qml object calls this function
        ctrl.inputString = text.String("text") //the ctrl's inputString field holds the message
        //you will want to send it to the server
        //but for now just send it back to the conv field
        //ctrl.updateText(ctrl.inputString)
        output <- ctrl.inputString
 
}
 
func (ctrl *Control) updateText(toAdd string) {
        //call this method whenever you want to add text to the qml object's conv field
        ctrl.convstring = ctrl.convstring + toAdd + "\n" //also keep track of everything in that field
        ctrl.Root.ObjectByName("conv").Set("text", ctrl.convstring)
        qml.Changed(ctrl, &ctrl.convstring)
}
 
func (ctrl *Control) updateList(list []string) {
        ctrl.userlist = ""
        for _, user := range list {
                ctrl.userlist += user + "\n"
        }
        ctrl.Root.ObjectByName("userlist").Set("text", ctrl.userlist)
        qml.Changed(ctrl, &ctrl.userlist)
}
