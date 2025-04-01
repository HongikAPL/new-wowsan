package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/igeon510/new-wowsan/pkg/broker"
	"github.com/igeon510/new-wowsan/pkg/proto"
)

// 커맨드라인에서 addbroker만 대충 실험하기 위한 코드를 짰습니다.
// 잘 작동합니다. 다른 기능들을 실험하기 위해 main.go에서 받아 브로커 유즈케이스로 전달하고, 각 기능을 테스트해야 합니다.
// 수정 요망.

func main() {
	// 커맨드라인 인자로 IP와 포트 받기
	ip := flag.String("ip", "127.0.0.1", "IP address of the broker")
	port := flag.String("port", "50051", "Port for the broker to listen on")
	flag.Parse()

	// NodeInfo 생성
	nodeInfo := &proto.NodeInfo{
		Id:   *ip + ":" + *port,
		Ip:   *ip,
		Port: *port,
	}

	// Broker 생성
	b := broker.NewBroker(nodeInfo)

	// 브로커 서버 시작
	go func() {
		broker.StartBrokerServer(b)
	}()

	// 명령줄 입력 처리
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.HasPrefix(input, "add ") {
			parts := strings.Split(input, " ")
			if len(parts) == 3 {
				peerIp := parts[1]
				peerPort := parts[2]
				peerNodeInfo := &proto.NodeInfo{
					Id:   peerIp + ":" + peerPort,
					Ip:   peerIp,
					Port: peerPort,
				}

				// 중복 확인
				if _, exists := b.Peers[peerNodeInfo.Id]; exists {
					log.Printf("Peer %s:%s already exists", peerIp, peerPort)
					continue
				}

				// BrokerClient 생성
				client, err := broker.NewBrokerClient(peerNodeInfo)
				if err != nil {
					log.Printf("Failed to create client for peer %s:%s: %v", peerIp, peerPort, err)
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// 원격 브로커에 자신 추가 요청
				req := &proto.AddBrokerRequest{Node: nodeInfo}
				_, err = client.AddBroker(ctx, req)
				if err != nil {
					log.Printf("Failed to add broker to peer %s:%s: %v", peerIp, peerPort, err)
					client.Close()
					continue
				}

				// 로컬 브로커 상태 업데이트
				b.AddBroker(peerNodeInfo)
				log.Printf("Successfully added broker to peer %s:%s", peerIp, peerPort)

				client.Close()
			} else {
				log.Println("Invalid command format. Use: add <ip> <port>")
			}
		} else if input == "peer" {
			b.PrintPeers()
		} else {
			log.Println("Unknown command.")
		}
	}
}
