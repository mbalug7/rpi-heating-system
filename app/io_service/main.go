package main

import pb "rpi-heating-system/lib/protobuf/io_service/pb" // Import the generated code

func main() {

	pb.RegisterIOServiceServer(s, &server{})
}
