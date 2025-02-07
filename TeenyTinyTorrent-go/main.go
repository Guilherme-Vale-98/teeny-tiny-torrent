package main

import(
  "fmt"
  "os"
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
      fmt.Println(k)
    }
  }


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
