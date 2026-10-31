package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leader, Super-Genius — Legendary Creature — Gamma Scientist Villain
// {2}{U}{U}, 1/3 (Marvel Super Heroes Commander, #4455, seam #4589):
//
//	"If a creature you control would connive, instead you draw a card, then
//	 that creature connives.
//	 At the beginning of combat on your turn, target creature you control
//	 connives."
//
// The first line replaces the connive (CR 614.1a): connive is an event
// (game.ConniveForEffect, CR 701.50a) and the replacement makes its
// controller draw a card before it (DrawACardThenConniveInstead). It applies
// to every connive of a creature you control, whoever's effect makes it, and
// the connive that follows still says the creature has connived (CR 701.50f),
// so "whenever a creature you control connives" triggers once. It does not
// apply to an opponent's creature. The beginning-of-combat trigger is the
// ordinary targeted connive, so with Leader on the battlefield as it
// resolves the creature draws two cards before it discards one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bec686dd-a9ea-4db5-bc20-8e40420b9a9b",
		Name:         "Leader, Super-Genius",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			DrawACardThenConniveInstead("Leader, Super-Genius — draw a card, then that creature connives"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Leader, Super-Genius — target creature you control connives",
					theTargetCreatureConnives),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
