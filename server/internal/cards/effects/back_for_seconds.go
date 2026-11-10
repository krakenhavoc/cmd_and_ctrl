package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Back for Seconds — Sorcery {2}{B} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Return up to two target creature cards from your graveyard to your
//	 hand. If this spell was bargained, you may put one of those cards with
//	 mana value 4 or less onto the battlefield instead of putting it into
//	 your hand."
//
// Bargain is the shared optional cost (g2Bargain). An unbargained spell
// returns every target still legal. A bargained one first offers the
// targets with mana value 4 or less (at most one, and the ruling lets the
// player decline), puts the pick onto the battlefield under your control
// and returns the rest to hand. The pick is keyed (an on-disk identity)
// and reads the targets back off the asking item.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "ee3d21b9-58b4-480c-bae3-26de50f1a5c6",
		Name:          "Back for Seconds",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerValue},
		OptionalCosts: []game.AdditionalCost{g2Bargain()},
		Targets:       TargetCardInGraveyard("up to two target creature cards from your graveyard", YouOwn(), Creature()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if !g2WasBargained(ctx) {
				return returnLegalGraveyardTargetsToHand(ctx)
			}
			var candidates []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.ManaValue() <= 4 {
					candidates = append(candidates, t.ID)
				}
			}
			if len(candidates) == 0 {
				return returnLegalGraveyardTargetsToHand(ctx)
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Back for Seconds — you may put one of those cards with mana value 4 or less onto the battlefield instead of into your hand",
				Cards:    candidates,
				Min:      0,
				Max:      1,
				LiveZone: game.ZoneGraveyard,
				ThenKey:  backForSecondsPick,
				Asked:    game.NewAskedItem(ctx.Item),
			})
			return nil
		},
	})
}

// backForSecondsPick puts the chosen card onto the battlefield and the other
// legal targets into your hand. An on-disk key: never renamed, never reused.
var backForSecondsPick = askedPick("choose-cards/back-for-seconds-pick", func(ctx *Context, r game.CardsChosen) error {
	var onto uuid.UUID
	if len(r.Picked) > 0 {
		onto = r.Picked[0]
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if t.ID == onto {
			if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
