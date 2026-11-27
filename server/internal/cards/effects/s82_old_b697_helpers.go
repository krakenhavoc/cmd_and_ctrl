package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b697_helpers.go — Mercadian Masques, "Instants and sorceries" (S82
// OLD batch 697, #5619). Append-only.

// S82OldB697TargetPlayerSkipsNextCombatThisTurn is "Target player skips their
// next combat phase this turn" (Moment of Silence): a skip record that ends
// with the turn (game.SkipNextCombatPhaseThisTurnForEffect). A player whose
// combat has begun or passed loses nothing, and the record never waits for a
// later turn (CR 614.10).
func S82OldB697TargetPlayerSkipsNextCombatThisTurn(_ *game.StackItem, ctx *Context) error {
	label := ""
	if ctx.Item != nil {
		label = ctx.Item.Label
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			ctx.Game.SkipNextCombatPhaseThisTurnForEffect(t.ID, ctx.Source(), label)
		}
	}
	return nil
}

// S82OldB697TectonicBreak is "Each player sacrifices X lands of their
// choice": every living seat owes X lands, chosen in APNAP order and
// sacrificed together (CR 101.4, 701.21a). A player with fewer lands
// sacrifices what they have.
func S82OldB697TectonicBreak(_ *game.StackItem, ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), S82OldB475SeatsOwing(ctx.Game, x),
		SacrificeSpec("a land", Land()), "Tectonic Break — sacrifice X lands")
	return nil
}

// S82OldB697ClearTheLand is "Each player reveals the top five cards of their
// library, puts all land cards revealed this way onto the battlefield tapped,
// and exiles the rest". Every player reveals first, every land enters as one
// simultaneous batch under its owner's control, and the rest are exiled once
// the lands are in (a keyed tail, so a table waiting on an entry question is
// a restore point).
func S82OldB697ClearTheLand(_ *game.StackItem, ctx *Context) error {
	g := ctx.Game
	var lands []game.BatchEntry
	var rest []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		for _, id := range g.RevealTopOfLibraryForEffect(seat, ctx.Source(), 5, "Clear the Land — revealed from the top of the library") {
			if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
				lands = append(lands, game.BatchEntry{CardID: id, From: game.ZoneLibrary})
			} else {
				rest = append(rest, id)
			}
		}
	}
	tail := ctx.Tail(s82OldB697ClearTheLandExile)
	tail.Carry = rest
	return g.PutOntoBattlefieldTogetherKeyedForEffect(lands, game.ZoneEntryOptions{Tapped: true}, tail)
}

// s82OldB697ClearTheLandExile is "and exiles the rest": every revealed
// nonland card still in its library. An on-disk key: never renamed, never
// reused.
var s82OldB697ClearTheLandExile = CardTail("s82-old-b697/clear-the-land-exile", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.Tail.Carry {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			if err := ctx.Game.ExileCardForEffect(id); err != nil {
				return err
			}
		}
	}
	return nil
})
