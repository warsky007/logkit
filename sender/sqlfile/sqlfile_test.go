package sqlfile

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/qiniu/logkit/conf"
	"github.com/qiniu/logkit/sender"
	"github.com/qiniu/logkit/utils/models"
)

func TestSQLFileSender(t *testing.T) {
	conf := conf.MapConf{
		sender.KeySQLFileRotateSize: "2097152",
		sender.KeySQLFileTable:      "table1",
		sender.KeyMaxSendRate:       "10000",
	}
	sender, err := NewSender(conf)
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()

	nr := 1000
	for i := 0; i < nr; i++ {
		data := []models.Data{
			{
				"name": fmt.Sprintf("annonym %d", i),
				"uid":  strconv.FormatInt(rand.Int63(), 10),
				"age":  rand.Int31n(100),
			},
		}
		if err := sender.Send(data); err != nil {
			t.Error(err)
		}
	}

	file := sender.(*Sender).w.sqlfile
	if file != nil {
		file.Seek(0, io.SeekStart)
		defer func() {
			os.Remove(file.Name())
			file = nil
		}()

		i := 0
		buf := bufio.NewReader(file)
		for {
			line, _, err := buf.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Error(err)
			}
			i++

			if !strings.HasPrefix(string(line), "INSERT INTO table1(age,name,uid) VALUES") {
				t.Errorf("unexpect sql statement, got %s, want prefixed with 'INSERT INTO table1(age,name,uid) VALUES'", line)
			}
		}

		if i != nr {
			t.Errorf("unexpect record count, got %d, want %d", i, nr)
		}
	}
}
