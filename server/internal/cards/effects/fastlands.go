package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fastlands.go — the Scars of Mirrodin / Kaladesh "fastland" cycle:
//
//	"This land enters tapped unless you control two or fewer other
//	 lands."
//	"{T}: Add {X} or {Y}."
//
// Three of the ten, the three this deck plays.
//
// "Two or fewer OTHER lands" is the clause that makes the condition
// free to evaluate: the replacement pipeline runs before the land is
// pushed, so a walk of the battlefield counts other lands and
// nothing else. No exclusion logic, no off-by-one — the land
// genuinely is not there yet.
//
// The mirror-image slowlands ("two or MORE other lands") share the
// same helper with the comparison flipped; see slowlands.go.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"2d899466-b1eb-4901-b626-1f2fb09b786d", "Concealed Courtyard", "W", "B"},
		{"a2b48695-f7d7-42ce-a8a0-2a723428542a", "Darkslick Shores", "U", "B"},
		{"9e7a240d-dc33-47ac-9f17-77fab4c1c340", "Seachrome Coast", "W", "U"},
	} {
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Replacements:  []game.ReplacementEffect{EntersTappedUnless(otherLandsAtMost(2))},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
