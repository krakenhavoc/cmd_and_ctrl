package cardsmi

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mudslide — Enchantment {2}{R} (Ice Age, S82 OLD batch 848, #5636):
//
//	"Creatures without flying don't untap during their controllers' untap steps.
//	 At the beginning of each player's upkeep, that player may choose any number
//	 of tapped creatures without flying they control and pay {2} for each
//	 creature chosen this way. If the player does, untap those creatures."
//
// The first line is Rising Waters' untap-step restriction over every creature
// without flying (read through the layers). The second is one trigger for every
// upkeep: the player whose upkeep it is (CR 608.2d) picks any number of their
// tapped nonflying creatures, from none to all, and is then asked to pay {2} for
// each (MayPay over the picked count); only a payment untaps them, all of them
// that are still tapped creatures they control.
//
// No simplification.
func init() {
	nonflying := effects.And(effects.Creature(), effects.WithoutKeyword("flying"))
	effects.Register(effects.Spec{
		OracleID:              "ac1d7dfe-1454-46a9-acf6-66ae86b6a6c5",
		Name:                  "Mudslide",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(nonflying)},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Mudslide — that player may choose any number of tapped creatures without flying they control and pay {2} for each; if they do, untap those creatures",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil {
						return nil
					}
					var tapped []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == player && c.Tapped && nonflying(g, player, c) {
							tapped = append(tapped, c.InstanceID)
						}
					}
					if len(tapped) == 0 {
						return nil
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  player,
						Source:   item.SourceCardID,
						Question: "Mudslide — choose any number of tapped creatures without flying you control to untap ({2} for each)",
						Cards:    tapped,
						Min:      0,
						Max:      len(tapped),
						ThenKey:  mudslideChosen,
						LiveZone: game.ZoneBattlefield,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// mudslideChosen is the pick: pay {2} for each creature chosen, or nothing is
// untapped. An on-disk key: never renamed, never reused.
var mudslideChosen = effects.AskedPick("choose-cards/mudslide-untap", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.MayPay{
		Chooser:  r.Chooser,
		Cost:     strings.Repeat("{2}", len(r.Picked)),
		Question: "Mudslide — pay {2} for each chosen creature to untap them?",
		Then:     mudslideUntap,
		Carry:    r.Picked,
	}.Apply(ctx)
})

// mudslideUntap is the paid branch: untap the creatures chosen, each only while
// it is still on the battlefield. An on-disk key: never renamed, never reused.
var mudslideUntap = effects.PayUnlessBranches("pay-unless/mudslide-untap",
	func(ctx *effects.Context, ans game.PayUnlessAnswer) error {
		for _, id := range ans.Carry {
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}, nil)
