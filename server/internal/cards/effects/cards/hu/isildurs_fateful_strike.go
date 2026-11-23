package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isildur's Fateful Strike — Legendary Instant {2}{B}{B} (The Lord of the
// Rings: Tales of Middle-earth, #5948):
//
//	"(You may cast a legendary instant only if you control a legendary
//	 creature or planeswalker.)
//	 Destroy target creature. If its controller has more than four cards in
//	 hand, they exile cards from their hand equal to the difference."
//
// The cast restriction is CR 205.4e's, shared with the legendary sorceries and
// checked at announce only. The two sentences are two steps of one
// resolution, so the second waits for a prompt the first asks (CR 608.2c).
// "Its controller" is read before the destroy, since the card is in its
// owner's graveyard afterwards (CR 608.2h). The clause does not depend on the
// creature actually being destroyed: a regenerated or indestructible creature's
// controller still exiles down to four. The controller, not the caster,
// chooses the cards (a prompt for that player over their own hand, exactly
// the excess, CR 701.13), and the count is read after the destroy has
// settled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "a43e98d1-684a-4198-bc7e-877b9f6d5357",
		Name:               "Isildur's Fateful Strike",
		Completeness:       effects.CompletenessFull,
		CastCondition:      effects.LegendarySorcery(),
		CastConditionLabel: effects.LegendarySorceryLabel,
		Purpose:            game.Purpose{Answers: game.AnswerRemove},
		Targets:            effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				targets := ctx.LegalTargets()
				if len(targets) == 0 || targets[0].Kind != game.TargetCard {
					return nil
				}
				if c, ok := ctx.Game.LookupCardForEffect(targets[0].ID); ok {
					ctx.RecordStep(isildurFatefulStrikeRecord, []uuid.UUID{c.Controller})
				}
				return effects.DestroyTarget{Target: targets[0].ID}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(isildurFatefulStrikeRecord)
				if len(r.IDs) == 0 {
					return nil
				}
				p := ctx.Game.PlayerByIDForEffect(r.IDs[0])
				if p == nil || p.Hand == nil || p.Hand.Size() <= 4 {
					return nil
				}
				excess := p.Hand.Size() - 4
				cards := make([]uuid.UUID, 0, p.Hand.Size())
				for _, c := range p.Hand.Cards {
					cards = append(cards, c.InstanceID)
				}
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  p.ID,
					Source:   ctx.Item.SourceCardID,
					Question: "Isildur's Fateful Strike — exile cards from your hand equal to the number over four",
					Cards:    cards,
					Min:      excess,
					Max:      excess,
					LiveZone: game.ZoneHand,
					ThenKey:  isildurFatefulStrikeExile,
					Asked:    game.NewAskedItem(ctx.Item),
				})
				return nil
			}),
		),
	})
}

const isildurFatefulStrikeRecord = "isildurs-fateful-strike/controller"

// isildurFatefulStrikeExile exiles the chosen cards from the hand. An
// on-disk key: never renamed, never reused.
var isildurFatefulStrikeExile = effects.AskedPick("choose-cards/isildurs-fateful-strike-exile", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
