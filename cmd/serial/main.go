package main

import (
	"fmt"
	"io"
	"log"

	// "time"
	"go.bug.st/serial"
	"net"
)

func listenTCP() net.Listener {
  tcp, tcpErr := net.Listen("tcp", ":9000")
  if tcpErr != nil {
    log.Fatal(tcpErr)
  }
  fmt.Println("Listening on", tcp.Addr())

  return tcp
}

func readFromTCP(conn net.Conn) []byte {
  val, readErr := io.ReadAll(conn)
  if readErr != nil {
    log.Fatal(readErr)
  }
  return val 
}

func writeToArduino(usbPort serial.Port, msg string) {
  n, err := usbPort.Write([]byte("<" + msg + ">"))
  if err != nil {
    log.Fatal(err)
  }
  fmt.Printf("Sent %v bytes\n", n) 
  
  buff := make([]byte, 100)
  n2, n2Err := usbPort.Read(buff)
  if n2Err != nil {
    log.Fatal(n2Err)
  }
  if n2 == 0 {
    fmt.Println("\nEOF")
  }
}

func main() {
  // listen for TCP
  tcp := listenTCP()
  defer tcp.Close()

  // connect to arduino
  var arduinoReady bool = false
  usbPort, portErr := serial.Open("/dev/cu.usbmodem14401", &serial.Mode{})
  if portErr != nil {
    log.Fatal(portErr)
  } 

  // read on uspPort until arduino is ready 
  buff := make([]byte, 100)
  var ardMsg string 
  for !arduinoReady {
    n, serialErr := usbPort.Read(buff)
    if serialErr != nil {
      log.Fatal(serialErr)
      break
    }
    if n == 0 {
      fmt.Println("\nEOF")
      break
    }

    for i := 0; i < n; i++ {
      byteStr := string(buff[i])
      if byteStr == "<" { continue } 
      if byteStr == ">" { break }
      ardMsg = ardMsg + byteStr 
    }
    arduinoReady = ardMsg == "Arduino is ready"
  }
  fmt.Println(ardMsg)
  
  for {
   conn, connErr := tcp.Accept()
    if connErr != nil {
      log.Fatal(connErr)
    }
    m := readFromTCP(conn)  
    
    writeToArduino(usbPort, string(m))

    conn.Close()
  }

 
 
}

 //    n, serialErr := usbPort.Read(buff)
 //    if serialErr != nil {
 //      log.Fatal(serialErr)
 //      break
 //    }
 //    if n == 0 {
 //      fmt.Println("\nEOF")
 //      break
 //    }
 //    fmt.Printf("%v", string(buff[:n]))
 //  }
 //  // stdin, err := io.ReadAll(os.Stdin)
 //  // if err != nil {
 //  //   log.Fatalln(err)
 //  // }
 //  var msg []byte = []byte("<24y>")
	//
 //  // time.Sleep(10 * time.Second)
	//
 //  n, err2 := usbPort.Write(msg)
 //  if err2 != nil {
	// 	log.Fatal(err2)
	// }
 //  fmt.Printf("Sent %v bytes\n", n)
	//
  

  // ports, err := serial.GetPortsList()
  // if err != nil {
  //   log.Fatal(err)
  // }
  // if len(ports) == 0 {
  //   log.Fatal("No serial ports found!")
  // }
  // for _, port := range ports {
  //   fmt.Printf("Found port: %v\n", port)
  // }
// if err != nil {
  //   log.Fatal(err)
  // }
  // 
  // usbPort, err1 := serial.Open(ports[5], &serial.Mode{})
  // if err1 != nil {
  //   log.Fatal(err1)
  // }
