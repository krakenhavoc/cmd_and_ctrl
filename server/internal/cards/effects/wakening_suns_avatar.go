package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wakening Sun's Avatar — Creature — Dinosaur Avatar {5}{W}{W}{W},
// 7/7 (EDHREC rank 3233):
//
//	"When this creature enters, if you cast it from your hand,
//	 destroy all non-Dinosaur creatures."
//
// The Dinosaur deck's one-sided wrath on an 8-drop body. The entry
// trigger carries an intervening-if (CR 603.4): "you cast it from
// your hand" is read off the event log — the entry came from the
// stack, and the cast that put it there left the caster's hand
// (b30CastFromHand) — when the trigger fires and again as it
// resolves. A reanimated, blinked, cheated-in or commander-zone
// Avatar enters quietly, as printed. The wipe is the shared
// DestroyAllMatching over every creature at the table without the
// Dinosaur subtype — the Avatar is a Dinosaur and survives its own
// trigger.
//
// No simplifications. The one this used to declare — the mass-destroy
// path bypassing indestructible (#446), so an indestructible
// non-Dinosaur died where printed it would survive — was an engine
// gap, closed for every "destroy all" card at once in S30 (#470).
func init() {
	Register(Spec{
		OracleID:     "3c2aec69-ffd9-4a34-888c-58adbbb99bb5",
		Name:         "Wakening Sun's Avatar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b30CastFromHand(g, source.InstanceID)
			}, "Wakening Sun's Avatar — destroy all non-Dinosaur creatures", func(g *game.Game, item *game.StackItem) error {
				if !b30CastFromHand(g, item.SourceCardID) {
					return nil
				}
				return DestroyAllMatching{Match: And(Creature(), Not(HasSubtype("Dinosaur")))}.Apply(NewContext(g, item))
			}),
		},
	})
}
