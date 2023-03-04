package csvqlctl_test

import (
	"adrianolaselva.github.io/csvql/cmd/csvqlctl"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	FileModeDefault os.FileMode = 0644
)

var asserts = []struct {
	args      []string
	filePath  string
	fileName  string
	data      string
	delimiter string
	out       string
	queries   []string
}{
	{
		args:     []string{"-f", "./../../.tmp/0001.csv", "-d", ";"},
		filePath: "./../../.tmp",
		fileName: "0001.csv",
		data: strings.Join([]string{
			"id;name;email",
			"0001;teste_1;teste_1@gmail.com",
			"0002;teste_2;teste_2@gmail.com",
			"0003;teste_3;teste_3@gmail.com",
			"0004;teste_4;teste_4@gmail.com",
			"0005;teste_5;teste_5@gmail.com",
		}, "\n"),
		delimiter: ";",
		out:       "",
		queries: []string{
			"select * from rows;",
		},
	},
}

func TestShouldExecuteWithSuccess(t *testing.T) {
	cmd, err := csvqlctl.New().Command()
	assert.NoError(t, err)

	for _, test := range asserts {
		_ = os.MkdirAll(test.filePath, os.ModePerm)
		err := os.WriteFile(filepath.Join(test.filePath, test.fileName), bytes.NewBufferString(test.data).Bytes(), FileModeDefault)
		assert.NoError(t, err)

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs(test.args)

		err = cmd.Execute()
		assert.NoError(t, err)

		for _, q := range test.queries {
			_, err = cmd.OutOrStdout().Write([]byte(q))
			assert.NoError(t, err)
			if out := cmd.OutOrStdout(); out != os.Stdout {
				assert.Equal(t, q, fmt.Sprintf("%s", out))
			}
		}
	}
}
