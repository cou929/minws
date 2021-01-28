package minws

import "testing"

func Test_calcSecWebsocketAccept(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{"dGhlIHNhbXBsZSBub25jZQ=="},
			want: "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcSecWebsocketAccept(tt.args.key); got != tt.want {
				t.Errorf("calcSecWebsocketAccept() = %v, want %v", got, tt.want)
			}
		})
	}
}
