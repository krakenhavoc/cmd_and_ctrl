package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Path of Mettle // Metzali, Tower of Triumph (Rivals of Ixalan, S82 OLD batch
// 218, #5531).
//
// Front face, Legendary Enchantment {R}{W}:
//
//	"When Path of Mettle enters, it deals 1 damage to each creature that doesn't
//	 have first strike, double strike, vigilance, or haste.
//	 Whenever you attack with at least two creatures that have first strike,
//	 double strike, vigilance, and/or haste, transform Path of Mettle."
//
// Back face, Legendary Land:
//
//	"(Transforms from Path of Mettle.)
//	 {T}: Add one mana of any color.
//	 {1}{R}, {T}: Metzali deals 2 damage to each opponent.
//	 {2}{W}, {T}: Choose a creature at random that attacked this turn. Destroy
//	 that creature."
//
// The enters trigger reads every creature on the battlefield (anyone's) through
// the layers as it resolves and deals all the 1 damage in one damage instance.
// The attack trigger fires on the declaration that brings you to two attackers
// with one of those keywords (the keywords read through the layers). Metzali's
// 2 damage goes to each opponent at once; its last ability picks at random among
// the creatures on the battlefield that attacked this turn, whoever controls
// them (a creature that left and came back is a new object that has not
// attacked, CR 400.7), and destroys it; with none, nothing happens.
//
// No simplification.
func init() {
	const attackLabel = "Path of Mettle — you attacked with two or more creatures with first strike, double strike, vigilance and/or haste: transform it"
	effects.Register(effects.Spec{
		OracleID:     pathOfMettleOracleID,
		Name:         "Path of Mettle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisEnters("Path of Mettle — 1 damage to each creature that doesn't have first strike, double strike, vigilance, or haste", pathOfMettleEnters),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}}),
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OldB218AttackedWithKeywordCreatures(ev, source, g, attackLabel)
			}, attackLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.TransformThis{}.Apply(effects.NewContext(g, item))
			})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     pathOfMettleOracleID + "#1",
		Name:         "Metzali, Tower of Triumph",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, {T}: Metzali deals 2 damage to each opponent.",
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  metzaliDamage,
			},
			{
				Label:   "{2}{W}, {T}: Choose a creature at random that attacked this turn. Destroy that creature.",
				Cost:    effects.Plus(effects.ManaCost("{2}{W}"), effects.TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Effect:  metzaliDestroy,
			},
		},
	})
}

const pathOfMettleOracleID = "db9ea3f9-c723-422f-98cc-a3ef7ca2c290"

func pathOfMettleEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	g.RecomputeLayersIfStaleLocked()
	var victims []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && !effects.OldB218HasFirstStrikeDoubleStrikeVigilanceOrHaste(c) {
			victims = append(victims, c.InstanceID)
		}
	}
	return g.DamageInstanceForEffect(func() error {
		for _, id := range victims {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 1}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	})
}

func metzaliDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opponents := ctx.Opponents()
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range opponents {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: opp, Amount: 2}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	})
}

func metzaliDestroy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), effects.CreaturesThatAttackedThisTurn(g), 1)
	if len(pick) == 0 {
		return nil
	}
	return effects.DestroyTarget{Target: pick[0]}.Apply(ctx)
}
