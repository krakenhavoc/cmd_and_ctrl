package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hulk's Thunderclap — Sorcery {1}{G} (Marvel Super Heroes Commander):
//
//	"As an additional cost to cast this spell, you may behold a Gamma
//	 creature. (You may choose a Gamma creature you control or reveal a
//	 Gamma creature card from your hand.)
//	 Target creature you control deals damage equal to its power to
//	 another target creature. If this spell's additional cost was paid,
//	 destroy target noncreature artifact or noncreature enchantment."
//
// The behold is an optional additional cost (OptionalBehold, #4642): the
// caster names the Gamma on reveal_ids when they take it, and the offer
// is not made when they have none to behold. The third target exists only
// for a cast that took it (CR 601.2c), which is Falling Timber's shape:
// WhenPaid rewrites the whole target statement, here the printed two
// clauses plus the artifact-or-enchantment one. An unbeholding cast is
// announced with two targets and cannot name a third.
//
// The damage is Bite Down's (oneSidedBite): slot 0 deals damage equal to
// its power as the spell resolves, slot 1 takes it. A source that has
// left is an illegal target (CR 608.2b) and deals none; the destruction is
// its own sentence and still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e377059d-a918-462f-920a-6b062bd5fcfd",
		Name:         "Hulk's Thunderclap",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.OptionalBehold("a", "Gamma"), effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("another target creature")),
			effects.TargetPermanent("target noncreature artifact or noncreature enchantment",
				effects.Or(effects.Artifact(), effects.Enchantment()), effects.Noncreature()),
		))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.OneSidedBite(item, ctx); err != nil {
				return err
			}
			if !ctx.WasBeheld() {
				return nil
			}
			t, ok := ctx.ClauseTarget(2)
			if !ok {
				return nil
			}
			return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
