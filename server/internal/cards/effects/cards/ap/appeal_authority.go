package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Appeal // Authority — Sorcery // Sorcery with aftermath (Hour of Devastation):
//
//	Appeal {G}: "Until end of turn, target creature gains trample and gets
//	 +X/+X, where X is the number of creatures you control."
//	Authority {1}{W}: "Aftermath (Cast this spell only from your graveyard.
//	 Then exile it.) Tap up to two target creatures your opponents control.
//	 Creatures you control gain vigilance until end of turn."
//
// X is counted as Appeal resolves (CR 608.2h: the target's size is fixed
// then). The creatures that gain vigilance are the ones you control as
// Authority resolves (CR 611.2c).
func init() {
	const oracle = "3ecc4b57-3e0c-48b8-9952-e4b63b6bc45d"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Appeal",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			x := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsCreature() })
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Appeal — trample"}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Appeal — +X/+X"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Authority",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to two target creatures your opponents control", effects.OpponentControls()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance"},
				Label:    "Authority — vigilance",
			}.Apply(ctx)
		},
	})
}
