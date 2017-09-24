package transfer

import (
	"errors"
	pb "github.com/transfer/proto"
	"golang.org/x/net/context"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Server struct {
	Addr          string
	ReadDirectory string
}
type fileService struct {
	server  *Server
	session *Session
}

//	if err := rpc.Register(&Rpc{server: srv, session: session}); err != nil {

var FileService = fileService{}

func (f fileService) Open(ctx context.Context, in *pb.FileRequest) (*pb.Response, error) {
	path := filepath.Join(f.server.ReadDirectory, in.Filename)
	println(path)
	file, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res := pb.Response{}
	res.Id = int64(f.session.Add(file))
	res.Result = true

	return &res, nil
}

func (f fileService) Close(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	f.session.Delete(SessionId(in.Id))
	res := new(pb.Response)
	res.Result = true
	res.Id = in.Id
	log.Printf("Close sessionId=%d", res.Id)
	return res, nil
}

func (f fileService) Stat(ctx context.Context, in *pb.FileRequest) (*pb.StatResponse, error) {
	path := filepath.Join(f.server.ReadDirectory, in.Filename)
	res := pb.StatResponse{}
	if fi, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	} else {

		if fi.IsDir() {
			res.Type = "Directory"
		} else {
			res.Type = "File"
			res.Size = fi.Size()
		}
		res.LastModified = fi.ModTime().String()
	}

	log.Printf("Stat %s, %#v", in.Filename, res)
	return &res, nil
}
func (f fileService) ReadAt(ctx context.Context, in *pb.ReadRequest) (*pb.ReadResponse, error) {
	file := f.session.Get(SessionId(in.Id))
	if file == nil {
		return nil, errors.New("You must call open first.")
	}
	res := pb.ReadResponse{}

	res.Date = make([]byte, in.Size)
	n, err := file.ReadAt(res.Date, in.Offset)
	if err != nil && err != io.EOF {
		return nil, err
	}

	if err == io.EOF {
		res.EOF = true
	}

	res.Size = int64(n)
	res.Date = res.Date[:n]

	log.Printf("ReadAt sessionId=%d, Offset=%d, n=%d,%d", in.Id, in.Offset, res.Size, cap(res.Date))

	return &res, nil
}
