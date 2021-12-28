package pkg

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_floatIsNumber(t *testing.T) {
	type args struct {
		float    string
		isNumber bool
	}
	tests := []args{
		{
			"NaN",
			false,
		},
		{
			"inf",
			false,
		},
		{
			"-inf",
			false,
		},
		{
			"+inf",
			false,
		},
		{
			"-0",
			true,
		},
		{
			"0",
			true,
		},
		{
			"1.2",
			true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Test float %q", test.float), func(t *testing.T) {
			val, err := strconv.ParseFloat(test.float, 64)
			require.NoError(t, err)
			require.Equal(t, test.isNumber, FloatIsNumber(val))

		})

	}
}
