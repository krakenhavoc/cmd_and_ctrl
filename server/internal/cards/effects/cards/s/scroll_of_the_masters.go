package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scroll of the Masters — {2} Artifact (Fate Reforged):
//
//	"Whenever you cast a noncreature spell, put a lore counter on this
//	 artifact.
//	 {3}, {T}: Target creature you control gets +1/+1 until end of turn for
//	 each lore counter on this artifact."
//
// The count is read as the ability resolves; an artifact that has left the
// battlefield is read from its last-known counters (CR 608.2h).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2c7adde-492b-4585-8385-42d7e0d6e914",
		Name:         "Scroll of the Masters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Scroll of the Masters — put a lore counter on this artifact",
				effects.Do(effects.CounterOnThis{Kind: game.CounterLore, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature you control gets +1/+1 until end of turn for each lore counter on this artifact.",
			Purpose: game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0)), Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				n := 0
				if src, ok := g.LookupCardForEffect(item.SourceCardID); ok {
					n = src.Counters[game.CounterLore]
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: n, Toughness: n, Label: "Scroll of the Masters — +1/+1 for each lore counter"}.Apply(ctx)
			},
		}},
	})
}
