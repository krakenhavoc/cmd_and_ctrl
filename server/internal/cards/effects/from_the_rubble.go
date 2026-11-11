package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// From the Rubble — Enchantment {4}{W}{W}:
//
//	"As this enchantment enters, choose a creature type.
//	 At the beginning of your end step, return target creature card of the
//	 chosen type from your graveyard to the battlefield with a finality
//	 counter on it."
func init() {
	Register(Spec{
		OracleID:     "76d07c99-8c6b-4c2c-a0f6-47541999015b",
		Name:         "From the Rubble",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("From the Rubble"),
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtYourEndStep("From the Rubble — return target creature card of the chosen type from your graveyard to the battlefield with a finality counter on it",
					fromTheRubbleReturn),
				OfTheChosenType(TargetCardInGraveyard("target creature card of the chosen type from your graveyard", YouOwn(), Creature()))),
		},
	})
}

func fromTheRubbleReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.ReturnFromGraveyardWithCountersKeyedForEffect(t.ID, item.Controller, false,
			map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
	}
	return nil
}
