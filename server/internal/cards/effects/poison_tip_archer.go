package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Poison-Tip Archer — Creature — Elf Archer {2}{B}{G}, 2/3 (EDHREC
// rank 1648):
//
//	"Reach
//	 Deathtouch
//	 Whenever another creature dies, each opponent loses 1 life."
//
// The Golgari aristocrats' table-wide Blood Artist: any creature,
// anyone's, and every opponent pays. Syr Konrad's condition — the
// EventLTB-into-a-graveyard read post-move, with the Archer's own
// death excluded by ID — and the batch 01 drain for the payoff.
// Life loss, not damage, so nothing prevents it. Fires once per
// creature, so a wrath is a trigger per body, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0e810bb-5f38-4045-a718-30d423c05659",
		Name:            "Poison-Tip Archer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherCreatureDied(ev, source, g)
			}, "Poison-Tip Archer — each opponent loses 1 life", func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 1)
			}),
		},
	})
}
