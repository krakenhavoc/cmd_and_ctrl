package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	otherKnights := TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "ecb198fc-5663-4380-b2da-ad54421e966d",
		Name:            "Knight Exemplar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{
			TribalAnthem(otherKnights, 1, 1),
			TribalKeywordGrant(otherKnights, "indestructible"),
		},
	})
}
