package cfg

import (
	"testing"
)

func TestGetEnv(t *testing.T) {
	testCases := []struct {
		name     string
		envName  string
		envValue string
		wantErr  bool
	}{
		{
			"should return error if env doesn't exists",
			"NON_EXISTING_ENV",
			"",
			true,
		},
		{
			"should return env value",
			"DATABASE_URL",
			"postgresql://postgres:postgres@localhost:5432/pulse_db",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.wantErr {
				t.Setenv(tc.envName, tc.envValue)
			}

			got, err := GetEnv(tc.envName)
			if (err != nil) != tc.wantErr {
				t.Fatalf("GetEnv() unexpected error state: got error = %v, wantErr = %v", err, tc.wantErr)
			}

			if !tc.wantErr && got != tc.envValue {
				t.Errorf("GetEnv() = %q, want %q", got, tc.envValue)
			}
		})
	}
}
