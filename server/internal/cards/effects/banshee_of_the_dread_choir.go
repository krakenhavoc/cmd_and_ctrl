package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banshee of the Dread Choir — Creature — Spirit {3}{B}{B}, 4/4 (Commander
// 2015, #5546):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)
//	 Whenever this creature deals combat damage to a player, that player
//	 discards a card."
//
// Myriad is the shared trigger (myriad.go, CR 702.116). The discard is
// Blazing Specter's: the damaged player chooses the card. Each myriad copy
// is a Banshee too, so every player a copy hits discards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "282b4dbf-ec00-48ba-8010-73bee489fa7a",
		Name:         "Banshee of the Dread Choir",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Banshee of the Dread Choir"),
			WheneverThisDealsCombatDamageToAPlayer("Banshee of the Dread Choir — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return discardOne(g, item, item.Trigger.Event.Target)
				}),
		},
	})
}
