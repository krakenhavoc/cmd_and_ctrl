package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baleful Mastery — Instant {3}{B}:
//
//	"You may pay {1}{B} rather than pay this spell's mana cost.
//	 If the {1}{B} cost was paid, an opponent draws a card.
//	 Exile target creature or planeswalker."
//
// Snuff Out's shape (CR 118.9's alternative cost) with a mana
// component instead of a life one and no condition gating the
// offer — declared directly as a bare game.AlternativeCost rather
// than through one of alternative_cost.go's keyword constructors,
// because this isn't a keyword: the card prints no reminder-text
// name for it, so there is no bundled rewrite (no cleared targets, no
// sacrifice trigger, no zone change) for a constructor to own.
//
// "AN opponent draws a card" is a resolution-time choice with no
// target word, exactly the clause effects.ChoosePlayer exists for
// (#929) — Slithermuse's "choose an opponent" is the same primitive.
// The draw happens BEFORE the exile, matching the printed order, and
// runs only when the discount was actually taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "adfcdadd-ddda-477b-8e72-0cae2430fb63",
		Name:         "Baleful Mastery",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			{
				Key:      "baleful-mastery",
				Label:    "Pay {1}{B} rather than pay this spell's mana cost",
				ManaCost: "{1}{B}",
			},
		},
		Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if !ctx.PaidAltCost("baleful-mastery") {
				return balefulMasteryExile(ctx)
			}
			return ChoosePlayer{
				Among:    Opponents,
				Question: "Baleful Mastery — an opponent draws a card",
				Then:     balefulMasteryDrawThenExile,
			}.Apply(ctx)
		},
	})
}

// balefulMasteryDrawThenExile is the ChoosePlayer.Then contract: it
// reads only the Context it is handed (ctx.Item, ctx.ChosenPlayer()),
// so an undo across the prompt resolves it against the restored game.
func balefulMasteryDrawThenExile(ctx *Context) error {
	if p := ctx.ChosenPlayer(); p != uuid.Nil {
		if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return balefulMasteryExile(ctx)
}

// balefulMasteryExile is "Exile target creature or planeswalker",
// re-checked against the item's own targets rather than a captured
// slice, for the same undo-safety reason.
func balefulMasteryExile(ctx *Context) error {
	if ctx.Item == nil || len(ctx.Item.Targets) == 0 || !ctx.IsTargetLegal(ctx.Item.Targets[0]) {
		return nil
	}
	return ExileTarget{Target: ctx.Item.Targets[0].ID}.Apply(ctx)
}
