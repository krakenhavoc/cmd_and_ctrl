package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Moment — Legendary Artifact {2} (Doctor Who, #5948):
//
//	"At the beginning of your upkeep, put a time counter on The Moment.
//	 {2}, {T}: Untap target creature you control. It phases out until The
//	 Moment leaves the battlefield.
//	 {3}, {T}: Destroy each nonland permanent with mana value less than or
//	 equal to the number of time counters on The Moment. Then sacrifice The
//	 Moment. Activate only as a sorcery."
//
// The phase-out is Oubliette's duration (PhaseOutUntilLeaves, CR 702.26):
// the creature comes back the moment The Moment leaves, not at an untap step.
// The sweep reads the counters as the ability resolves (CR 608.2h) and is one
// destroy over the class, then the sacrifice as a second instruction (so The
// Moment, at mana value 2, is destroyed by its own sweep from two counters on
// and the sacrifice finds it gone). Regeneration and indestructible apply to
// the destroy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a5259cc-224d-4748-b764-63ad0e4e02f1",
		Name:         "The Moment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("The Moment — put a time counter on The Moment", func(g *game.Game, item *game.StackItem) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterTime, N: 1}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Untap target creature you control. It phases out until The Moment leaves the battlefield.",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ids := effects.LegalTargetCards(item, g)
					for _, id := range ids {
						if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.PhaseOutUntilLeaves{Targets: ids, Until: item.SourceCardID}.Apply(ctx)
				},
			},
			{
				Label:        "{3}, {T}: Destroy each nonland permanent with mana value less than or equal to the number of time counters on The Moment. Then sacrifice The Moment. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := 0
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						n = c.Counters[game.CounterTime]
					}
					return effects.DestroyAllMatching{Match: effects.And(effects.Not(effects.Land()), effects.ManaValueLE(n)), ThenTail: ctx.ItemTail(theMomentThenSacrifice)}.Apply(ctx)
				},
			},
		},
	})
}

// theMomentThenSacrifice is "Then sacrifice The Moment", run once the sweep
// has settled: if The Moment is still on the battlefield it is sacrificed. An
// on-disk key: never renamed, never reused.
var theMomentThenSacrifice = effects.MassTail("tail/the-moment-then-sacrifice", func(ctx *effects.Context, _ []game.Card, _ int) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})
