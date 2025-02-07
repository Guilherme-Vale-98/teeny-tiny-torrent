package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackpal/bencode-go"
)
type Torrent struct {
  Announce string
  AnnounceList [][]string `bencode:"announce-list"` 
}
func main(){
  if(len(os.Args) < 2){
    fmt.Println("Usage go run main <torrentFile>")
    os.Exit(1)
  }
  torrentPath := os.Args[1]

  torrent, err := parseTorrentFile(torrentPath)
  if(err != nil){
    fmt.Println("Error parsing torrent err:" , err)
    os.Exit(1)
  }
  var trackersUrl []string
  for _ , v := range torrent.AnnounceList{
    for _, k := range v {
      trackersUrl = append(trackersUrl, k)
    }
  }

  parsedUrl, err := url.Parse(trackersUrl[3])
  fmt.Println(parsedUrl.Host)
  if(err != nil ){
    fmt.Println("Error parsing url:", err)
  }
  connectTracker(parsedUrl.Host)
}

func parseTorrentFile(torrentPath string) (*Torrent, error){

  torrentData, err := os.Open(torrentPath)
  if(err != nil){
    return nil, fmt.Errorf("Error reading file, err: %w", err)
  }
  defer torrentData.Close()
  var torrent Torrent
  err = bencode.Unmarshal(torrentData, &torrent)
  if(err != nil){
    return nil, fmt.Errorf("Error unmarshaling torrent err: %w", err)
  }
  return &torrent, nil  

}
func connectTracker(url string) {
  fmt.Printf("Connecting to tracker: %s \n", url)
  requestBuf, err := buildConnReq()
  if(err != nil){
    fmt.Println("error building connection")
  }
  udpAddr, err := net.ResolveUDPAddr("udp", url)
  if err != nil {
    fmt.Println("Error resolving udp address: ", err)
    os.Exit(1)
	}

  conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("failed to dial UDP: %w", err)
    os.Exit(1)
	}
	defer conn.Close()

	_, err = conn.Write(requestBuf)
  if err != nil {
    fmt.Println("failed to write buffer", err)
  }


  conn.SetDeadline(time.Now().Add(10 * time.Second))

  responseBuffer := make([]byte,16)

  _,_, err = conn.ReadFromUDP(responseBuffer)
  if err != nil {
    fmt.Println("Error reading from responseBuffer")
    os.Exit(1)
  }

  fmt.Println(responseBuffer)

}


func buildConnReq() ([]byte, error) {
  buf := make([]byte, 16)
 
	binary.BigEndian.PutUint32(buf[0:4], 0x417)
	binary.BigEndian.PutUint32(buf[4:8], 0x27101980)

	binary.BigEndian.PutUint32(buf[8:12], 0)

	if _, err := rand.Read(buf[12:16]); err != nil {
		return nil, err
	}

	return buf, nil

}
