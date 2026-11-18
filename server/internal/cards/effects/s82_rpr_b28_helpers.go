package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rprB28StealAllMatching is "gain control of all creatures <matching> until
// end of turn. Untap those creatures. They gain haste until end of turn"
// (Mob Rule). The creatures are fixed as the spell resolves (CR 608.2c);
// every creature is taken, then untapped, then given haste, in that order.
func rprB28StealAllMatching(ctx *Context, label string, match CardPredicate) error {
	creatures := MatchingBattlefield(ctx, match)
	for _, c := range creatures {
		if err := (GainControl{
			Target:   c.InstanceID,
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    label + " — gain control until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, c := range creatures {
		if err := (UntapTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, c := range creatures {
		if err := (GrantKeywordUntilEOT{
			Target:   c.InstanceID,
			Keywords: []string{"haste"},
			Label:    label + " — haste until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// rprB28CreatureWithPower is a creature whose layered power passes `ok`.
func rprB28CreatureWithPower(ok func(power int) bool) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && ok(c.CurrentPower())
	}
}
