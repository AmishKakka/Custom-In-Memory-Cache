package internal

import (
	"bufio"
	"fmt"
	"os"
	"time"
	"strconv"
	"strings"
)

func (c *cache) ReadFromWAL(filename string) error {
	// openign the file in read-only mode
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Could not open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// reading the file line-by-line
		text := scanner.Text()
		parts := strings.Split(text, "|")
		// incase a process crashes
		if len(parts) < 2 {
			continue
		}
		key := parts[1]
		if parts[0] == "S" {
			// 1: key		2: val		3: ttl
			// parsing string back to Float or Int or bool or string
			var parsedVal any
			val := parts[2]
			if intVal, err := strconv.Atoi(val); err == nil {
				parsedVal = intVal
			} else if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
				parsedVal = floatVal
			} else if boolVal, err := strconv.ParseBool(val); err == nil {
				parsedVal = boolVal
			} else {
				parsedVal = val
			}
			// parsing string back to time.Time
			ttl, err := strconv.ParseInt(parts[3], 10, 64)
			if err != nil {
				continue
			}
			// Call the set() method directly to constantly update the linked list and map of keys
			c.set(key, parsedVal, time.Unix(0, ttl))
		}
		if parts[0] == "D" {
			c.delete(key)
		}
	}
	return scanner.Err()
}