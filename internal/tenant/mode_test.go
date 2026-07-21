package tenant

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Mode
		wantErr bool
	}{
		{name: "B2BLowercase", input: "b2b", want: B2B},
		{name: "B2CLowercase", input: "b2c", want: B2C},
		{name: "B2BUppercase", input: "B2B", want: B2B},
		{name: "MixedCase", input: "b2C", want: B2C},
		{name: "UnknownValue", input: "multi", wantErr: true},
		{name: "EmptyString", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseMode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
