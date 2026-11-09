package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// New Argive — Plane — Dominaria (Commander Masters, #4870):
//
//	"Whenever a historic creature you control attacks, it gets +2/+2
//	 until end of turn. (Artifacts, legendaries, and Sagas are historic.)
//	 Whenever chaos ensues, reveal cards from the top of your library until
//	 you reveal a historic card. Put that card into your hand and the rest
//	 on the bottom of your library in a random order."
//
// A plane (ADR 0216 §6). The first ability triggers once for each historic
// creature the planar controller attacks with (CR 700.6); the +2/+2 is
// locked in as it resolves, so it lasts the turn even if the creature stops
// being historic (the ruling). The chaos ability (CR 311.7) reveals from the
// planar controller's library until a historic card, puts it into their hand
// and the rest on the bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08427dab-3874-4a29-bd63-cbd16c1d229d",
		Name:         "New Argive",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && attacker.IsCreature() && b09IsHistoric(attacker)
			}, "New Argive — that historic creature gets +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return BoostUntilEOT{Target: item.Trigger.Event.CardID, Power: 2, Toughness: 2, Label: "New Argive — +2/+2"}.Apply(NewContext(g, item))
				}),
			WheneverChaosEnsues("New Argive — reveal until a historic card; put it into your hand and the rest on the bottom of your library in a random order",
				func(g *game.Game, item *game.StackItem) error {
					return revealUntilToHandRestOnBottom(NewContext(g, item), b09IsHistoric, "New Argive — revealed until a historic card")
				}),
		},
	})
}
