package transfer

import (
	"net"
	"log"
	"sync"
	"os"
	"testing"
	"google.golang.org/grpc"
	pb "github.com/transfer/proto"
)

func Test_server(t *testing.T)  {
	listen,err := net.Listen("tcp",addressFile)
	if err!=nil {
		log.Fatal("failed to listen :%v",err)
	}
	s := grpc.NewServer()
	FileService.server = &Server{ReadDirectory:"/home/li"}
	FileService.session = &Session{mu: &sync.Mutex{},files:make(map[SessionId]*os.File)}
	pb.RegisterFileTransferServer(s,FileService)
	s.Serve(listen)
}
