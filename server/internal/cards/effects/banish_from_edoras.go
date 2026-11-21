package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Banish from Edoras — Sorcery {4}{W} (The Lord of the Rings: Tales of
// Middle-earth):
//
//	"This spell costs {2} less to cast if it targets a tapped creature.
//	 Exile target creature."
//
// Ride's End's reduction (CostsLessIfItTargets) over a tapped creature,
// judged against the announced target as the cost is determined
// (CR 601.2f). The exile is exileFirstLegalCardTarget: a target that is
// no longer legal on resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4508afd-7cc3-4176-8291-cfc2c155ac11",
		Name:         "Banish from Edoras",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a tapped creature.",
				And(Creature(), tappedPermanent())),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
