package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Induce Despair — Instant {2}{B} (Duel Decks: Zendikar vs. Eldrazi, #5548):
//
//	"As an additional cost to cast this spell, reveal a creature card from
//	 your hand.
//	 Target creature gets -X/-X until end of turn, where X is the revealed
//	 card's mana value."
//
// The card is revealed as the spell is cast, so a hand with no creature card
// can't cast it; X is the mana value it had when revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e8ff8b3-5b7f-47f2-9321-7a3b6f6e88aa",
		Name:         "Induce Despair",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		AdditionalCost: &game.AdditionalCost{
			Reveal: &game.RevealCost{CardType: "creature", Chosen: true},
			Label:  "Reveal a creature card from your hand",
		},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x, ok := ctx.Game.ChosenManaValueForEffect(item)
			if !ok {
				return nil
			}
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x, Label: "Induce Despair — -X/-X"}.Apply(ctx)
		},
	})
}
