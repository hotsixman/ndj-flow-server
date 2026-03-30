package server

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"sync"
)

type Server struct {
	Listener         net.Listener
	Client           map[string]net.Conn
	ClientWriteMutex map[string]*sync.Mutex
	ClientMapMutex   *sync.Mutex
	Listening        bool
	KeyChecker       KeyChecker
}

type ClientCheckData struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

func (this *Server) Listen() {
	this.Listening = true
	go func() {
		for {
			conn, err := this.Listener.Accept()
			if err != nil {
				log.Println("Error accepting:", err)
				continue
			}
			go handleRequest(conn, this)
		}
	}()
}

func CreateServer(port string, keyChecker KeyChecker) (*Server, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, err
	}

	server := &Server{
		Listener:         listener,
		Client:           make(map[string]net.Conn),
		ClientWriteMutex: make(map[string]*sync.Mutex),
		ClientMapMutex:   &sync.Mutex{},
		Listening:        false,
		KeyChecker:       keyChecker,
	}

	return server, nil
}

func handleRequest(conn net.Conn, server *Server) {
	name := ""
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CRITICAL] Panic recovered in handleRequest for %s: %v", name, r)
		}
	}()
	defer func() {
		server.ClientMapMutex.Lock()
		delete(server.Client, name)
		delete(server.ClientWriteMutex, name)
		conn.Close()
		server.ClientMapMutex.Unlock()
	}()
	reader := bufio.NewReader(conn)

	// check client
	name, err := checkClient(reader, server.KeyChecker)
	if err != nil {
		log.Println("[Error] Error checking client\n", err)
		return
	}
	server.ClientMapMutex.Lock()
	if _, exists := server.Client[name]; exists {
		log.Println("[Error]", name, "already exists.")
		server.ClientMapMutex.Unlock()
		return
	}
	server.Client[name] = conn
	server.ClientWriteMutex[name] = &sync.Mutex{}
	server.ClientMapMutex.Unlock()

	// read
	for {
		alive := readMessage(reader, name, server)
		if !alive {
			return
		}
	}
}

func checkClient(reader *bufio.Reader, keyChecker KeyChecker) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	clientCheckData := ClientCheckData{}
	err = json.Unmarshal([]byte(line), &clientCheckData)
	if err != nil {
		return "", err
	}

	check := keyChecker.Check(clientCheckData.Name, clientCheckData.Key)
	if !check {
		return "", &ServerException{code: "INVALID_NAME_OR_KEY"}
	}

	return clientCheckData.Name, nil
}

func readMessage(reader *bufio.Reader, name string, server *Server) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CRITICAL] Panic recovered in handleRequest for %s: %v", name, r)
		}
	}()
	header, err := readHeader(reader)
	validMessage := (err == nil)
	var writeMutex *sync.Mutex = nil
	var writer *bufio.Writer = nil
	if validMessage {
		server.ClientMapMutex.Lock()
		destination := server.Client[header["destination"]]
		writeMutex = server.ClientWriteMutex[header["destination"]]
		if writeMutex != nil {
			writeMutex.Lock()
		}
		defer func() {
			if writeMutex != nil {
				writeMutex.Unlock()
			}
		}()
		if destination != nil {
			writer = bufio.NewWriter(destination)
		}
		server.ClientMapMutex.Unlock()
	} else {
		log.Println("[Error] Reading header from", name, ":\n", err)
		return false
	}

	if writer != nil {
		header["from"] = name
		headerJSON, err := json.Marshal(header)
		if err == nil {
			_, err = writer.WriteString(string(headerJSON) + "\n")
			if err == nil {
				writer.Flush()
			} else {
				writer = nil
			}
		} else {
			writer = nil
		}
	}

	endFlag := 0
	for {
		b, err := reader.ReadByte()
		if err != nil {
			log.Println("[Error] Error reading stream", err)
			return false
		}

		if b == 'e' {
			endFlag = 1
		} else if b == 'n' && endFlag == 1 {
			endFlag++
		} else if b == 'd' && endFlag == 2 {
			endFlag++
		} else if b == '\n' && endFlag == 3 {
			endFlag++
		} else {
			endFlag = 0
		}

		if writer != nil {
			err := writer.WriteByte(b)
			if err != nil {
				log.Println("[Error] Error sending message: ", err)
				writer = nil
			}
		}

		if endFlag == 4 {
			if writer != nil {
				writer.Flush()
			}
			break
		}
	}
	return true
}

func readHeader(reader *bufio.Reader) (map[string]string, error) {
	headerJSON, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	headerJSON = headerJSON[:len(headerJSON)-1]

	header := make(map[string]string)
	err = json.Unmarshal([]byte(headerJSON), &header)
	if err != nil {
		return nil, err
	}

	if header["destination"] != "" && header["id"] != "" {
		return header, nil
	} else {
		return nil, &ServerException{code: "NO_DESTINATION_OR_ID_IN_HEADER"}
	}
}
