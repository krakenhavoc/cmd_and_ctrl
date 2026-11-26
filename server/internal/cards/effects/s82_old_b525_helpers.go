package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b525_helpers.go — the "you may have target creature get ±N/±N if a
// colour was spent" half-clauses of Eventide's hybrid spells (Cankerous
// Thirst).

// S82OldB525MayPump asks the caster whether target creature gets the
// power/toughness change until end of turn, then goes on to the answer's own
// follow-up. Registered keyed answers hold the card's wording.
func S82OldB525MayPump(key, label string, power, toughness int, then func(ctx *Context, r game.ConfirmAnswer) error) game.ConfirmThen {
	return MayAnswer(key, func(ctx *Context, r game.ConfirmAnswer) error {
		if r.Accepted && len(r.Carry) > 0 {
			if err := (BoostUntilEOT{Target: r.Carry[0], Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		if then == nil {
			return nil
		}
		return then(ctx, r)
	})
}

// S82OldB525Ask queues the yes/no for a pump on target.
func S82OldB525Ask(ctx *Context, question string, target uuid.UUID, then game.ConfirmThen, word string) error {
	return MayChoice{Question: question, Then: then, Carry: []uuid.UUID{target}, Word: word}.Apply(ctx)
}
