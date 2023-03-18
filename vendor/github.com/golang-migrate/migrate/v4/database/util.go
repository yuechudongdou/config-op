package database

import (
	"fmt"
	"go.uber.org/atomic"
	"hash/crc32"
	"net"
	"strings"
	"time"
)

const advisoryLockIDSalt uint = 1486364155

// GenerateAdvisoryLockId inspired by rails migrations, see https://goo.gl/8o9bCT
func GenerateAdvisoryLockId(databaseName string, additionalNames ...string) (string, error) { // nolint: golint
	if len(additionalNames) > 0 {
		databaseName = strings.Join(append(additionalNames, databaseName), "\x00")
	}
	sum := crc32.ChecksumIEEE([]byte(databaseName))
	sum = sum * uint32(advisoryLockIDSalt)
	return fmt.Sprint(sum), nil
}

// CasRestoreOnErr CAS wrapper to automatically restore the lock state on error
func CasRestoreOnErr(lock *atomic.Bool, o, n bool, casErr error, f func() error) error {
	if !lock.CAS(o, n) {
		return casErr
	}
	if err := f(); err != nil {
		// Automatically unlock/lock on error
		lock.Store(o)
		return err
	}
	return nil
}

func CheckConnection(addr string, timeout time.Duration) bool {
	afterC := time.After(timeout)
	ticker := time.NewTicker(1 * time.Second)
	_, err := net.Dial("tcp", addr)
	if err == nil {
		return true
	}
	for {
		select {
		case <- ticker.C:
			_, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Println("connection is not ready, wait for next")
			} else {
				return true
			}
		case <- afterC:
			fmt.Println("wait connection timeout")
			return false
		}

	}
}