package golldb

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type LLDB struct {
	conn   net.Conn
	target string
	*sync.Mutex
}

type Address struct {
	value uint64
}

func (a *Address) String() string {
	return fmt.Sprintf("%x", a.value)
}

// NewLLDBServer returns new instance of LLDB struct that is used to interact
// with remote gdbserver/lldb-server
func NewLLDBServer(ip, port string) (*LLDB, error) {
	address := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	buffer := make([]byte, 24)
	conn.Write([]byte("$QStartNoAckMode#b0"))
	conn.Read(buffer)
	conn.Write([]byte("+"))

	lldb := &LLDB{
		conn: conn,
	}

	return lldb, nil
}

// GetThreads returns information about the threads running.
func (l *LLDB) GetThreads() (map[string]any, error) {
	msg := "jThreadsInfo"
	repl := string(l.execSimple(msg))
	repl = strings.TrimLeft(repl, "$")
	repl = strings.TrimRight(repl, "#00")

	mp := make(map[string]any)
	json.Unmarshal([]byte(repl), &mp)
	return mp, nil
}

// Interrupt interrupts the running binary as if it has been sent CTRL+C
func (l *LLDB) Interrupt() error {
	l.execSimple("vCtrlC")
	return nil
}

// Close closes underlying connection to the gdbserver/lldb-server
func (l *LLDB) Close() error {
	if l.target != "" {
		if err := l.Detach(); err != nil {
			return err
		}
	}
	return l.conn.Close()
}

// SetStdout sets stdout for the target that we will create.
func (l *LLDB) SetStdout(path string) error {
	stdout := hex.EncodeToString([]byte(path))
	msg := fmt.Sprintf("QSetSTDOUT:%s", stdout)
	l.execSimple(msg)
	return nil
}

// SetStdin sets stdin for the target that we will create.
func (l *LLDB) SetStdin(path string) error {
	stdin := hex.EncodeToString([]byte(path))
	msg := fmt.Sprintf("QSetSTDIN:%s", stdin)
	l.execSimple(msg)
	return nil
}

// SetStderr sets stderr for the target that we will create.
func (l *LLDB) SetStderr(path string) error {
	stderr := hex.EncodeToString([]byte(path))
	msg := fmt.Sprintf("QSetSTDERR:%s", stderr)
	l.execSimple(msg)
	return nil
}

// SetEnv sets environment variables for the target that we will create.
func (l *LLDB) SetEnv(env map[string]string) error {
	return nil
}

// SetEnvEscaped sets environment variables for the target that we will create, useful when
// name or value contains characters such as "#"
func (l *LLDB) SetEnvEscaped(env map[string]string) error {
	return nil
}

// Create creates new target for debugging.
func (l *LLDB) Create(target string, argv ...string) error {
	msg := "A"
	encodedTarget := hex.EncodeToString([]byte(target))
	msg += strconv.Itoa(len(encodedTarget))
	msg += ",0,"
	msg += encodedTarget

	for idx, arg := range argv {
		msg += ","
		encoded := hex.EncodeToString([]byte(arg))
		msg += strconv.Itoa(len(encoded))
		msg += ","
		msg += strconv.Itoa(idx + 1)
		msg += ","
		msg += encoded
	}

	l.target = target
	l.execSimple(msg)
	return nil
}

// Run runs the target previously created.
func (l *LLDB) Run() error {
	if l.target == "" {
		return errors.New("cannot run; target is not created")
	}
	l.execSimple("c")
	return nil
}

// Continue continues the execution of the debugged target.
func (l *LLDB) Continue() error {
	if l.target == "" {
		return errors.New("cannot continue; target is not created or not attach to process")
	}
	l.execSimple("c")
	return nil
}

// Allocate will allocate size bytes with the permissions passed.
func (l *LLDB) Allocate(size int, permissions string) (*Address, error) {
	msg := "_M" + strconv.Itoa(size) + "," + permissions
	res := string(l.execSimple(msg))
	res = strings.Replace(res, "$", "", -1)
	res = strings.Replace(res, "#00", "", -1)
	addr, _ := strconv.ParseUint(res, 16, 64)
	return &Address{
		value: addr,
	}, nil
}

// WriteAtAddress writes specified byte slice at the address provided.
func (l *LLDB) WriteAtAddress(addr *Address, data []byte) error {
	encoded := hex.EncodeToString(data)
	msg := "M"
	msg += fmt.Sprintf("%x", addr.value)
	msg += ","
	msg += strconv.Itoa(len(encoded))
	msg += ":"
	msg += encoded
	l.execSimple(msg)
	return nil
}

// Attach attaches to the running program by name.
func (l *LLDB) Attach(name string) error {
	l.target = name
	msg := "vAttachName;" + hex.EncodeToString([]byte(name))
	l.execSimple(msg)
	return nil
}

// Detach detaches from the debugger program.
func (l *LLDB) Detach() error {
	l.execSimple("D")
	return nil
}

// SaveRegisters saves current snapshot of the registers.
func (l *LLDB) SaveRegisters() error {
	l.execSimple("QSaveRegisterState")
	return nil
}

func (l *LLDB) execSimple(msg string) []byte {
	buffer := make([]byte, 2048)
	content := "$"
	content += msg
	content += "#00"
	l.conn.Write([]byte(content))
	read, _ := l.conn.Read(buffer)
	return buffer[:read]
}
