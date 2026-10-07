package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sami, Wildcat Captain — Legendary Creature — Human Artificer Rogue
// {4}{R}{W}, 4/4:
//
//	"Double strike, vigilance
//	 Spells you cast have affinity for artifacts. (They cost {1} less to
//	 cast for each artifact you control.)"
//
// The grant is an ordinary battlefield cost modifier (ADR 0048 §1) with
// no spell filter, one instance per Sami, so a second Sami stacks
// (CR 702.41b). Affinity only ever reduces generic mana (CR 702.41a),
// which CostsLessEach does. It reads the artifacts at cast time, so a
// Treasure sacrificed to pay for the spell does not retroactively
// change the price. Sami's own cost is not reduced by herself: she is in
// hand when cast, and the modifier only counts from the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c4175a34-70bb-47e5-8406-efdc4d2cd079",
		Name:            "Sami, Wildcat Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance"},
		CostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsYouControl(Artifact()),
				"Spells you cast have affinity for artifacts.", YourSpell()),
		},
	})
}
