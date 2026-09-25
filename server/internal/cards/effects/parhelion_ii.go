package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Parhelion II — Legendary Artifact — Vehicle, 5/5, for {6}{W}{W}:
//
//	"Flying, first strike, vigilance
//	 Whenever Parhelion II attacks, create two 4/4 white Angel
//	 creature tokens with flying and vigilance that are attacking.
//	 Crew 4"
//
// "That are attacking" is the card. The Angels are PUT onto the
// battlefield attacking (CR 506.3c) rather than declared, so they
// fire no attack triggers — including Parhelion's own, which is what
// stops the ability making infinite Angels. The engine half is
// CreateTokensAttackingForEffect; this card is the reason it exists.
//
// The defending player is captured off the EventAttack that fired
// the trigger, not recomputed at resolution: by the time the ability
// resolves Parhelion II may have been removed, and the Angels still
// join the attack on whoever it was attacking. Capturing a plain
// uuid value is the same pattern PayUnless uses for its payer.
func init() {
	Register(Spec{
		OracleID:        "24d22bcb-8a77-4c47-a508-6f4bc093c1d0",
		Name:            "Parhelion II",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance"},
		Activated: []ActivatedAbility{{
			Label:  "Crew 4",
			Cost:   CrewCost(4),
			Effect: CrewEffect("Parhelion II"),
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Key: "Parhelion II — two attacking 4/4 Angels",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.CreateTokensAttackingForEffect(
					item.Controller, TokenCard("4/4 white Angel with flying and vigilance"), 2, item.Trigger.Event.Target)
			},
		}},
	})
}
