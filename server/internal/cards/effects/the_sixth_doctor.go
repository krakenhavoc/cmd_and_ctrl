package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Sixth Doctor — Legendary Creature — Time Lord Doctor {4}{G}{U}, 3/3
// (Doctor Who):
//
//	"Time Lord's Prerogative — Whenever you cast a historic spell, copy it, except
//	 the copy isn't legendary. This ability triggers only once each turn.
//	 (Artifacts, legendaries, and Sagas are historic. A copy of a permanent spell
//	 becomes a token.)"
//
// Iron Man, Bleeding Edge's copy without the "may": a CopySpell whose except
// clause removes Legendary (CR 707.10a), keeping the original's targets; a copy of
// a permanent spell becomes a token as it resolves (CR 707.10f). The once-a-turn
// limit is on the TRIGGER (Mirelurk Queen's TriggersOnlyOnceEachTurn): a second
// historic spell cast the same turn does not trigger it.
//
// No simplification.
func init() {
	const label = "Time Lord's Prerogative — copy that historic spell, except the copy isn't legendary"
	Register(Spec{
		OracleID:     "49cb8d81-d8a0-4ae7-9750-6788a3e9c81b",
		Name:         "The Sixth Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b17HistoricSpellCastByYou(ev, source, g)
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CopySpell{
					StackID: ctx.Trigger().Event.CardID,
					Except:  func(v *game.PrintedValues) { v.RemoveSupertype("Legendary") },
				}.Apply(ctx)
			})),
		},
	})
}
