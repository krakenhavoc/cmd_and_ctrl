package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vanish into Eternity — Instant {2}{W} (Phyrexia: All Will Be One):
//
//	"This spell costs {3} more to cast if it targets a creature.
//	 Exile target nonland permanent."
//
// Dragon's Prey's target-reading increase: the announced target is judged as
// the cost is determined (CR 601.2f, after 601.2c), so a creature costs
// {5}{W} and any other nonland permanent {2}{W}. The exile is a plain one: a
// token ceases to exist and a card goes to exile; nothing is returned.
//
// No simplification.
func init() {
	creatureTax := CostsMore(3, "This spell costs {3} more to cast if it targets a creature.", ItTargets(Creature()))
	creatureTax.ReadsTargets = true
	Register(Spec{
		OracleID:          "c14e422b-6412-45ce-8e82-d47ba61557a0",
		Name:              "Vanish into Eternity",
		Completeness:      CompletenessFull,
		Purpose:           game.Purpose{Answers: game.AnswerRemove},
		Targets:           TargetPermanent("target nonland permanent", Not(Land())),
		SelfCostModifiers: []game.CostModifier{creatureTax},
		OnResolve:         func(_ *game.StackItem, ctx *Context) error { return exileTheFirstTargetIfLegal(ctx) },
	})
}
