package collector

type UDPv1 struct {
}

func (v1 UDPv1) HandleUDPpacket(buf []byte, addr string) {
	statsUDPv1()
}

func (v1 UDPv1) Stop() {

}
