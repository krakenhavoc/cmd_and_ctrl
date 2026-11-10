package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Irreverent Gremlin — Creature — Gremlin {1}{R}, 2/2 (Duskmourn: House of
// Horror, #4466):
//
//	"Menace
//	 Whenever another creature you control with power 2 or less enters, you
//	 may discard a card. If you do, draw a card. Do this only once each
//	 turn."
//
// Menace is a printed keyword. The trigger is Mentor of the Meek's: one per
// small creature entering under your control, the entering creature's power
// read as it enters (counters included) and not again as the ability
// resolves (the ruling). The rummage is the shared "you may discard a card.
// If you do, draw a card" prompt with one change: the discard is the doing,
// so only a card really discarded uses up the turn (Corruption of Towashi's
// record, b28DoneThisTurn). "Do this only once each turn" limits the DOING,
// not the triggering (CR 603.2h): the ability stops triggering once the
// turn's rummage is done, and a second trigger already on the stack does
// nothing when the first has been done, while a declined "may" or a trigger
// that was countered leaves the turn's rummage for a later creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44594004-52e7-4ce6-8eb0-61f72fd9c73b",
		Name:            "Irreverent Gremlin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.CurrentPower() <= 2 &&
					!b28DoneThisTurn(g, source.InstanceID, irreverentGremlinDone)
			}, irreverentGremlinLabel, func(g *game.Game, item *game.StackItem) error {
				if b28DoneThisTurn(g, item.SourceCardID, irreverentGremlinDone) {
					return nil
				}
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   item.Controller,
					Source:   item.SourceCardID,
					N:        1,
					UpTo:     true,
					Question: "Irreverent Gremlin — you may discard a card; if you do, draw a card",
				}, irreverentGremlinRummage, nil)
			}),
		},
	})
}

const (
	irreverentGremlinLabel = "Irreverent Gremlin — you may discard a card; if you do, draw a card (once each turn)"
	// irreverentGremlinDone is the once-each-turn record: a card was
	// discarded for the rummage this turn.
	irreverentGremlinDone = "Irreverent Gremlin — rummaged this turn"
)

// irreverentGremlinRummage is "If you do, draw a card" after the discard,
// keyed: it records the turn's one rummage only when a card left the hand.
// An on-disk key: never renamed, never reused.
var irreverentGremlinRummage = game.RegisterDiscardThenRule("choose-cards/irreverent-gremlin-discard-draw", nil,
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() == 0 {
			return nil
		}
		b28NoteDoneThisTurn(g, r.Source, irreverentGremlinDone)
		return g.DrawNForEffect(r.Player, 1)
	})
