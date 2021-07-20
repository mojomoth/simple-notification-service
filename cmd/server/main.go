package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/dev-zipida.com/simple-notification-service/protos/notification"
	"google.golang.org/grpc"
)

const (
	address = "localhost:5001"
)

type Server struct {
	clients       map[string]*pb.ClientDetail
	clientStreams map[string]*pb.Notification_ConnectToServerServer
	pb.UnimplementedNotificationServer
}

func (server *Server) init() {
	server.clients = make(map[string]*pb.ClientDetail)
	server.clientStreams = make(map[string]*pb.Notification_ConnectToServerServer)
}

func (server *Server) addNewClient(in *pb.ClientDetail, stream *pb.Notification_ConnectToServerServer) {
	log.Println("adding new client")
	server.clientStreams[in.ClientName] = stream
	server.clients[in.ClientName] = in
	// log.Println(server.clients) // <------------ 이걸 추가하면 정상
}

func (server *Server) ConnectToServer(in *pb.ClientDetail, stream pb.Notification_ConnectToServerServer) error {
	server.addNewClient(in, &stream)
	for {
		// time.Sleep(time.Second)
	}
	return nil
}

func (server *Server) sendNotification(clientID string, msg string) {
	client := server.clients[clientID]
	stream := server.clientStreams[clientID]

	notificationMessage := &pb.NotificationMessage{
		Message: fmt.Sprintf("%s(age : %d) currently living in %s :: %s", client.ClientName, client.ClientAge, client.Address, msg),
		Time:    time.Now().UnixNano(),
	}

	(*stream).Send(notificationMessage)
}

func main() {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := &Server{}
	server.init()

	options := []grpc.ServerOption{}
	options = append(options, grpc.MaxMsgSize(100*1024*1024))
	options = append(options, grpc.MaxRecvMsgSize(100*1024*1042))
	s := grpc.NewServer(options...)

	pb.RegisterNotificationServer(s, server)
	// go routine to get server notification essage from stdin
	// go waitForMessage(server)
	go Test(server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}

// 1초에 한번씩 체크하는 함수 만들어보기
func Test(server *Server) {
	for {
		fmt.Println(server.clients)
		time.Sleep(time.Second)
	}
}

func waitForMessage(server *Server) {
	for {
		// get the server notification message
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Notification Msg : ")
		msg, _ := reader.ReadString('\n')

		// send the message to all the clients
		for clientID := range server.clients {
			server.sendNotification(clientID, msg)
		}
	}
}
