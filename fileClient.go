package transfer

import (
	"context"
	pb "github.com/transfer/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
)

type Client struct {
	Addr string
	cli  pb.FileTransferClient
}

func NewClient(addr string) *Client {
	return &Client{Addr: addr}

}
func (c *Client) Dial() error {

	conn, err := grpc.Dial(AddressFileC, grpc.WithInsecure(), grpc.WithMaxMsgSize(BLOCK_SIZE*2))
	if err != nil {
		return err
	}
	c.cli = pb.NewFileTransferClient(conn)
	return nil
}

func (c *Client) Open(filename string) (SessionId, error) {
	req := new(pb.FileRequest)
	req.Filename = filename

	response, err := c.cli.Open(context.Background(), req)
	if err != nil {
		log.Println("---------------------")
		return -1, err
	}
	return (SessionId(response.Id)), nil
}

func (c *Client) Stat(filename string) (*pb.StatResponse, error) {
	req := new(pb.FileRequest)
	req.Filename = filename
	res, err := c.cli.Stat(context.Background(), req)
	return res, err
}
func (c *Client) WriteAt(sessionId SessionId, buf []byte, offset int64, size int64) (int, error) {
	res := &pb.WriteResponse{}
	writeReq := new(pb.WriteRequest)
	writeReq.Id = int64(sessionId)
	writeReq.Offset = offset
	writeReq.Size = int64(size)
	writeReq.Date = buf
	res, err := c.cli.WriteAt(context.Background(), writeReq)
	if err != nil {
		log.Println("in WriteAt", offset/BLOCK_SIZE, size)
		log.Println(err)
	}
	return int(res.Size), nil
}

func (c *Client) GetBlock(sessionId SessionId, blockId int) ([]byte, error) {
	log.Println("in GetBlock,blockId is", blockId)
	return c.ReadAt(sessionId, int64(blockId)*BLOCK_SIZE, BLOCK_SIZE)
}

func (c *Client) ReadAt(sessionId SessionId, offset int64, size int64) ([]byte, error) {
	res := &pb.ReadResponse{Date: make([]byte, size)}
	readReq := new(pb.ReadRequest)
	readReq.Id = int64(sessionId)
	readReq.Offset = offset

	readReq.Size = int64(size)
	res, err := c.cli.ReadAt(context.Background(), readReq)
	if err != nil {
		log.Println("in ReadAt", offset/BLOCK_SIZE, size)

		log.Println(err)
	}

	if res.EOF {
		err = io.EOF
	}

	if int64(size) != res.Size {
		return res.Date[:res.Size], err
	}

	return res.Date, nil
}

func (c *Client) CloseSession(sessionId SessionId) error {
	_, err := c.cli.Close(context.Background(), &pb.Request{Id: int64(sessionId)})
	return err
}

func (c *Client) Download(filename, saveFile string) error {
	return c.DownloadAt(filename, saveFile, 0)
}

func (c *Client) DownLoadBlock(filename, saveFile string, blockId int) error {
	file, err := os.OpenFile(saveFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	sessionId, err := c.Open(filename)
	if err != nil {
		return err
	}

	buf, rerr := c.GetBlock(sessionId, blockId)
	if rerr != nil && rerr != io.EOF {
		return rerr
	}
	if _, werr := file.WriteAt(buf, int64(blockId)*BLOCK_SIZE); werr != nil {
		return werr
	}
	log.Printf("Download %s %d completed", filename, blockId)

	c.CloseSession(sessionId)

	return nil

}

func (c *Client) UploadBlock(uploadSaveFile string) error {
	file, err := os.OpenFile("/home/li/"+uploadSaveFile, os.O_CREATE|os.O_RDONLY, 0666)
	stat, err := file.Stat()
	log.Printf("get file %s ,size if %v ", uploadSaveFile, stat.Size())
	size := stat.Size()
	blocks := int(size / BLOCK_SIZE)

	if size%BLOCK_SIZE != 0 {
		blocks += 1
	}

	if err != nil {
		return err
	}
	defer file.Close()

	sessionId, err := c.Open(uploadSaveFile + "_back")
	if err != nil {
		return err
	}

	buf := make([]byte, BLOCK_SIZE)

	for i := 0; i < blocks; i++ {
		n, rerr := file.ReadAt(buf, int64(i)*BLOCK_SIZE)
		if rerr != nil && rerr != io.EOF {
			return rerr
		}
		if _, werr := c.WriteAt(sessionId, buf, int64(i)*BLOCK_SIZE, int64(n)); werr != nil {
			log.Printf("writeAt get error[%v]", werr)
			return werr
		}

		//if i%((blocks-i)/100+1) == 0 {
		log.Printf("Uploading %s [%d/%d] blocks", uploadSaveFile, i, blocks)
		//}

		if rerr == io.EOF {
			break
		}
	}

	c.CloseSession(sessionId)

	return nil

}

func (c *Client) DownloadAt(filename, saveFile string, blockId int) error {
	stat, err := c.Stat(filename)
	if err != nil {
		log.Println("in DownloadAt :%s", err)
		log.Println(err)
		return err
	}
	log.Println("in DownLoad at")
	blocks := int(stat.Size / BLOCK_SIZE)
	if stat.Size%BLOCK_SIZE != 0 {
		blocks += 1
	}

	log.Printf("Download %s in %d blocks\n", filename, blocks-blockId)

	file, err := os.OpenFile(saveFile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	sessionId, err := c.Open(filename)
	if err != nil {
		return err
	}

	for i := blockId; i < blocks; i++ {
		buf, rerr := c.GetBlock(sessionId, i)
		if rerr != nil && rerr != io.EOF {
			return rerr
		}
		if _, werr := file.WriteAt(buf, int64(i)*BLOCK_SIZE); werr != nil {
			return werr
		}

		if i%((blocks-blockId)/100+1) == 0 {
			log.Printf("Downloading %s [%d/%d] blocks", filename, i-blockId+1, blocks-blockId)
		}

		if rerr == io.EOF {
			break
		}
	}
	log.Printf("Download %s completed", filename)

	c.CloseSession(sessionId)

	return nil
}
