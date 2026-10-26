package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Plume Adventurer — Creature — Orc Cleric {2}{W}, 3/3:
//
//	"When this creature enters, you take the initiative.
//	 At the beginning of each opponent's upkeep, untap a creature you
//	 control. If you've completed a dungeon, untap all creatures you
//	 control instead."
//
// The ETB is TakeTheInitiative (ADR 0203, #4096). "Untap a creature you
// control" is a choice made as the trigger resolves, not a target: any
// creature you control may be chosen, hexproof or not, tapped or not
// (choosing an untapped one does nothing). "If you've completed a
// dungeon" is a clause of the effect, read as it resolves (CR 608.2), so
// a dungeon completed while the trigger waited untaps them all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "51f091a7-9b0b-4362-8e9c-c174752f369b",
		Name:         "White Plume Adventurer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("White Plume Adventurer"),
			AtEachOpponentsUpkeep("White Plume Adventurer — untap a creature you control, or all of them if you've completed a dungeon",
				whitePlumeUntap),
		},
	})
}

// whitePlumeUntap untaps one chosen creature, or every creature for a
// player who has completed a dungeon.
func whitePlumeUntap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if YouCompletedADungeon(g, item.Controller) {
		for _, id := range creaturesControlledByPlayer(g, item.Controller) {
			if err := (UntapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
	return ChoosePermanents{
		Question:   "White Plume Adventurer — untap a creature you control",
		Candidates: humiliateCreatures,
		Keyed:      whitePlumeChosen,
	}.Apply(ctx)
}

// whitePlumeChosen is untapping the chosen creature, keyed (#4290).
var whitePlumeChosen = RegisterPermanentsPick("prompt-run/white-plume-adventurer-untap", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
