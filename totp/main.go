// go-toybox/totp
// MIT License Copyright(c) 2021 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package main

import (
    "crypto/hmac"
    "crypto/sha1"
    "encoding/base32"
    "encoding/binary"
    "fmt"
    "os"
    "time"
)

func HOTP(k []byte, c uint64) int {
    code := make([]byte, 8)
    binary.BigEndian.PutUint64(code, c)
    m := hmac.New(sha1.New, k)
    m.Write(code)
    h := m.Sum(nil)
    off := h[19] & 0xf
    p := h[off:off + 4]
    return (int(binary.BigEndian.Uint32(p)) & 0x7fffffff) % 1000000
}

func hotptest() {
    k := []byte("12345678901234567890")
    ex := []int{755224, 287082, 359152, 969429, 338314, 254676, 287922, 162583}
    for c, v := range ex {
	if val := HOTP(k, uint64(c)); val != v {
	    fmt.Printf("bad counter=%d returns %d expected %d\n", c, val, v)
	}
    }
}

func TOTP(k []byte) int {
    t0 := int64(0)
    x := uint64(30)
    c := uint64(time.Now().Unix() - t0) / x
    return HOTP(k, c)
}

func main() {
    if len(os.Args) < 2 {
	fmt.Println("Usage: totp <base32 encoded secret>")
	return
    }
    k, _ := base32.StdEncoding.DecodeString(os.Args[1])
    //hotptest()
    fmt.Printf("%06d\n", TOTP(k))
}
