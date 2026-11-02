package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extinction Event — Sorcery {3}{B} (Marvel Universe, #4460):
//
//	"Choose odd or even. Exile each creature with mana value of the
//	 chosen quality. (Zero is even.)"
//
// The controller is asked odd or even as the spell resolves (a
// PickOption keyed to extinctionEventParity, so a table waiting on it
// restores), then every creature of that parity on the battlefield,
// theirs included, is exiled. A token or face-down creature has mana
// value 0, which is even (CR 202.3e); a creature's mana value is read as
// the exile happens. Exile is not destruction, so indestructible does
// not save one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a11a3a8-ee5f-4caf-bf84-c8248e07c825",
		Name:         "Extinction Event",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PickOption{
				Question: "Extinction Event — choose odd or even",
				Options:  []game.ChoiceOption{{Label: "Odd"}, {Label: "Even"}},
				ThenKey:  extinctionEventParity,
			}.Apply(ctx)
		},
	})
}

// extinctionEventParity exiles each creature whose mana value has the
// chosen parity. An on-disk key.
var extinctionEventParity = OptionPickThen("option-pick/extinction-event-parity", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 {
		return nil
	}
	odd := r.Index == 0
	return ExileAllMatching{
		Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			if !c.IsCreature() {
				return false
			}
			mv, ok := g.ManaValueForEffect(c)
			return ok && (mv%2 != 0) == odd
		},
	}.Apply(ctx)
})
