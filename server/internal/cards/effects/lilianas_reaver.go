package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana's Reaver — Creature — Zombie {2}{B}{B}, 4/3 (Foundations
// Commander):
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, that player
//	 discards a card and you create a tapped 2/2 black Zombie creature
//	 token."
//
// "That player" is the player dealt the damage (the trigger event's
// target), who chooses their own discard (CR 701.9a). The Zombie is made
// whether or not the hand had a card to discard (printed ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "97332d4e-a55c-4e4f-9040-bf426a7c8e39",
		Name:            "Liliana's Reaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Liliana's Reaver — that player discards a card and you create a tapped 2/2 black Zombie token",
				Do(
					Instruction(func(ctx *Context) error {
						victim := ctx.Trigger().Event.Target
						if p := ctx.Game.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
							return nil
						}
						ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player:   victim,
							Source:   ctx.Source(),
							N:        1,
							Question: "Liliana's Reaver — discard a card",
						})
						return nil
					}),
					CreateTokenAdvanced{Spec: Token(TokenCard("2/2 black Zombie")).EntersTapped(), N: 1},
				)),
		},
	})
}
