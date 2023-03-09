package gone

import (
	"fmt"
	"os"
)

func build() {
	fmt.Println(os.Executable())
}
