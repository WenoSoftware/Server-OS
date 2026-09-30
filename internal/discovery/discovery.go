package discovery

import (
	"fmt"
	"net"
	"time"

	"serveros/internal/network"
)

const DiscoveryPort = 9999
const BroadcastMessage = "WENO_NODE_DISCOVER"

// StartBroadcaster periodically announces this node's presence to the local network
func StartBroadcaster(daemonPort string) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", DiscoveryPort))
	if err != nil {
		fmt.Printf("[DISCOVERY] Error resolving broadcast address: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("[DISCOVERY] Error creating broadcast connection: %v\n", err)
		return
	}
	defer conn.Close()

	payload := []byte(fmt.Sprintf("WENO_ANNOUNCE:%s", daemonPort))

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		_, err := conn.Write(payload)
		if err != nil {
			continue
		}
	}
}

// StartListener listens for broadcast announcements from other nodes on the local network
func StartListener(myDaemonPort string) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		fmt.Printf("[DISCOVERY] Error resolving listener address: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("[DISCOVERY] Error listening for UDP broadcasts: %v\n", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 512)
	for {
		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}

		msg := string(buf[:n])
		if len(msg) > 13 && msg[:13] == "WENO_ANNOUNCE:" {
			peerPort := msg[13:]

			// Type assert net.Addr to *net.UDPAddr to access the IP field safely
			udpAddr, ok := remoteAddr.(*net.UDPAddr)
			if !ok {
				continue
			}

			peerIP := udpAddr.IP.String()
			peerAddress := fmt.Sprintf("%s:%s", peerIP, peerPort)

			// Don't add ourselves
			localTarget := fmt.Sprintf("127.0.0.1:%s", myDaemonPort)
			if peerAddress != localTarget && peerAddress != fmt.Sprintf("localhost:%s", myDaemonPort) {
				network.SwarmPeers.AddPeer(peerAddress)
			}
		}
	}
}