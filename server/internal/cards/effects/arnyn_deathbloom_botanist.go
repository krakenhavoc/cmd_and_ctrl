package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arnyn, Deathbloom Botanist — Legendary Creature — Vampire Druid {2}{B},
// 2/2 (Secrets of Strixhaven):
//
//	"Deathtouch
//	 Whenever a creature you control with power or toughness 1 or less
//	 dies, target opponent loses 2 life and you gain 2 life."
//
// The size is the creature's last-known power and toughness (CR 603.10a),
// so a 3/3 that was shrunk to 1/1 when it died counts and a 1/1 pumped to
// 3/3 does not; Arnyn dying with the right size counts as well. The
// opponent is a real target, re-checked at resolution (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:        "3debd5c7-da3c-4130-8ed9-e156e10034b6",
		Name:            "Arnyn, Deathbloom Botanist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventLTB, socB20SmallCreatureYouControlDied,
					"Arnyn, Deathbloom Botanist — target opponent loses 2 life and you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return targetOpponentLosesAndYouGain(g, item, 2)
					}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
