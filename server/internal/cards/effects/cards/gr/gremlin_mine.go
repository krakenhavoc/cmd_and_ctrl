package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gremlin Mine — Artifact {1} (New Phyrexia, #5575):
//
//	"{1}, {T}, Sacrifice this artifact: It deals 4 damage to target artifact
//	 creature.
//	 {1}, {T}, Sacrifice this artifact: Remove up to four charge counters from
//	 target noncreature artifact."
//
// Two activated abilities, each paying the sacrifice as a cost (CR 602.2), so
// both resolve from last-known information (CR 608.2h) once the Mine has gone:
// the first deals 4 damage from it to an artifact creature, the second lets the
// controller choose how many charge counters, up to four, come off a noncreature
// artifact. A target that became illegal means the ability does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33510024-b1ff-428a-8ccd-2ca2a332d53c",
		Name:         "Gremlin Mine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}, Sacrifice this artifact: It deals 4 damage to target artifact creature.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(4),
			},
			{
				Label:   "{1}, {T}, Sacrifice this artifact: Remove up to four charge counters from target noncreature artifact.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetPermanent("target noncreature artifact", effects.Artifact(), effects.Noncreature()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						c, ok := g.LookupCardForEffect(t.ID)
						if !ok || c.Counters[game.CounterCharge] <= 0 {
							return nil
						}
						most := min(4, c.Counters[game.CounterCharge])
						return effects.ChooseNumber{
							Min:      0,
							Max:      most,
							Question: "Gremlin Mine — remove up to four charge counters",
							Unit:     game.PayAmountCounters,
							Goal: func(ctx *effects.Context) int {
								// The bot strips an opponent's artifact and leaves its own alone.
								if o, ok := ctx.Game.LookupCardForEffect(t.ID); ok && o.Controller != ctx.Controller() {
									return most
								}
								return 0
							},
							Then:  gremlinMineRemoved,
							Carry: []uuid.UUID{t.ID},
						}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}

// gremlinMineRemoved is the number chosen: that many charge counters come off the
// artifact. The artifact rides the prompt's Carry. An on-disk key.
var gremlinMineRemoved = effects.NumberThen("number/gremlin-mine-remove-charge-counters",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		if len(r.Carry) == 0 || r.Number <= 0 {
			return nil
		}
		return effects.AddCounter{Target: r.Carry[0], Kind: game.CounterCharge, N: -r.Number}.Apply(ctx)
	})
