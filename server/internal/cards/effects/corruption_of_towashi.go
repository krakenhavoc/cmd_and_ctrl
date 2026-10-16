package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corruption of Towashi — Enchantment {4}{U} (March of the Machine):
//
//	"When this enchantment enters, incubate 4.
//	 Whenever a permanent you control transforms or a permanent you
//	 control enters transformed, you may draw a card. Do this only once
//	 each turn."
//
// One ability watching two events (OnAny): a permanent of yours turning
// over in place (EventTransform; a convert is not a transform, so it is
// declined), or one entering on its back face (EventETB read through
// enteredTransformedUnderYourControl, which explains why no new event
// is needed).
//
// "You may draw a card" is decided as the ability resolves (MayChoice),
// not when it goes on the stack. "Do this only once each turn" limits
// the DOING, as G'raha Tia's and Pantlaza's do: the use is recorded
// (Game.NoteDoneThisTurn) only when the card is actually drawn, and it
// is read when the ability triggers (so no trigger is put on the stack
// once it is spent) and again as it resolves (so the second of two
// triggers stacked before either resolved does nothing). A declined
// "may", or a trigger that was countered, leaves the turn's one draw
// for a later one, as printed. Gating at trigger time alone would be
// weaker than printed in exactly that case, and the resolution check
// removes the other difference, so the two are equivalent except that
// the do-nothing second trigger never reaches the stack.
func init() {
	const doneLabel = "Corruption of Towashi — draw (done)"
	notYetDone := func(g *game.Game, source uuid.UUID) bool {
		return g.ResolvedThisTurn(source, doneLabel) == 0
	}
	Register(Spec{
		OracleID:     "242c39b4-5ef0-469a-a852-738f79d93bd6",
		Name:         "Corruption of Towashi",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Corruption of Towashi — incubate 4", Do(Incubate{N: 4})),
			OnAny([]game.EventKind{game.EventTransform, game.EventETB},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if source == nil || !notYetDone(g, source.InstanceID) {
						return false
					}
					if ev.Kind == game.EventTransform {
						return ev.Actor == source.Controller && !ev.Converted
					}
					return enteredTransformedUnderYourControl(ev, source, g)
				},
				"Corruption of Towashi — you may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					if !notYetDone(g, item.SourceCardID) {
						return nil
					}
					return MayChoice{
						Question: "Corruption of Towashi — draw a card?",
						Draws:    1,
						OnYes: func(ctx *Context) error {
							if !notYetDone(ctx.Game, ctx.Item.SourceCardID) {
								return nil
							}
							ctx.Game.NoteDoneThisTurn(ctx.Item.SourceCardID, doneLabel)
							return DrawCards{N: 1}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
