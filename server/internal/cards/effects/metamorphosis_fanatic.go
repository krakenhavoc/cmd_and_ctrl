package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metamorphosis Fanatic — Creature — Human Cleric {4}{B}{B}, 4/4 (Duskmourn
// Commander):
//
//	"Lifelink
//	 When this creature enters, return up to one target creature card from
//	 your graveyard to the battlefield with a lifelink counter on it.
//	 Miracle {1}{B} (...)"
//
// The return goes through the keyed graveyard return with a lifelink
// counter (CR 122.1b: a keyword counter), so the creature has lifelink for
// as long as the counter stays. Cast for its miracle cost the Fanatic
// enters from the stack like any cast creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "017aa9b3-a8ea-4588-9c50-e914a7d8e4ee",
		Name:             "Metamorphosis Fanatic",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{Miracle("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Metamorphosis Fanatic — return up to one target creature card from your graveyard to the battlefield with a lifelink counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ids := legalTargetCardIDs(NewContext(g, item))
					if len(ids) == 0 {
						return nil
					}
					if z := g.FindCardZoneForEffect(ids[0]); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return g.ReturnFromGraveyardWithCountersKeyedForEffect(ids[0], item.Controller, false,
						map[string]int{game.CounterLifelink: 1}, game.KeyedTail{})
				}),
				TargetCardInGraveyard("up to one target creature card from your graveyard", YouOwn(), Creature()).WithCount(0, 1)),
		},
	})
}
