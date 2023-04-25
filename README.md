# golldb

Module to interact with remote lldb-server/gdb-server.

```bash
$ debugserver '127.0.0.1:1234'
debugserver-@(#)PROGRAM:LLDB  PROJECT:lldb-1403.0.17.64
 for arm64.
Listening to port 1234 for a connection from 127.0.0.1...
Got a connection, waiting for process information for launching or attaching.
Attach succeeded, ready to debug.
Exiting.
```

```go
package main

import (
	"fmt"
	"github.com/nsecho/golldb"
)

func main() {
	lldb, err := golldb.NewLLDBServer("127.0.0.1", "1234")
	if err != nil {
		panic(err)
	}

	lldb.Attach("creator")
	fmt.Println(lldb.GetThreads())
	count := 16
	addr, _ := lldb.Allocate(count, "rw")
	fmt.Printf("Allocated %d bytes at address", addr)
	lldb.WriteAtAddress(addr, []byte{0xde, 0xad, 0xbe, 0xef})
	lldb.Continue()
}

```