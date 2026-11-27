package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB582BloodletterQuillDraw is Bloodletter Quill's resolution: draw a
// card, then lose 1 life for each blood counter on the artifact. The life
// waits for the draw (CR 608.2c) through the draw continuation, and counts the
// counters as the Quill is then, or as it last existed if it has left
// (CR 608.2h).
func S82OldB582BloodletterQuillDraw(g *game.Game, item *game.StackItem) error {
	return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{Ref: s82OldB582QuillLoseLife, Player: item.Controller, Source: item.SourceCardID})
}

// s82OldB582QuillLoseLife is "then you lose 1 life for each blood counter". An
// on-disk key: never renamed, never reused.
var s82OldB582QuillLoseLife = game.RegisterDrawThen("s82-old-b582/bloodletter-quill-lose-life", func(g *game.Game, d game.DrawThen) error {
	n := 0
	if c, ok := g.LookupCardForEffect(d.Source); ok && g.Battlefield.Contains(d.Source) {
		n = c.Counters["blood"]
	} else {
		n = g.LastKnownCountersForEffect(d.Source)["blood"]
	}
	if n <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(d.Source, d.Player, -n)
})

// S82OldB582FirstNoncreatureSpellOfTheTurn is "the first noncreature spell of
// a turn is cast" (Nullstone Gargoyle): any player's noncreature spell, and
// the per-player cast tallies (bumped before EventCast fires) say it is the
// only one this turn so far, spells cast before the source arrived included.
func S82OldB582FirstNoncreatureSpellOfTheTurn(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || spell.IsCreature() {
		return false
	}
	total := 0
	for _, p := range g.Seats {
		if p != nil {
			total += g.CastTallyFor(p.ID).Noncreature
		}
	}
	return total == 1
}

// S82OldB582CounterThatSpell is the body of "counter that spell" for a
// trigger on a cast.
func S82OldB582CounterThatSpell(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return CounterTarget{StackID: item.Trigger.Event.CardID}.Apply(NewContext(g, item))
}
