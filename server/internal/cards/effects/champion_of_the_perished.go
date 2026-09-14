package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champion of the Perished — Creature — Zombie {B}, 1/1 (EDHREC rank
// 2777):
//
//	"Whenever another Zombie you control enters, put a +1/+1 counter
//	 on this creature."
//
// Champion of the Parish for Zombies. "Another Zombie you control" is
// any other permanent with the type entering under the controller's
// control — a token, a changeling, a Zombie stolen and re-entering —
// read post-layer. The counter goes on the Champion if it is still
// on the battlefield when the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3654fbd-16a5-4953-84ac-534e8421032f",
		Name:         "Champion of the Perished",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, "Zombie")
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Champion of the Perished — put a +1/+1 counter on it",
					func(g *game.Game, item *game.StackItem) error {
						if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
							return nil
						}
						return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
