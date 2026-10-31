package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const kingpinLabel = "Kingpin — create two Treasure tokens"

// Kingpin, Wilson Fisk — Legendary Creature — Human Villain {3}{B}, 3/6
// (Marvel Super Heroes Commander, #4455):
//
//	"Menace
//	 Whenever you sacrifice Kingpin or another creature, create two
//	 Treasure tokens. This ability triggers only once each turn."
//
// The trigger is Ravenous Squirrel's EventSacrifice watch narrowed to
// creatures; the sacrifice event fires before the zone move, so Kingpin
// sacrificing himself is seen. The once-per-turn clause is the
// b11TriggeredThisTurn tally keyed on the label (Interface Ace), checked
// when the event is judged, so a second sacrifice the same turn never
// reaches the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "062ffa63-1985-4409-b3ad-d42407ee755a",
		Name:            "Kingpin, Wilson Fisk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && !b11TriggeredThisTurn(g, source.InstanceID, kingpinLabel)
			}, kingpinLabel, Do(CreateToken{Template: TreasureToken(), N: 2})),
		},
	})
}
