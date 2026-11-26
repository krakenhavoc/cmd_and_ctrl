package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brain Maggot — Enchantment Creature — Insect {1}{B}, 1/1 (Journey into Nyx,
// S82 OLD, #5558):
//
//	"When this creature enters, target opponent reveals their hand and you
//	 choose a nonland card from it. Exile that card until this creature
//	 leaves the battlefield."
//
// Kitesail Freebooter's reveal, pick and linked exile with "nonland" for
// "noncreature, nonland". The card returns to its owner's hand when the Maggot
// object leaves the battlefield (CR 610.3); a hand with no nonland card is
// revealed and nothing is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "419b5b4f-cac7-44bf-b1e2-655b5a5d37b1",
		Name:         "Brain Maggot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Brain Maggot — target opponent reveals their hand; you choose a nonland card from it and exile it until this creature leaves the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.ChooseFromRevealedHand{
							Player: effects.TargetedPlayer(ctx),
							Filter: effects.Nonland(),
							Label:  "nonland card",
							Exile:  true,
							Linked: true,
							Then:   brainMaggotReturn,
						}.Apply(ctx)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// brainMaggotReturn records that the exiled card goes back to its owner's
// hand when the Maggot leaves the battlefield. An on-disk key: never renamed,
// never reused.
var brainMaggotReturn = effects.RevealedPickThen("revealed-pick/s82-old-b369-brain-maggot-return",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		id := pick.Chosen[0].InstanceID
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.ExiledWith.ID == uuid.Nil {
			return nil
		}
		ctx.Game.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   pick.Chooser,
			SourceCardID: pick.Source,
			Label:        "Brain Maggot — the exiled card returns to its owner's hand",
			Leaves:       game.ObjectRef(c.ExiledWith),
			ToHand:       true,
		})
		return nil
	})
