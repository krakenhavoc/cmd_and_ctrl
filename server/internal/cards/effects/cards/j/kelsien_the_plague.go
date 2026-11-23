package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kelsien, the Plague — Legendary Creature — Human Assassin {R}{W}{B}, 2/2
// (Commander 2020, #5513):
//
//	"Vigilance, haste
//	 Kelsien gets +1/+1 for each experience counter you have.
//	 {T}: Kelsien deals 1 damage to target creature you don't control. When
//	 that creature dies this turn, you get an experience counter."
//
// The static is a layer 7c modify on Kelsien, re-read on every recompute
// off its controller's experience counters (Aang, Airbending Master's
// count). The tap ability deals the damage, then creates Grim Javelineer's
// "when that creature dies this turn" delayed trigger about the creature as
// the object it is now (CR 603.7a, 603.7c), controlled by the ability's
// controller; it gives that player one experience counter (CR 122.1) if the
// creature dies before the turn ends, whatever killed it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b3326a5-18c5-4d45-90e1-f6d00ca2bced",
		Name:            "Kelsien, the Plague",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
		Static: []game.StaticAbility{{
			Label:     "Kelsien gets +1/+1 for each experience counter you have.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.ExperienceCounters(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Kelsien deals 1 damage to target creature you don't control. When that creature dies this turn, you get an experience counter.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			Effect:  kelsienPing,
		}},
	})
}

// kelsienExperienceBody is the delayed trigger's resolution.
var kelsienExperienceBody = game.SimpleDelayedBody("kelsien-the-plague/experience-counter", effects.YouGetAnExperienceCounter)

// kelsienPing is the tap ability: 1 damage, then the delayed trigger about
// the creature, which is still on the battlefield (state-based actions have
// not been checked yet), so a creature the damage kills fires it.
func kelsienPing(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.WhenThatCreatureDiesThisTurn(ctx, id, "Kelsien, the Plague — when that creature dies this turn, you get an experience counter", kelsienExperienceBody)
}
