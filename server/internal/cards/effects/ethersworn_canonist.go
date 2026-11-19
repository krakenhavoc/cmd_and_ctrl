package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethersworn Canonist — Artifact Creature — Human Cleric {1}{W}, 2/2:
//
//	"Each player who has cast a nonartifact spell this turn can't cast
//	 additional nonartifact spells."
//
// Read off the per-turn tally: nonartifact spells cast so far is Total
// minus Artifact (CastTally counts both as each spell became cast, CR
// 601.2i). The spell being announced is judged by its own type line, so an
// artifact spell is never refused however many nonartifact spells came
// first. Symmetrical: it binds its controller too.
func init() {
	Register(Spec{
		OracleID:     "ec3471ad-b65c-4e35-a7d1-68da9555a9ec",
		Name:         "Ethersworn Canonist",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{{
			Label: "Each player who has cast a nonartifact spell this turn can't cast additional nonartifact spells.",
			Forbids: func(q game.CastQuery) bool {
				if q.Card.IsArtifact() {
					return false
				}
				t := q.Game.CastTallyFor(q.Controller)
				return t.Total-t.Artifact >= 1
			},
		}},
	})
}
