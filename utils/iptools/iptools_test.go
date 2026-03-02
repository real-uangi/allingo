/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 16:26
 */

// Package iptools
package iptools

import "testing"

func TestIptools(t *testing.T) {

	la, err := GetLocalIP()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(la.GetV4().Strings())
	t.Log(la.GetV6().Strings())

	sps, err := la.GetIPByCIDR("192.0.0.0/8")
	if err != nil {
		t.Log(err)
	}
	t.Log(sps)

	t.Log(GetBestMatchIp())

}
