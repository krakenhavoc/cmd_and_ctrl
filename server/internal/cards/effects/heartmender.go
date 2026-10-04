package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartmender — Creature — Elemental {2}{G/W}{G/W}, 2/2:
//
//	"At the beginning of your upkeep, remove a -1/-1 counter from each
//	 creature you control.
//	 Persist"
//
// Each creature you control as the trigger resolves loses one -1/-1
// counter, itself included, so a persisted Heartmender resets and can
// persist again. Persist is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab4cf0ba-617f-4ef6-a831-eb3b41b4ab34",
		Name:            "Heartmender",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Heartmender — remove a -1/-1 counter from each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || !c.IsCreature() || c.Counters[game.CounterMinusOne] <= 0 {
							continue
						}
						if err := g.AddCounterForEffect(c.InstanceID, game.CounterMinusOne, -1); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
