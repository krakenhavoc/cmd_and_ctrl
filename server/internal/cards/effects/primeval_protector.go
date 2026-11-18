package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primeval Protector — Creature — Avatar {10}{G}, 10/10:
//
//	"This spell costs {1} less to cast for each creature your opponents
//	 control.
//	 When this creature enters, put a +1/+1 counter on each other
//	 creature you control."
//
// Wall Off's self cost modifier (it lowers only the generic part, CR
// 601.2f) and Dawnstrike Vanguard's counters on each creature but the
// source, read as the trigger resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9581345-87dc-4fbd-acc1-bb629f2cf40f",
		Name:         "Primeval Protector",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(And(Creature(), OpponentControls())),
				"This spell costs {1} less to cast for each creature your opponents control."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Primeval Protector — put a +1/+1 counter on each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					return putPlusOneCountersOnEachYouControlThat(1,
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != source })(g, item)
				}),
		},
	})
}
