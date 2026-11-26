package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning Reaver — Creature — Zombie Beast {3}{B}{R}, 3/3 (Alara Reborn,
// #5586):
//
//	"Haste; fear
//	 Whenever this creature deals combat damage to a player, put a charge
//	 counter on it.
//	 At the beginning of your end step, this creature deals damage equal to the
//	 number of charge counters on it to each opponent."
//
// Haste and fear are printed keywords. The counter is put on only while the
// Reaver is still the permanent that dealt the damage. At the end step the
// damage is dealt to every opponent at once, the amount being the charge counters
// the Reaver has as the ability resolves, or had when it last was on the
// battlefield if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "687abe1c-b2fc-49e0-bd36-3db3dc3a5039",
		Name:            "Lightning Reaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "fear"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Lightning Reaver — put a charge counter on it",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.AtYourEndStep("Lightning Reaver — it deals damage equal to the number of charge counters on it to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					n := s82OldB503ChargeCounters(g, item.SourceCardID)
					if n <= 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return g.DamageInstanceForEffect(func() error {
						for _, opp := range ctx.Opponents() {
							if err := (effects.DealDamage{Source: ctx.Source(), Target: opp, Amount: n}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}),
		},
	})
}

// s82OldB503ChargeCounters is the charge counters on the permanent: as it is
// now while it is on the battlefield, as it last was once it has left.
func s82OldB503ChargeCounters(g *game.Game, id uuid.UUID) int {
	if c, ok := effects.BattlefieldCardForEffect(g, id); ok {
		return c.Counters[game.CounterCharge]
	}
	if info, ok := g.LastKnownPermanentForEffect(id); ok {
		return info.Counters[game.CounterCharge]
	}
	return 0
}
