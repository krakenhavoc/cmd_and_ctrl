package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyoki, Sanity's Eclipse — Legendary Creature — Demon Spirit {4}{B}{B},
// 6/4 (Betrayers of Kamigawa, #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, target opponent exiles a
//	 card from their hand."
//
// The Champions of Kamigawa cast trigger (chkB1SpiritOrArcane) with
// Skullcap Snail's body: the target opponent picks the card from their
// own hand (targetsExileACardFromTheirHand). An opponent with no cards
// in hand exiles nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5240f9a-597b-43e7-9876-74559df1354e",
		Name:         "Kyoki, Sanity's Eclipse",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(chkB1SpiritOrArcane(),
				"Kyoki, Sanity's Eclipse — target opponent exiles a card from their hand",
				func(g *game.Game, item *game.StackItem) error {
					return targetsExileACardFromTheirHand(NewContext(g, item), "Kyoki, Sanity's Eclipse")
				}),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
