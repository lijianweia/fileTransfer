package transfer

import (
	pb "github.com/transfer/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
	"testing"
)

func Test_server(t *testing.T) {
	listen, err := net.Listen("tcp", addressFileC)
	if err != nil {
		log.Fatal("failed to listen :%v", err)
	}
	s := grpc.NewServer(grpc.MaxRecvMsgSize(BLOCK_SIZE*2), grpc.MaxSendMsgSize(BLOCK_SIZE*2))
	FileService.server = &Server{ReadDirectory: "/home/li"}
	FileService.session = &Session{mu: &sync.Mutex{}, files: make(map[SessionId]*os.File)}
	pb.RegisterFileTransferServer(s, FileService)

	s.Serve(listen)
}
