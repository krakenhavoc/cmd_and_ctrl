package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// do_walk_test.go — #4012 (CR 608.2c): a Do's later steps wait for the
// answer to a prompt an earlier step asked, and a table waiting on that
// prompt keeps the later steps across a restore and an undo.

const DoWalkProbeOracle = "test-4012-do-walk-probe"

// doWalkProbeDamage is "this deals N damage to the first target", a step
// no catalog primitive spells as a value (DealDamage takes fixed IDs).
type doWalkProbeDamage struct{ N int }

func (s doWalkProbeDamage) Apply(ctx *effects.Context) error {
	if len(ctx.Item.Targets) == 0 {
		return nil
	}
	return ctx.Game.DealDamageToCreatureForEffect(ctx.Item.SourceCardID, ctx.Item.Targets[0].ID, s.N)
}

// doWalkProbeKharnAndYou is "this deals 1 damage to the first target
// and 2 damage to you" as one instance, so a Khârn question and a
// Selfless Squire announcement are owed by the same damage.
type doWalkProbeKharnAndYou struct{}

func (doWalkProbeKharnAndYou) Apply(ctx *effects.Context) error {
	g, it := ctx.Game, ctx.Item
	return g.DamageInstanceForEffect(func() error {
		if err := g.DealDamageToCreatureForEffect(it.SourceCardID, it.Targets[0].ID, 1); err != nil {
			return err
		}
		return g.DealDamageToPlayerForEffect(it.SourceCardID, it.Controller, 2)
	})
}

// doWalkSquireSeen records, as the next step begins, how many Selfless
// Squire triggers the earlier step's damage has produced (-1: not run).
var (
	DoWalkSquire     uuid.UUID
	DoWalkSquireSeen = -1
)

type doWalkProbeSeeSquire struct{}

func (doWalkProbeSeeSquire) Apply(ctx *effects.Context) error {
	DoWalkSquireSeen = squireTriggersLocked(ctx.Game, DoWalkSquire)
	return nil
}

// doWalkAlways is a condition that holds, so an If's steps run.
type doWalkAlways struct{}

func (doWalkAlways) Holds(*effects.Context) bool           { return true }
func (doWalkAlways) Condition() (game.FootprintStep, bool) { return game.FootprintStep{}, false }

func init() {
	effects.Register(effects.Spec{
		OracleID: DoWalkProbeOracle,
		Name:     "Do Walk Probe",
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Do Walk Probe deals 1 damage to target creature. Draw a card.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.Do(doWalkProbeDamage{N: 1}, effects.DrawCards{N: 1}),
			},
			{
				Label:   "{T}: Scry 2, then draw a card.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.Do(effects.Scry{N: 2}, effects.DrawCards{N: 1}),
			},
			{
				Label:   "{T}: If it holds, scry 1, then draw a card, then gain 1 life.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.Do(effects.If{Cond: doWalkAlways{}, Then: []effects.Applier{effects.Scry{N: 1}, effects.DrawCards{N: 1}}}, effects.GainLife{Amount: 1}),
			},
			{
				Label:   "{T}: Do Walk Probe deals 1 damage to target creature and 2 damage to you. Then look.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.Do(doWalkProbeKharnAndYou{}, doWalkProbeSeeSquire{}),
			},
		},
	})
}
