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

func TestContains(t *testing.T) {
	type args struct {
		list  []string
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "countains",
			args: struct {
				list  []string
				value string
			}{list: []string{"a", "b"}, value: "a"},
			want: true,
		},
		{
			name: "does not countain",
			args: struct {
				list  []string
				value string
			}{list: []string{"a", "b"}, value: "c"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.list, tt.args.value); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
