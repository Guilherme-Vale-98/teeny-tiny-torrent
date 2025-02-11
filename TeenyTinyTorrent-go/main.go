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

type TorrentInfo struct {
	Name        string          `bencode:"name"`
	PieceLength int64           `bencode:"piece length"`
	Pieces      string          `bencode:"pieces"`
	Private     int64           `bencode:"private"`
	Files       []TorrentFile `bencode:"files"`
	Length      int64           `bencode:"length"`
}

type TorrentFile struct {
	Length int64    `bencode:"length"`
	Path   []string `bencode:"path"`
}


var bitTorrentPort int = 6881

type connResponse struct {
  Action  uint32
  TransactionId uint32
  ConnectionId uint64
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

  id, err := generateId()
  if err != nil {
    fmt.Println("Error generating id")
    os.Exit(1)
  }
  fmt.Println(string(id))

  getPeers(*torrent)
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

func getPeers(torrent Torrent) {

  var trackersUrl []string
  for _ , v := range torrent.AnnounceList{
    for _, k := range v {
      trackersUrl = append(trackersUrl, k)
    }
  }

  parsedUrl, err := url.Parse(trackersUrl[3])
  if(err != nil ){
    fmt.Println("Error parsing url:", err)
  }
  
  url := parsedUrl.Host
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
    fmt.Println("Error reading from responseBuffer: ", err)
    os.Exit(1)
  }

  parsedResponse, err := parseConnRes(responseBuffer)
  if err != nil {
    fmt.Println("Error parsing error")
  }


  if parsedResponse.TransactionId != binary.BigEndian.Uint32(requestBuf[12:16]){
    fmt.Printf("mismatch on transactionId: sent %d received: %d", binary.BigEndian.Uint32(requestBuf[12:16]),parsedResponse.TransactionId )
    os.Exit(1)
  }

  fmt.Println("Connection success")
  buildAnnounceRequest(torrent, parsedResponse.ConnectionId)

}


func buildConnReq() ([]byte, error) {
  buf := make([]byte, 16)
  //connectionId 
	binary.BigEndian.PutUint32(buf[0:4], 0x417)
	binary.BigEndian.PutUint32(buf[4:8], 0x27101980)
  //action
	binary.BigEndian.PutUint32(buf[8:12], 0)
  //TransactionId
	if _, err := rand.Read(buf[12:16]); err != nil {
		return nil, err
	}
	return buf, nil

}
func parseConnRes(resp []byte) (connResponse, error) {
    if len(resp) < 16 {
        return connResponse{}, fmt.Errorf("response too short: got %d bytes, need at least 16", len(resp))
    }
    var parsedResponse connResponse
    parsedResponse.Action = binary.BigEndian.Uint32(resp[0:4])
    parsedResponse.TransactionId = binary.BigEndian.Uint32(resp[4:8])
    parsedResponse.ConnectionId = binary.BigEndian.Uint64(resp[8:16])
    return parsedResponse, nil
}

func generateId() ([]byte, error) {
  buf := make([]byte, 20)
  
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
  copy(buf, []byte("-GV0001-"))

  return buf, nil
}

func buildAnnounceRequest(torrent Torrent, connId uint64) ([]byte,error){
  buf := make([]byte, 98) 
  binary.BigEndian.PutUint64(buf[0:8], connId)

  binary.BigEndian.PutUint32(buf[8:12], 1)

	if _, err := rand.Read(buf[12:16]); err != nil {
		return nil, err
	}



  fmt.Println("----**-----")
  fmt.Println(connId)
  fmt.Println(binary.BigEndian.Uint64(buf))
  fmt.Println(binary.BigEndian.Uint32(buf[8:12]))
  return buf, nil
}








