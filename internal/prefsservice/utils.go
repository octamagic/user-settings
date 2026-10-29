package prefsservice

import (
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
)

func GetUnmatchedUserUrns(userUrns []string, settings []prefsrepo.Setting) []string {
	unmatchedUserUrns := []string{}
	for _, userUrn := range userUrns {
		found := false
		for _, setting := range settings {
			if setting.UserUrn == userUrn {
				found = true
				break
			}
		}
		if !found {
			unmatchedUserUrns = append(unmatchedUserUrns, userUrn)
		}
	}
	return unmatchedUserUrns
}
