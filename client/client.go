package client

import "github.com/jack/duck-cc-client-http/httpclient"

func main() {
	client := httpclient.NewClient()
	client.Server()
}
