package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunscape Battlemage — Creature — Human Wizard {2}{W}, 2/2:
//
//	"Kicker {1}{G} and/or {2}{U} (You may pay an additional {1}{G}
//	 and/or {2}{U} as you cast this spell.)
//	 When this creature enters, if it was kicked with its {1}{G}
//	 kicker, destroy target creature with flying.
//	 When this creature enters, if it was kicked with its {2}{U}
//	 kicker, draw two cards."
//
// Thornscape Battlemage's shape with multi-symbol kickers (#2153): each
// enters ability is linked to one kicker cost (CR 702.33f) and reads it
// off the permanent's record as an intervening if (CR 603.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c8fd2232-e499-4889-a2ca-e486ea87d02b",
		Name:          "Sunscape Battlemage",
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers("{1}{G}", "{2}{U}"),
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, AllOf(Self, ThisKickedWith("{1}{G}")),
				"Sunscape Battlemage — kicked with {1}{G}, destroy target creature with flying",
				destroyChosenPermanent), TargetCreature("target creature with flying", HasKeyword("flying"))),
			On(game.EventETB, AllOf(Self, ThisKickedWith("{2}{U}")),
				"Sunscape Battlemage — kicked with {2}{U}, draw two cards",
				Do(DrawCards{N: 2})),
		},
	})
}
