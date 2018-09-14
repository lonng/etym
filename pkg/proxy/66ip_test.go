package proxy

import (
	"fmt"
	"testing"
)

func TestIp66_Fetch(t *testing.T) {
	p := ip66{}
	x, _, err := p.Fetch(1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", x)
}

func TestXicidaili_Fetch(t *testing.T) {
	p := xicidaili{baseUrl: "http://www.xicidaili.com/nn"}
	x, _, err := p.Fetch(2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", x)
}

func TestU5_Fetch(t *testing.T) {
	p := u5{}
	x, _, err := p.Fetch(1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", x)
}

func TestPlp_Fetch(t *testing.T) {
	p := plp{}
	x, page, err := p.Fetch(1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("page:%d, %+v\n", page, x)
}
