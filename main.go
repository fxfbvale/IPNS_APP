package main

import (
	"fmt"
	api "github.com/ipfs/go-ipfs-api"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	keys := []string{"k51qzi5uqu5dgygf5hcjp8m7jyj0pkyhms2xv31x5rfy910gnppdai8boegm7l"}

	// Create a shell to the local node
	sh := api.NewLocalShell()

	// Obtain the local IPFS node's keypair ID (default is "self")
	//keyID := "self"

	//ctx := context.Background()

	wg.Add(2)

	go allResolves(sh, keys)

	go publish(sh)

	wg.Wait()

	/**
		//fazer um for para cada chave dá resolve e fazer um for tmb para o publish
		file, err := os.OpenFile("/Users/vale/hello.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		cid, err := sh.Add(bufio.NewReader(file))
		if err != nil {
			fmt.Printf("error adding file: %s", err)
		}

		lifetime, err := time.ParseDuration("24h")
		if err != nil {
			fmt.Errorf("failed to parse lifetime: %w", err)
		}
		ttl, err := time.ParseDuration("1ns")

		//key2, err := sh.KeyGen(ctx, "1")

		// Publish the IPNS record using the default keypair
		ipnsEntry, err := sh.PublishWithDetails(cid, "2", lifetime, ttl, false)
		if err != nil {
			fmt.Errorf("failed to publish IPNS record: %w", err)
		}

		fmt.Println("Published", ipnsEntry.Name)

		sh.Resolve("k51qzi5uqu5dk2z33rmgr0ynsntwsoyudfge15ej7u9y3a67gcmwi9ksikbugm")
		//sh.Unpin(cid)
	**/
}

func publish(sh *api.Shell) {
	counter := 0
	for {
		file, err := os.OpenFile("test.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
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

		file, err = os.OpenFile("test.txt", os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0666)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		cid, err := sh.Add(file)
		if err != nil {
			fmt.Printf("error adding file: %s", err)
		}

		lifetime, err := time.ParseDuration("24h")
		if err != nil {
			fmt.Errorf("failed to parse lifetime: %w", err)
		}

		//so we dont have cache
		ttl, err := time.ParseDuration("1ns")

		// Publish the IPNS record using the default keypair
		ipnsEntry, err := sh.PublishWithDetails(cid, "self", lifetime, ttl, false)
		if err != nil {
			fmt.Errorf("failed to publish IPNS record: %w", err)
		}

		fmt.Println("Published", ipnsEntry.Name)

		//waits for the next republish
		time.Sleep(3 * time.Minute)

		//After the time, unpin the file so that it can be garbage collected
		sh.Unpin(cid)
	}

}

func resolve(sh *api.Shell, key string) {
	for {

		go func() {
			// Resolve the IPNS record to a valid IPFS path
			ipfsPath, err := sh.Resolve(key)

			if err != nil {
				fmt.Errorf("failed to resolve IPNS record: %w, %", ipfsPath)
			}

			//waits 30 seconds to make each resolve
			time.Sleep(30 * time.Second)
		}()

	}
}

func allResolves(sh *api.Shell, keys []string) {
	for _, key := range keys {
		go resolve(sh, key)
	}
}
