package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Okiba-Gang Shinobi — Creature — Rat Ninja {3}{B}{B}, 3/2 (Betrayers of
// Kamigawa, #5602):
//
//	"Ninjutsu {3}{B} ({3}{B}, Return an unblocked attacker you control
//	 to hand: Put this card onto the battlefield from your hand tapped
//	 and attacking.)
//	 Whenever this creature deals combat damage to a player, that player
//	 discards two cards."
//
// The keyword is Ninjutsu("{3}{B}") (ninjutsu.go). "That player" is the
// player the damage event named; they choose the two cards they discard
// (CR 701.9b), through the ordinary discard prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efe3b58f-6f83-45d5-beb1-7dde1cae9dd7",
		Name:         "Okiba-Gang Shinobi",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Okiba-Gang Shinobi — that player discards two cards",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player: item.Trigger.Event.Target,
						Source: item.SourceCardID,
						N:      2,
					})
					return nil
				}),
		},
	})
}
