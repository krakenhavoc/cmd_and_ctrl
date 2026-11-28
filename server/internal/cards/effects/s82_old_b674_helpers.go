package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b674_helpers.go — S82 OLD batch 674 (Prophecy, #5617).

// CheckImposedSacrifices is the boot-time refusal for a Spec.ImposedSacrifices
// entry that charges nothing: no printed Label, no kind of permanent to
// sacrifice, or neither Spells nor Activations (#6205).
func CheckImposedSacrifices(card string, is []game.ImposedSacrifice) {
	for i, s := range is {
		switch {
		case s.Label == "":
			panic(fmt.Sprintf("effects.Register: %q imposed sacrifice %d has no printed Label -- the move log and the picker quote it", card, i))
		case !s.Spells && !s.Activations:
			panic(fmt.Sprintf("effects.Register: %q imposed sacrifice %q reaches neither spells nor activated abilities", card, s.Label))
		case s.Kind.Label == "" || (!s.Kind.Any && len(s.Kind.Subtypes) == 0 && len(s.Kind.CardTypes) == 0):
			panic(fmt.Sprintf("effects.Register: %q imposed sacrifice %q names no kind of permanent to sacrifice", card, s.Label))
		}
	}
}

// S82OldB674EnteringCreaturePutter is the player who put the triggering
// creature onto the battlefield: the entry event's PutBy, or, on an entry that
// names nobody, the player it entered under (CR 110.2a's default). Overburden's
// "that player".
func S82OldB674EnteringCreaturePutter(g *game.Game, item *game.StackItem) uuid.UUID {
	return chantPutter(g, item)
}

// S82OldB674ReturnALandYouControl is "that player returns a land they control
// to its owner's hand": a mandatory pick of exactly one land from the
// player's own board (none if they control none), made as the trigger
// resolves (Kefnet the Mindful's pick without the "may").
func S82OldB674ReturnALandYouControl(player uuid.UUID, question string, keyed PermanentsThen) ChoosePermanents {
	return ChoosePermanents{
		Player:   player,
		Question: question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range PermanentsControlledBy(g, of) {
				if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
					out = append(out, id)
				}
			}
			return out, 1, 1
		},
		Keyed: keyed,
	}
}

// S82OldB674PlusOnePlusOneAndMinusOneMinusOne is Steal Strength's text, the
// same as Leeching Bite's: "Target creature gets +1/+1 until end of turn.
// Another target creature gets -1/-1 until end of turn." Each half is judged
// on its own target at resolution (CR 608.2b).
func S82OldB674PlusOnePlusOneAndMinusOneMinusOne(name string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
			if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: name + " — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
		}
		if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
			return BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1, Label: name + " — -1/-1"}.Apply(ctx)
		}
		return nil
	}
}
