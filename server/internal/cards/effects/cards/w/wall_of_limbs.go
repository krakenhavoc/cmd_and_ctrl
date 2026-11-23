package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Limbs — Creature — Zombie Wall {2}{B}, 0/3 (Tarkir: Dragonstorm
// Commander):
//
//	"Defender
//	 Whenever you gain life, put a +1/+1 counter on this creature.
//	 {5}{B}{B}, Sacrifice this creature: Target player loses X life,
//	 where X is this creature's power."
//
// Ajani's Pridemate's life-gain row: once per life-gain event however
// much was gained. The sacrifice is a cost, so X is the Wall's power as it
// last existed on the battlefield (CR 608.2h), counters included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1aabb6bd-8061-42a5-9a71-f9eac718bd22",
		Name:            "Wall of Limbs",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.LifeGainPaying(
				effects.WheneverYouGainLife("Wall of Limbs — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				game.LifeGainPayoff{Counters: 1}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}{B}, Sacrifice this creature: Target player loses X life, where X is this creature's power.",
			Cost:    effects.Plus(effects.ManaCost("{5}{B}{B}"), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  wallOfLimbsDrain,
		}},
	})
}

// wallOfLimbsDrain makes the target player lose life equal to the Wall's
// power as it last existed (the sacrifice was a cost, so the Wall is
// gone by now).
func wallOfLimbsDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	x := effects.B43PowerNowOrLastKnown(g, item.SourceCardID)
	if x <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -x)
		}
	}
	return nil
}
