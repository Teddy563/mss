package conn

import (
	"fmt"
	"testing"
	"time"

	"github.com/dreamscached/minequery/v2"

	"msh/lib/config"
)

func Test_QueryFull(t *testing.T) {
	config.ProxyHost, config.ProxyPortQuery = "127.0.0.1", 25555

	go HandlerQuery()

	minequery.WithUseStrict(true)

	for i := 0; i < 3; i++ {
		fmt.Println("--------------------")

		res, err := minequery.QueryFull(config.ProxyHost, config.ProxyPortQuery)
		if err != nil {
			t.Fatalf(err.Error())
		}

		fmt.Printf("result: %+v\n", res)

		time.Sleep(time.Second)
	}
}

func Test_QueryBasic(t *testing.T) {
	config.ProxyHost, config.ProxyPortQuery = "127.0.0.1", 25555

	go HandlerQuery()

	minequery.WithUseStrict(true)

	for i := 0; i < 3; i++ {
		fmt.Println("--------------------")

		res, err := minequery.QueryBasic(config.ProxyHost, config.ProxyPortQuery)
		if err != nil {
			t.Fatalf(err.Error())
		}

		fmt.Printf("result: %+v\n", res)

		time.Sleep(time.Second)
	}
}
