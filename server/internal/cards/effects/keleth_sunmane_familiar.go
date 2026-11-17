package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keleth, Sunmane Familiar — Legendary Creature — Horse {1}{W}, 1/1
// (Commander Legends, #5507):
//
//	"Whenever a commander you control attacks, put a +1/+1 counter on it.
//	 Partner (You can have two commanders if both have partner.)"
//
// One trigger per commander declared as an attacker by Keleth's
// controller (Keleth itself included when it is one), over
// Card.IsCommander under that controller, so a stolen commander counts
// and your own commander under another player does not. The counter goes
// on the attacker if it is still on the battlefield as the trigger
// resolves (Good-Fortune Unicorn's body). Partner is the deck rule
// internal/deck reads off the oracle text (ADR 0144).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6df7f4b4-a3a8-44d0-845f-84304bf6a3f8",
		Name:         "Keleth, Sunmane Familiar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, kelethCommanderYouControlAttacked,
				"Keleth, Sunmane Familiar — put a +1/+1 counter on the attacking commander",
				func(g *game.Game, item *game.StackItem) error {
					return b33PutCounterOnEnteredCreature(item.Trigger.Event.CardID)(g, item)
				}),
		},
	})
}

// kelethCommanderYouControlAttacked is "whenever a commander you control
// attacks": a commander under the source's controller is declared as an
// attacker by that player.
func kelethCommanderYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Controller == source.Controller
}
