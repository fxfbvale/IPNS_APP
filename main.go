package main

import (
	"context"
	"fmt"
	api "github.com/ipfs/go-ipfs-api"
	"os"
	"strconv"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {

	keys := []string{}

	// Create a shell to the local node
	sh := api.NewLocalShell()

	ctx := context.Background()

	machine := os.Args[1]
	nrMachines, _ := strconv.Atoi(os.Args[2])

	sleep := time.Minute
	switch os.Args[3] {
	case "5m":
		sleep = 5 * time.Minute
	case "15m":
		sleep = 15 * time.Minute
	case "30m":
		sleep = 30 * time.Minute
	case "1h":
		sleep = time.Hour
	case "2h":
		sleep = 2 * time.Hour
	case "3h":
		sleep = 3 * time.Hour
	case "6h":
		sleep = 6 * time.Hour
	case "12h":
		sleep = 12 * time.Hour
	default:
		os.Exit(1)
	}

	for i := 1; i <= nrMachines; i++ {
		fileName := fmt.Sprintf("pre%d.key", i)
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
		}

		err = sh.KeyImport(ctx, strconv.Itoa(i), file)
		if err != nil {
			fmt.Println(err)
		}
	}

	keyList, _ := sh.KeyList(ctx)
	fmt.Println(keyList)

	for _, key := range keyList {
		if key.Name != machine && key.Name != "self" {
			keys = append(keys, key.Id)
			sh.KeyRm(ctx, key.Name)
			fmt.Println("Removed key: ", key.Name)
		}
	}

	fmt.Println(keys)

	wg.Add(2)

	go allResolves(sh, keys)

	go publish(sh, sleep, machine)

	wg.Wait()

}

func publish(sh *api.Shell, sleep time.Duration, machine string) {
	counter := 0
	for {
		file, err := os.OpenFile("file.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		//writes the version number to the file so the hash changes
		counter++
		s := fmt.Sprintf("%s Version: %d \n", time.Now(), counter)
		_, err = file.WriteString(s)
		if err != nil {
			fmt.Println(err)
		}
		file.Close()

		file, err = os.OpenFile("file.txt", os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0666)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		cid, err := sh.Add(file)
		if err != nil {
			fmt.Printf("error adding file: %s", err)
		}

		file.Close()

		lifetime, err := time.ParseDuration("24h")
		if err != nil {
			fmt.Errorf("failed to parse lifetime: %w", err)
		}

		//so we dont have cache
		ttl, err := time.ParseDuration("1ns")
		if err != nil {
			fmt.Errorf("failed to parse ttl: %w", err)
		}

		// Publish the IPNS record using the default keypair
		ipnsEntry, err := sh.PublishWithDetails(cid, machine, lifetime, ttl, false)
		if err != nil {
			fmt.Errorf("failed to publish IPNS record: %w", err)
		}

		fmt.Println("Published", ipnsEntry)

		//waits for the next republish
		time.Sleep(sleep)

		//After the time, unpin the file so that it can be garbage collected
		sh.Unpin(cid)

	}

}

func resolve(sh *api.Shell, key string) {
	for {
		go func() {
			// Resolve the IPNS record to a valid IPFS path
			ipfsPath, err := sh.ResolvePath().Resolve(key)

			if err != nil {
				fmt.Errorf("failed to resolve IPNS record: %s", ipfsPath)
			}

		}()
		//waits 30 seconds to make each resolve
		time.Sleep(30 * time.Second)
	}
}

func allResolves(sh *api.Shell, keys []string) {
	for _, key := range keys {
		go resolve(sh, key)
	}
}
