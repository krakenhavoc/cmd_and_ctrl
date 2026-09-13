package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// fastlands.go — the Scars of Mirrodin / Kaladesh "fastland" cycle:
//
//	"This land enters tapped unless you control two or fewer other
//	 lands."
//	"{T}: Add {X} or {Y}."
//
// Three of the ten, the three this deck plays, plus Blackcleave Cliffs
// from the roadmap's batch 09 (#302), Razorverge Thicket from batch
// 12 (#305), Inspiring Vantage and Copperline Gorge from batch 13
// (#306), Blooming Marsh and Spirebluff Canal from batch 14 (#307),
// and Botanical Sanctum from batch 15 (#308) — all ten.
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
		// Roadmap batch 09 (#302).
		{"5ad94412-6f79-4c5d-bbd4-4ef5779a7b6d", "Blackcleave Cliffs", "B", "R"},
		// Roadmap batch 12 (#305).
		{"94f6c407-e665-4032-be13-a01e40c1f306", "Razorverge Thicket", "G", "W"},
		// Roadmap batch 13 (#306).
		{"3f17c60e-923a-4392-9da8-87d9ded009b7", "Inspiring Vantage", "R", "W"},
		{"a05f641c-15c9-43dc-ae0d-1ea372fd33d5", "Copperline Gorge", "R", "G"},
		// Roadmap batch 14 (#307).
		{"66fa2326-1b5d-41fb-b919-83bf9f383577", "Blooming Marsh", "B", "G"},
		{"eb0d8093-5f93-4b25-9384-08f9731bfb28", "Spirebluff Canal", "U", "R"},
		// Roadmap batch 15 (#308).
		{"88f8f683-738e-48f3-afff-c8f73f1033a2", "Botanical Sanctum", "G", "U"},
	} {
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Replacements:  []game.ReplacementEffect{EntersTappedUnless(otherLandsAtMost(2))},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
