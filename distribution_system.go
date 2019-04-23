package main

import (
    "flag"
    "fmt"
    "net"
    "strings"
    "strconv"
    "time"
    "math/rand"
    "encoding/json"
)

/* Metadata about node */
type NodeInfo struct {
    NodeId  int  `json:"nodeId"`
    NodeIpAddr  string  `json:"nodeIpAddr"`
    Port  string  `json:"port"`
}

/* A standard format for a Request/Response for adding node to cluster */
type AddToClusterMessage struct {
    Source NodeInfo  `json:"source"`
    Dest NodeInfo  `json:"dest"`
    Message string  `json:"message"`
}


/* The entry point for our System */
func main(){
    /* Parse the provided parameters on command line */
    makeMasterOnError := flag.Bool("makeMasterOnError", false, "make this node master if unable to connect to the cluster ip provided.")
    clusterip := flag.String("clusterip", "127.0.0.1:8001", "ip address of any node to connnect")
    myport := flag.String("myport", "8001", "ip address to run this node on. default is 8001.")
    flag.Parse()

    /* Generate id for myself */
    rand.Seed(time.Now().UTC().UnixNano())
    myid := rand.Intn(99999999)

    myIp,_ := net.InterfaceAddrs()
    me := NodeInfo{ NodeId: myid, NodeIpAddr: myIp[0].String(),  Port: *myport}
    dest := NodeInfo{ NodeId: -1, NodeIpAddr: strings.Split(*clusterip, ":")[0], Port: strings.Split(*clusterip, ":")[1]}
    fmt.Println("My details:", me.String())

    /* Try to connect to the cluster, and send request to cluster if able to connect */
    ableToConnect := connectToCluster(me, dest)

    /* 
       Listen for other incoming requests form other nodes to join cluster
    */
    if ableToConnect || (!ableToConnect && *makeMasterOnError) {
        if *makeMasterOnError {fmt.Println("Will start this node as master.")}
        listenOnPort(me)
    } else {
        fmt.Println("Quitting system. Set makeMasterOnError flag to make the node master.", myid)
    }
}


/* 
   This is a useful utility to format the json packet to send requests
 */
func getAddToClusterMessage(source NodeInfo, dest NodeInfo, message string) (AddToClusterMessage){
    return AddToClusterMessage{
        Source: NodeInfo{
                NodeId: source.NodeId,
                NodeIpAddr: source.NodeIpAddr,
                Port: source.Port,
                },
        Dest: NodeInfo{
                NodeId: dest.NodeId,
                NodeIpAddr: dest.NodeIpAddr,
                Port: dest.Port,
                },
        Message: message,
    }
}


func connectToCluster(me NodeInfo, dest NodeInfo) (bool){
    /* connect to this socket details provided */
    connOut, err := net.DialTimeout("tcp", dest.NodeIpAddr + ":" + dest.Port, time.Duration(10) * time.Second)
    if err != nil {
        if _, ok := err.(net.Error); ok {
            fmt.Println("Couldn't connect to cluster.", me.NodeId)
            return false
        }
    } else {
        fmt.Println("Connected to cluster. Sending message to node.")
        text := "Hi nody.. please add me to the cluster.."
        requestMessage := getAddToClusterMessage(me, dest, text)
        json.NewEncoder(connOut).Encode(&requestMessage)

        decoder := json.NewDecoder(connOut)
        var responseMessage AddToClusterMessage
        decoder.Decode(&responseMessage)
        fmt.Println("Got response:\n" + responseMessage.String())
        
        return true
    }
    return false
}



func listenOnPort(me NodeInfo){
    /* Listen for incoming messages */
    ln, _ := net.Listen("tcp", fmt.Sprint(":" + me.Port))
    /* accept connection on port */
    for{
        connIn, err := ln.Accept()
        if err != nil {
            if _, ok := err.(net.Error); ok {
                fmt.Println("Error received while listening.", me.NodeId)
            }
        } else {
            var requestMessage AddToClusterMessage
            json.NewDecoder(connIn).Decode(&requestMessage)
            fmt.Println("Got request:\n" + requestMessage.String())

            text := "Sure buddy.. too easy.."
            responseMessage := getAddToClusterMessage(me, requestMessage.Source, text)
            json.NewEncoder(connIn).Encode(&responseMessage)
            connIn.Close()
        }
    }
}