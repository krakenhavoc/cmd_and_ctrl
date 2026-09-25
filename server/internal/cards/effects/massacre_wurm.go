package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Massacre Wurm — Creature — Phyrexian Wurm, {3}{B}{B}{B}, 6/5:
//
//	"When this creature enters, creatures your opponents control get
//	 -2/-2 until end of turn.
//	 Whenever a creature an opponent controls dies, that player loses
//	 2 life."
//
// A one-sided Languish stapled to a 6/5, and then the second ability
// charges for every body it just killed. Against a token board at a
// four-player table it is routinely twenty damage spread across
// three opponents, on a creature that then attacks.
//
// # The two abilities are one combo and the order is what makes it
//
// The ETB shrink resolves, the creatures it killed die to the
// zero-toughness state-based action, and the death trigger — which
// is already on the battlefield by then — sees every one of those
// deaths. Since S23 that SBA sweep destroys its doomed set as ONE
// event (game/simultaneous.go), which is what makes "every one of
// those deaths" true rather than "whichever ones happened to be
// processed while the Wurm was still there". The Wurm is not in the
// sweep, so it would have survived either way, but the same fix is
// what lets a Blood Artist on the same board count correctly.
//
// # -2/-2, not damage and not destruction
//
// Same reasoning as Languish and Toxic Deluge: indestructible does
// not save a 2/2, survivors stay shrunk for the turn, and the
// affected set is snapshotted at resolution (CR 611.2c).
//
// # "An opponent controls" is relative to the WURM's controller
//
// Both abilities are read from the source's controller, not from
// whoever is taking the action — item.Controller on the trigger's
// stack item, which is the Wurm's controller. A Wurm you stole hits
// its new controller's opponents.
//
// The life loss is LIFE LOSS, not damage: no prevention, no
// lifelink, and the Wurm is not dealing it.
func init() {
	Register(Spec{
		OracleID:     "93cf50cf-0ecc-4d3e-abea-778c1ebacec4",
		Name:         "Massacre Wurm",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Massacre Wurm — opponents' creatures get -2/-2", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{
					Match:     And(Creature(), OpponentControls()),
					Power:     -2,
					Toughness: -2,
					Label:     "Massacre Wurm — -2/-2",
				}.Apply(ctx)
			}),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					dead, ok := diedCreature(ev, g)
					return ok && dead.Controller != source.Controller
				},
				Key: "Massacre Wurm — that player loses 2 life",
				// The dying creature's controller is read when the
				// trigger is BUILT, off the card in its
				// destination zone, because by the time the
				// trigger resolves that card may have moved again
				// (CR 603.10 — the ability uses last-known
				// information about the permanent).
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					dead, ok := diedCreature(ev, g)
					if !ok {
						return nil
					}
					item := game.NewTriggeredItem(source, "Massacre Wurm — that player loses 2 life")
					item.Params.Player = dead.Controller
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Params.Player, -2)
				},
			},
		},
	})
}
