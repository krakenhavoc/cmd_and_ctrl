package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spitting Dilophosaurus — Creature — Dinosaur {2}{B}, 3/2 (Jurassic World
// Collection, #5948):
//
//	"Whenever this creature enters or attacks, put a -1/-1 counter on up
//	 to one target creature.
//	 Creatures your opponents control with -1/-1 counters on them can't
//	 block."
//
// One enters-or-attacks trigger (CR 603.2) with an "up to one" target. The
// static is a layer-6 restriction on every creature an opponent controls
// that has at least one -1/-1 counter, read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19f2c2c1-cd29-4e66-b5d9-e85e1979d843",
		Name:         "Spitting Dilophosaurus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEntersOrAttacks("Spitting Dilophosaurus — put a -1/-1 counter on up to one target creature", effects.R35NightmareSowerCounter),
				effects.UpToOneTargetCreature("up to one target creature")),
		},
		Static: []game.StaticAbility{{
			Label: "Creatures your opponents control with -1/-1 counters on them can't block.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller && target.Counters[game.CounterMinusOne] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantBlock
			},
		}},
	})
}
