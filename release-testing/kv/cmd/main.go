package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"test/kv"
)

func main() {
	// Subcommands
	getCommand := flag.NewFlagSet("get", flag.ExitOnError)
	setCommand := flag.NewFlagSet("set", flag.ExitOnError)
	listCommand := flag.NewFlagSet("list", flag.ExitOnError)
	destroyCommand := flag.NewFlagSet("destroy", flag.ExitOnError)
	serveCommand := flag.NewFlagSet("serve", flag.ExitOnError)

	getKeyPtr := getCommand.String("k", "", "Key to get. (Required)")

	setKeyPtr := setCommand.String("k", "", "Key to set. (Required)")
	setValPtr := setCommand.String("v", "", "Value to set. (Required)")

	// Verify that a subcommand has been provided
	// os.Arg[0] is the main command
	// os.Arg[1] will be the subcommand
	if len(os.Args) < 2 {
		Fatal("subcommand is required")
	}

	// Switch on the subcommand
	// Parse the flags for appropriate FlagSet
	// FlagSet.Parse() requires a set of arguments to parse as input
	// os.Args[2:] will be all arguments starting after the subcommand at os.Args[1]
	switch os.Args[1] {
	case "list":
		listCommand.Parse(os.Args[2:])
	case "set":
		setCommand.Parse(os.Args[2:])
	case "get":
		getCommand.Parse(os.Args[2:])
	case "destroy":
		destroyCommand.Parse(os.Args[2:])
	case "serve":
		serveCommand.Parse(os.Args[2:])
	default:
		Fatal("unknown subcommand")
	}

	if serveCommand.Parsed() {
		runServer()
		return
	}

	client, err := kv.DefaultStoreClient()
	if err != nil {
		Fatal("unable to create store client: " + err.Error())
		return
	}

	if destroyCommand.Parsed() {
		err = client.Destroy()
		if err != nil && err != io.ErrUnexpectedEOF {
			Fatal(err.Error())
		}

		return
	}

	if listCommand.Parsed() {
		keys, err := client.List()
		if err != nil {
			Fatal(err.Error())
		}

		for _, key := range keys {
			fmt.Println(key)
		}
		return
	}

	if getCommand.Parsed() {
		// Required Flags
		if *getKeyPtr == "" {
			getCommand.PrintDefaults()
			os.Exit(1)
		}

		val, err := client.Get(*getKeyPtr)
		if err != nil {
			Fatal(err.Error())
		}

		fmt.Printf("%s", val)
		return
	}

	if setCommand.Parsed() {
		// Required Flags
		if *setKeyPtr == "" || *setValPtr == "" {
			setCommand.PrintDefaults()
			os.Exit(1)
		}

		if err != client.Set(*setKeyPtr, *setValPtr) {
			Fatal(err.Error())
		}

		return
	}
}

func Fatal(msg string) {
	_, _ = os.Stderr.Write([]byte("ERROR: " + msg))
	os.Exit(1)
}

func runServer()  {
	log.Println("#Serve")

	err := rpc.Register(kv.NewStore())
	if err != nil {
		log.Fatal("register error: " + err.Error())
	}

	rpc.HandleHTTP()
	l, err := net.Listen("tcp", "localhost:")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	log.Println("Using port:", l.Addr().(*net.TCPAddr).Port)

	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal("serve error:", err)
	}
}
