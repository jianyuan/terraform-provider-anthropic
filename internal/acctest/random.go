package acctest

import (
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func RandInt() int {
	return sdkacctest.RandInt()
}

func RandomWithPrefix(name string) string {
	return sdkacctest.RandomWithPrefix(name)
}
