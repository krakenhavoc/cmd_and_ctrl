package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tiger-Tribe Hunter — Creature — Human Barbarian {3}{R}{R}, 4/4 (Adventures
// in the Forgotten Realms):
//
//	"Trample
//	 Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, you may sacrifice
//	 another creature. When you do, this creature deals damage equal to the
//	 sacrificed creature's power to target creature."
//
// Bolg of the North's shape: the "may" is the sacrifice prompt as the trigger
// resolves, a creature that really went raises the CR 603.12 reflexive
// trigger, which picks its target as it goes on the stack (CR 603.3d). The
// damage is the sacrificed creature's power as it last stood (CR 608.2h),
// dealt by the Hunter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d5b4e9fd-3df6-4d1d-8135-d7c8615b05a7",
		Name:            "Tiger-Tribe Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.S82AfrB09PackTactics("Tiger-Tribe Hunter — pack tactics: you may sacrifice another creature", func(g *game.Game, item *game.StackItem) error {
				return effects.MaySacrificeOne(g, item, "Tiger-Tribe Hunter — sacrifice another creature to deal damage equal to its power to target creature?", effects.Creature(), true, tigerTribeHunterSacrificed)
			}),
		},
	})
}

// tigerTribeHunterSacrificed raises the "when you do" once a creature has
// really gone.
var tigerTribeHunterSacrificed = effects.RegisterSacrificedOne("prompt-run/tiger-tribe-hunter-sacrifice", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label: "Tiger-Tribe Hunter — damage equal to the sacrificed creature's power to target creature",
		Cards: []uuid.UUID{sacrificed},
		Body:  tigerTribeHunterDamageBody,
	}.Apply(ctx)
})

var tigerTribeHunterDamageBody = game.ReflexiveBody("tiger-tribe-hunter/damage", effects.SimpleBody(tigerTribeHunterDamage),
	func(uuid.UUID, game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("target creature")
	})

func tigerTribeHunterDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	sacrificed := ctx.PayloadCards()
	targets := ctx.LegalTargets()
	if len(sacrificed) == 0 || len(targets) == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: targets[0].ID, Amount: effects.B29PowerAsItLastStood(g, sacrificed[0])}.Apply(ctx)
}
