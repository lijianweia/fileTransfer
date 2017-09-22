package transfer

import "testing"

func Test_client(t *testing.T)  {
	c := NewClient(addressFileC)
	c.Dial()
	c.Download("google-chrome-stable_current_amd64.deb","google-chrome-stable_current_amd64.deb.1")
}

