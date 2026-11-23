package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b05_helpers.go — S82 ONE batch 5 (#5479): the sentences the
// toxic-matters cards share. "Has toxic" is a total toxic value above
// zero (CR 702.164b), read through game.ToxicTotal so a granted toxic
// counts exactly as a printed one.

// WithToxic is the target predicate for "creature with toxic".
func WithToxic() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.ToxicTotal(&c) > 0 }
}

// s82OneB05HasToxic reads, as the ability resolves, whether the
// permanent `id` has toxic. A permanent that has left has none.
func S82OneB05HasToxic(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && OnBattlefield(g, id) && game.ToxicTotal(&c) > 0
}

// s82OneB05AttachedHasToxic is the Aura trigger's "if enchanted
// creature has toxic": the host of the source, through the layers.
func S82OneB05AttachedHasToxic(g *game.Game, source *game.Card) bool {
	if source == nil || source.AttachedTo.Kind != game.TargetCard {
		return false
	}
	return S82OneB05HasToxic(g, source.AttachedTo.ID)
}

// s82OneB05CondEnchantedHasToxic is that sentence as an intervening
// "if" (CR 603.4), re-read as the trigger resolves.
type S82OneB05CondEnchantedHasToxic struct{}

func (S82OneB05CondEnchantedHasToxic) Holds(ctx *Context) bool {
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	return ok && S82OneB05AttachedHasToxic(ctx.Game, &src)
}

func (S82OneB05CondEnchantedHasToxic) Condition() (game.FootprintStep, bool) {
	return game.FootprintStep{}, false
}
