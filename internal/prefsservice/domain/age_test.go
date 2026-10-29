package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

func TestParseDateOfBirth(t *testing.T) {
	tests := []struct {
		name    string
		raw     *string
		want    time.Time
		wantErr bool
	}{
		{name: "nil is unknown", raw: nil, wantErr: true},
		{name: "empty is unknown", raw: ptr(""), wantErr: true},
		{name: "missing day is unknown", raw: ptr("2000/01"), wantErr: true},
		{name: "non-numeric is unknown", raw: ptr("2000/Jan/01"), wantErr: true},
		{name: "impossible date is unknown", raw: ptr("2000/02/31"), wantErr: true},
		{name: "zero-padded date parses", raw: ptr("1986/12/01"), want: time.Date(1986, time.December, 1, 0, 0, 0, 0, time.UTC)},
		{name: "unpadded date parses", raw: ptr("2000/1/2"), want: time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)},
		{name: "leap day parses", raw: ptr("2004/02/29"), want: time.Date(2004, time.February, 29, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDateOfBirth(tt.raw)
			if tt.wantErr {
				assert.ErrorIs(t, err, errUnknownDateOfBirth)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAgeAt(t *testing.T) {
	dob := time.Date(2006, time.March, 15, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		want int
	}{
		{name: "day before 18th birthday", now: time.Date(2024, time.March, 14, 23, 0, 0, 0, time.UTC), want: 17},
		{name: "on 18th birthday", now: time.Date(2024, time.March, 15, 0, 0, 0, 0, time.UTC), want: 18},
		{name: "day after 18th birthday", now: time.Date(2024, time.March, 16, 0, 0, 0, 0, time.UTC), want: 18},
		{name: "many years later is not affected by leap years", now: time.Date(2106, time.March, 15, 0, 0, 0, 0, time.UTC), want: 100},
		{name: "date of birth in the future is clamped to zero", now: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ageAt(dob, tt.now))
		})
	}
}

func TestAgeFromDateOfBirth(t *testing.T) {
	now := time.Date(2024, time.June, 1, 0, 0, 0, 0, time.UTC)

	t.Run("unknown date of birth is age zero with an error", func(t *testing.T) {
		age, err := ageFromDateOfBirth(nil, now)
		assert.ErrorIs(t, err, errUnknownDateOfBirth)
		assert.Equal(t, 0, age)
	})

	t.Run("one month safety margin keeps a user under the threshold near their birthday", func(t *testing.T) {
		// Exactly 18 years ago: shifted one month later, the user is 17.
		age, err := ageFromDateOfBirth(ptr("2006/06/01"), now)
		require.NoError(t, err)
		assert.Equal(t, 17, age)

		// 18 years and one month ago: the shift lands on today, the user is 18.
		age, err = ageFromDateOfBirth(ptr("2006/05/01"), now)
		require.NoError(t, err)
		assert.Equal(t, 18, age)
	})
}
