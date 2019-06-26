package main

import (
	"encoding/hex"

	"flag"
	"fmt"

	"log"
	"net"
	"strconv"
	"strings"
)

// func ReadUint8(reader io.Reader) (val uint8, err error) {
// 	var util [1]byte
// 	_, err = reader.Read(util[:1])
// 	val = util[0]
// 	return
// }

//protocol version 18 00 d4 02 11

// //read a varint type
// func ReadVarInt(reader io.Reader) (result int, err error) {
// 	var bytes byte = 0
// 	var b byte
// 	for {
// 		b, err = ReadUint8(reader)
// 		if err != nil {
// 			return
// 		}
// 		result |= int(uint(b&0x7F) << uint(bytes*7))
// 		bytes++
// 		if bytes > 5 {
// 			err = errors.New("Decode, VarInt is too long")
// 			return
// 		}
// 		//AND if is 128
// 		if (b & 0x80) == 0x80 {
// 			continue
// 		}
// 		break
// 	}
// 	return
// }

func ReadVarIntBytes(varint []byte) (result int) {
	var bytes byte

	for i := 0; i < len(varint); i++ {
		b := varint[i]
		//bitwise inclusive OR and assignment operator
		//	C |= 2 is same as C = C | 2
		//https://homerl.github.io/2016/03/29/golang-bitwise-operators/

		result |= int(uint(b&0x7F) << uint(bytes*7))
		//fmt.Println(result)//20 404
		bytes++
		fmt.Println(strconv.FormatInt(int64(0x80), 2))

		//10000000 most significant byte is set
		if (b & 0x80) == 0x80 {
			continue
			//there are more bytes to come
		} else {
			break //else break
		}
	}
	return
}

var (
	// addresses
	localAddr  = flag.String("lhost", ":4444", "proxy local address")
	targetAddr = flag.String("rhost", "panel.rapzprison.net:25565", "proxy remote address")
)

func modifyresp(b *[]byte) {

	if strings.Contains(string(*b), "Nieuwe") {
		log.Printf("nieuwe")
	}
	// strings.Replace(x, "400", "20", 1)
	// log.Printf("%T", x)
	// log.Printf(x)
	// *b = []byte(x)
}
func main() {
	flag.Parse()
	//1800940311
	// 18 00 94 03 11
	// 1800d40211 -> 24

	//	protocolversion, _ := hex.DecodeString("940311")
	//	fmt.Sprintf(%08b, protocolversion[0])
	//	fmt.Println("protocol version: " + strconv.Itoa(ReadVarIntBytes(protocolversion)))

	fmt.Println(strings.Replace("5 cookies", "5", "2", -1))

	//18 > 24 length
	// if err != nil {
	// 	panic(err)
	// } else {
	// //	protocolversion, err := ReadVarInt(bytes.NewBuffer(protocolversion))
	// 	if err != nil {
	// 		panic(err)
	// 	} else {
	// 	//	fmt.Println(protocolversion)

	// 	}
	// }
	p := Server{
		Addr:           *localAddr,
		Target:         *targetAddr,
		ModifyResponse: modifyresp,
	}

	log.Println("Proxying from " + p.Addr + " to " + p.Target)

	p.ListenAndServe()
}

// Server is a TCP server that takes an incoming request and sends it to another
// server, proxying the response back to the client.
type Server struct {
	// TCP address to listen on
	Addr string

	// TCP address of target server
	Target string

	// ModifyRequest is an optional function that modifies the request from a client to the target server.
	ModifyRequest func(b *[]byte)

	// ModifyResponse is an optional function that modifies the response from the target server.
	ModifyResponse func(b *[]byte)
}

// ListenAndServe listens on the TCP network address laddr and then handle packets
// on incoming connections.
func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)
	log.Printf("listening")
	if err != nil {
		return err
	}
	return s.serve(listener)
}

func (s *Server) serve(ln net.Listener) error {
	for {
		client, err := ln.Accept()
		log.Println("new conn")
		if err != nil {
			log.Println(err)
			continue
		}
		go s.handleConn(client)
	}
}

func (s *Server) handleConn(client net.Conn) {
	// connects to target server

	log.Println("conn")
	var server net.Conn
	var err error

	server, err = net.Dial("tcp", s.Target)

	if err != nil {
		return
	}

	// write to dst what it reads from src
	var pipe = func(src, dst net.Conn, typed string, filter func(b *[]byte)) {
		defer func() {
			client.Close()
			server.Close()
		}()

		buff := make([]byte, 65535)
		for {

			n, err := src.Read(buff)
			if err != nil {
				log.Println(err)
				return

			}
			b := buff[:n]
			//	fmt.Printf(hex.Dump(b[1:2]))
			nullbyte, _ := hex.DecodeString("00")
			if b[1] == nullbyte[0] {
				// len packet 94 03 11 73 6f 63 6b  65 74 2e 66 65 72 6f 78
				fmt.Println("login packet")
				var bytes []byte
				var bitindex = 2
				for {

					bytes = append(bytes, b[bitindex])
					
					if (b[bitindex] & 0x80) == 0x80 {
						fmt.Println("more bytes")

					} else {

						fmt.Println(hex.Dump(bytes))

						//		protocolversion, _ := hex.DecodeString("940311")
						//	fmt.Sprintf(%08b, protocolversion[0])
						fmt.Println("protocol version: " + strconv.Itoa(ReadVarIntBytes(bytes)))

						break
					}
					bitindex++

				}

				//	firstpacket := b[0:b[0]]
				//	fmt.Println(hex.Dump(firstpacket))
			}
			//log.Printf("received:\n%v", hex.Dump(b))
			//	log.Printf(string(b))

			if filter != nil {
				//modify the response
				filter(&b)
			}
			//fmt.Printf(typed+":\n%v", hex.Dump(b))
			_, err = dst.Write(b)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	go pipe(client, server, "C -> S", s.ModifyRequest)
	go pipe(server, client, "S -> C", s.ModifyResponse)
}
