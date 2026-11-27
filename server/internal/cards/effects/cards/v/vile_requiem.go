package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vile Requiem — Enchantment {2}{B}{B} (Commander 2013, S82 older sets):
//
//	"At the beginning of your upkeep, you may put a verse counter on this
//	 enchantment.
//	 {1}{B}, Sacrifice this enchantment: Destroy up to X target nonblack
//	 creatures, where X is the number of verse counters on this enchantment.
//	 They can't be regenerated."
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). X is
// fixed as the ability is announced, while the enchantment is still on the
// battlefield (CR 601.2b, 602.2b; TargetSpec.MaxFrom), so the sacrifice
// that is part of the cost doesn't change it. Each target is judged again as
// the ability resolves (CR 608.2b) and destroyed with regeneration
// forbidden.
//
// No simplification.
var vileRequiemCount = game.RegisterDivideAmount("s82-old-b381/vile-requiem-verse-counters",
	func(g *game.Game, a game.DivideAmountArgs) game.DivideSpec {
		c, ok := g.LookupCardForEffect(a.Source)
		if !ok {
			return game.DivideSpec{}
		}
		return game.DivideSpec{Total: c.Counters["verse"]}
	})

func init() {
	targets := effects.TargetCreature("up to X target nonblack creatures, where X is the number of verse counters on this enchantment", effects.NonBlack()).WithCount(0, 0)
	targets.MaxFrom = vileRequiemCount
	effects.Register(effects.Spec{
		OracleID:     "a72213e7-16f0-4fbe-91d3-750ca51a4336",
		Name:         "Vile Requiem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Vile Requiem — put a verse counter on this enchantment", effects.PutACounterOnThis("verse")),
				"Vile Requiem — put a verse counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this enchantment: Destroy up to X target nonblack creatures, where X is the number of verse counters on this enchantment. They can't be regenerated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			Targets: targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (effects.DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
