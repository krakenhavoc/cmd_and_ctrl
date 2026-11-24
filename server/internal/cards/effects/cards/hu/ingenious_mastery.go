package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ingenious Mastery — Sorcery {X}{2}{U} (Strixhaven: School of Mages, #5502):
//
//	"You may pay {2}{U} rather than pay this spell's mana cost.
//	 If the {2}{U} cost was paid, you draw three cards, then an opponent creates
//	 two Treasure tokens and they scry 2. If that cost wasn't paid, you draw X
//	 cards."
//
// Baleful Mastery's shape: a bare alternative cost (CR 118.9) the resolution
// reads (PaidAltCost). Paid, X is 0 and the spell draws three, then "an
// opponent" is a resolution-time choice with no target word, asked of the
// caster through ChoosePlayer (#929); that opponent creates the two Treasures
// and scries 2. Not paid, the spell draws X cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dad43a2d-5b30-45f2-a34d-9e68d1fbb716",
		Name:         "Ingenious Mastery",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		AlternativeCosts: []game.AlternativeCost{{
			Key:      "ingenious-mastery",
			Label:    "Pay {2}{U} rather than pay this spell's mana cost",
			ManaCost: "{2}{U}",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !ctx.PaidAltCost("ingenious-mastery") {
				return effects.DrawCards{N: ctx.X()}.Apply(ctx)
			}
			if err := (effects.DrawCards{N: 3}).Apply(ctx); err != nil {
				return err
			}
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Ingenious Mastery — an opponent creates two Treasure tokens and scries 2",
				ThenKey:  ingeniousMasteryOpponent,
			}.Apply(ctx)
		},
	})
}

// ingeniousMasteryOpponent is the opponent chosen, keyed (#4253): they create
// two Treasures, then scry 2.
var ingeniousMasteryOpponent = effects.ChoosePlayerThen("option-pick/s82-old-b046-ingenious-mastery-opponent", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	if err := (effects.CreateToken{Controller: chosen, Template: effects.TreasureToken(), N: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.Scry{Player: chosen, N: 2}.Apply(ctx)
})
