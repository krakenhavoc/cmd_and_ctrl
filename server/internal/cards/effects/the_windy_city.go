package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Windy City — Plane — Chicago (Unknown Event, #4870):
//
//	"When you planeswalk here and at the beginning of your upkeep, put a
//	 flying counter on target nontoken creature you control without flying.
//	 Whenever chaos ensues, exile target creature with flying. For as long
//	 as it remains exiled, its owner may cast it by paying {2} rather than
//	 paying its mana cost.
//	 When you planeswalk away from The Windy City, remove all flying
//	 counters from all creatures."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first ability is ONE trigger with two conditions
// (WhenYouPlaneswalkToThisAndAtYourUpkeep); a flying counter is the keyword
// counter the layer pass reads (CR 122.1b). The chaos ability exiles the
// creature and grants its OWNER, whoever controlled it, a cast-only
// permission for as long as it stays exiled, charging {2} instead of its
// mana cost (an alternative cost, so {2} is paid even when the printed cost
// is less). A token exiled this way ceases to exist (CR 111.7). The last
// ability looks back at the plane that has just left (CR 901.11b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04491781-0e6a-4a30-a69f-db226702adfc",
		Name:         "The Windy City",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenYouPlaneswalkToThisAndAtYourUpkeep("The Windy City — put a flying counter on target nontoken creature you control without flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterFlying, N: 1}.Apply(ctx)
				}),
				TargetCreature("target nontoken creature you control without flying",
					YouControl(), Not(IsTokenPredicate()), WithoutKeyword("flying"))),
			Targeting(WheneverChaosEnsues("The Windy City — exile target creature with flying; its owner may cast it for {2} while it remains exiled",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					card, ok := g.LookupCardForEffect(id)
					if !ok {
						return nil
					}
					return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
						Player:   card.Owner,
						Cost:     "{2}",
						Duration: game.WhileInZoneDuration(),
						CastOnly: true,
						Source:   item.SourceCardID,
					})
				}),
				TargetCreature("target creature with flying", HasKeyword("flying"))),
			WhenYouPlaneswalkAwayFromThis("The Windy City — remove all flying counters from all creatures",
				func(g *game.Game, item *game.StackItem) error {
					for _, c := range g.BattlefieldCardsForEffect() {
						if n := c.Counters[game.CounterFlying]; n > 0 && c.IsCreature() {
							if err := g.AddCounterByForEffect(item.Controller, c.InstanceID, game.CounterFlying, -n); err != nil {
								return err
							}
						}
					}
					return nil
				}),
		},
	})
}
