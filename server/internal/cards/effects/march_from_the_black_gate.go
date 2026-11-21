package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// March from the Black Gate — Enchantment {1}{B} (The Lord of the Rings:
// Tales of Middle-earth, #5948):
//
//	"When this enchantment enters and whenever an Army you control
//	 attacks, amass Orcs 1."
//
// One printed ability with two conditions, so ONE registered trigger over
// the enter and attack events (Sun Titan's shape). The attack half fires once
// per Army you control that attacks, read as the attack is declared; the
// Army's subtype is read from the card as it is now. Amass Orcs 1 is the
// ordinary keyword action (CR 701.47).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9259e918-ace5-4f13-9035-d8bc8c820d25",
		Name:         "March from the Black Gate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Amass: 1},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					if !attackDeclaredByYou(ev, source.Controller) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.HasSubtypeWord("Army")
				}, "March from the Black Gate — amass Orcs 1", Do(Amass{Subtype: "Orc", N: 1})),
		},
	})
}
