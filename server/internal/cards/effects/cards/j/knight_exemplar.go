package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight Exemplar — Creature — Human Knight {1}{W}{W}, 2/2 (Magic Commander
// Collection, #5908):
//
//	"First strike
//	 Other Knight creatures you control get +1/+1 and have indestructible."
//
// One tribal anthem and one tribal keyword grant over the same filter, both
// reading effective subtypes (a changeling Knight counts).
//
// No simplification.
func init() {
	otherKnights := effects.TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:        "ecb198fc-5663-4380-b2da-ad54421e966d",
		Name:            "Knight Exemplar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(otherKnights, 1, 1),
			effects.TribalKeywordGrant(otherKnights, "indestructible"),
		},
	})
}
