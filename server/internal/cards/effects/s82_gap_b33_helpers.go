package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB33EachPlayerLosesAThirdOfTheirLife is "each player loses a third of
// their life, round up" (Pox): the caster first, then each opponent in seat
// order. A player at zero or less loses nothing.
func gapB33EachPlayerLosesAThirdOfTheirLife(ctx *Context) error {
	for _, id := range tablePlayers(ctx) {
		p := ctx.PlayerByID(id)
		if p == nil || p.Eliminated || p.Life <= 0 {
			continue
		}
		if third := (p.Life + 2) / 3; third > 0 {
			if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -third); err != nil {
				return err
			}
		}
	}
	return nil
}

// gapB33SacrificeAThird asks every seat to sacrifice a third (rounded up) of
// the permanents it controls that match, counted as the instruction starts.
// then is the registered rest of the card (the zero value for none).
func gapB33SacrificeAThird(ctx *Context, match CardPredicate, label string, then game.PromptRunThen) error {
	var asks []uuid.UUID
	for _, seat := range tablePlayers(ctx) {
		if p := ctx.PlayerByID(seat); p == nil || p.Eliminated {
			continue
		}
		owed := (len(permanentsControlledByMatching(ctx.Game, seat, match)) + 2) / 3
		for i := 0; i < owed; i++ {
			asks = append(asks, seat)
		}
	}
	spec := sacrificeSpec(label, match)
	reason := "Pox — sacrifice a third of the " + label
	if len(asks) == 0 {
		if then.Valid() {
			return ctx.Game.PlayersSacrificeThenKeyedForEffect(ctx.Source(), nil, spec, reason, then, ctx.RunData(nil))
		}
		return nil
	}
	if then.Valid() {
		return ctx.Game.PlayersSacrificeThenKeyedForEffect(ctx.Source(), asks, spec, reason, then, ctx.RunData(nil))
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks, spec, reason)
	return nil
}
