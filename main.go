package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type User struct {
	conn net.Conn
	name string
}

func listenForNewConnections() <-chan net.Conn {
	newconns := make(chan net.Conn)
	go func() {
		listener, err := net.Listen("tcp", ":8888")
		if err != nil {
			panic(err)
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			newconns <- conn
		}
	}()

	return newconns
}

func listenForInputFromUser(user User, lines chan<- string, unregisterCh chan<- net.Conn) {
	reader := bufio.NewReader(user.conn)
	for {
		user.conn.Write([]byte("> "))
		line, err := reader.ReadString('\n')

		if err == io.EOF || strings.EqualFold("\\quit\n", line) {
			fmt.Println("user being de-registered ", user.name)
			unregisterCh <- user.conn
			return
		} else if err != nil {
			fmt.Println("user had err ", err)
			panic(err)
		}

		cleanline := strings.TrimSpace(line)

		if len(cleanline) > 0 {
			lines <- fmt.Sprintf("[%s]: %s\n", user.name, cleanline)
		}
	}

}

func sendToUser(user User, line string) {
	user.conn.Write([]byte(line))
}

func getName(conn net.Conn) (string, error) {
	conn.Write([]byte("Enter your name: "))
	name, err := bufio.NewReader(conn).ReadString('\n')
	return strings.TrimSpace(name), err

}

func registerUser(conn net.Conn, newusers chan<- User) {
	name, err := getName(conn)
	if err != nil {
		if err != io.EOF {
			log.Fatal(err)
		}
	} else {
		newusers <- User{conn, name}
	}
}

func beginChatLoop(newconns <-chan net.Conn) {

	users := []User{}
	lines := make(chan string)
	newusers := make(chan User)
	unregisterCh := make(chan net.Conn)

	for
	{
		select {
		case conn := <-newconns:
			{
				go registerUser(conn, newusers)
			}
		case line := <-lines:
			{
				for _, user := range users {
					go sendToUser(user, line)
				}
			}
		case user := <-newusers:
			{
				users = append(users, user)
				go func() {
				   lines <- fmt.Sprintf("%s has joined the chat.\n", user.name) 
				}()
				go listenForInputFromUser(user, lines, unregisterCh)
			}
		case c := <-unregisterCh:
			{
				for idx, user := range users {
					if user.conn == c {
						users[len(users)-1], users[idx] = users[idx], users[len(users)-1]
						users = users[:len(users)-1]
						go func() {
							lines <- fmt.Sprintf("%s has left the chat.\n", user.name)
						}()
						break
					}
				}
			}
		}
	}

}

func main() {
	newconns := listenForNewConnections()

	beginChatLoop(newconns)
}
