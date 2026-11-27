package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skullmane Baku — Creature — Spirit {3}{B}{B}, 2/1 (Betrayers of Kamigawa,
// #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a ki counter on
//	 this creature.
//	 {1}, {T}, Remove X ki counters from this creature: Target creature gets
//	 -X/-X until end of turn."
//
// Quillmane Baku's ki counter trigger and cost. X is the number of ki
// counters the cost removes, announced with the activation (CR 602.2b), so a
// response cannot spend the same counter twice; the target is a creature, and
// it gets -X/-X until end of turn if it is still legal as the ability
// resolves (CR 608.2b). X may be 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74466fa6-34d9-42f9-8590-c7f1fa672bf5",
		Name:         "Skullmane Baku",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Or(effects.Subtype("Spirit"), effects.Subtype("Arcane")),
				"Skullmane Baku — put a ki counter on this creature", effects.PutACounterOnThis("ki")),
				"Skullmane Baku — put a ki counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Remove X ki counters from this creature: Target creature gets -X/-X until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RemoveCountersXFromThis("ki", 0)),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetHarm}),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.CountersRemoved()
				for _, t := range ctx.LegalTargets() {
					return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x, Label: "Skullmane Baku — -X/-X"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
