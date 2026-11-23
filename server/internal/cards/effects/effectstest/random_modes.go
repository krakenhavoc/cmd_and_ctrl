package effectstest

func RandomModesKey(n byte) [32]byte {
	var k [32]byte
	for i := range k {
		k[i] = byte(i+1) ^ n
	}
	return k
}
