package metrics

import "testing"

func TestCounter_IsValid(t *testing.T) {
	type fields struct {
		Name  string
		Value int64
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "valid",
			fields:  fields{Name: "blah", Value: 1},
			wantErr: false,
		},
		{
			name:    "value < 0",
			fields:  fields{Name: "blah", Value: -1},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Counter{
				Name:  tt.fields.Name,
				Value: tt.fields.Value,
			}
			if err := c.IsValid(); (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
