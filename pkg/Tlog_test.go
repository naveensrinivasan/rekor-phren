package pkg

import (
	"testing"
)

func TestTLogSize(t1 *testing.T) {
	type fields struct {
		TreeSize int64
		host     string
	}
	tests := []struct {
		name    string
		fields  fields
		want    int64
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				TreeSize: 5000000,
				host:     "http://localhost:8081/",
			},
			want: 5000000,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &tlog{
				TreeSize: int64(tt.fields.TreeSize),
				host:     tt.fields.host,
			}
			got, err := t.Size()
			if (err != nil) != tt.wantErr {
				t1.Errorf("Size() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got < tt.want { //nolint:typecheck
				t1.Errorf("Size() got = %v, want %v", got, tt.want)
			}
			t1.Logf("Size() got = %v, want %v", got, tt.want)
		})
	}
}
