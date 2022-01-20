package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/OmAsana/yapraktikum/internal/metrics"
)

func readFile(reader io.Reader) []string {
	r := bufio.NewReader(reader)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func writeToFile(write io.Writer, s string) (int, error) {
	i, err := write.Write([]byte(s + "\n"))
	if err != nil {
		return i, err
	}
	return i, nil
}

func TestUnmarshalFromReader(t *testing.T) {

	var buffer bytes.Buffer
	gauge := metrics.Gauge{
		Name:  "Blah",
		Value: 4,
	}

	d, err := json.Marshal(&gauge)
	assert.NoError(t, err)
	_, err = writeToFile(&buffer, string(d))
	assert.NoError(t, err)
	_, err = writeToFile(&buffer, string(d))
	assert.NoError(t, err)
	_, err = writeToFile(&buffer, string(d))
	assert.NoError(t, err)

	lines := readFile(&buffer)

	for _, l := range lines {
		//fmt.Println(l)
		var metric map[string]interface{}
		err := json.Unmarshal([]byte(l), &metric)
		assert.NoError(t, err)

		switch metric["mType"] {
		case "gauge":
			g := metrics.Gauge{
				metric["name"].(string),
				metric["value"].(float64),
			}
			fmt.Println(g)

		default:
			fmt.Println("unknown metric")
		}
		//fmt.Println(metric["mType"])
		//switch v := metric["mType"].(string); {
		//}
		//g, ok := metric.(metrics.Gauge)
		//if !ok {
		//	fmt.Println(ok)
		//}

	}
}
