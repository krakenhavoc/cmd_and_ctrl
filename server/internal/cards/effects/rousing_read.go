package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rousing Read — Enchantment — Aura {2}{U} (Jumpstart 2020):
//
//	"Enchant creature
//	 When this Aura enters, draw two cards, then discard a card.
//	 Enchanted creature gets +1/+1 and has flying."
//
// Angelic Gift's shape with Racers' Scoreboard's loot. The enters
// ability is a trigger on the stack (WhenThisEnters), and the draws
// land before the discard prompt so a drawn card is a legal discard
// (b16DrawThenDiscard). The +1/+1 is a layer 7c modify and the flying
// a layer 6 grant, both scoped to the enchanted creature.
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:     "306549ad-84c7-4983-840b-9663799e9f40",
		Name:         "Rousing Read",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 2, Discards: 1},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rousing Read — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return b16DrawThenDiscard(g, item, 2)
				}),
		},
	})
}
