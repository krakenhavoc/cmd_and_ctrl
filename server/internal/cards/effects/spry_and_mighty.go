package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spry and Mighty — Sorcery {4}{G} (Lorwyn Eclipsed, #4458):
//
//	"Choose exactly two creatures you control. You draw X cards and the
//	 chosen creatures get +X/+X and gain trample until end of turn, where X
//	 is the difference between the chosen creatures' powers."
//
// Nothing is targeted: the two are chosen as the spell resolves, and with
// fewer than two creatures the spell does nothing (the ruling). X is read
// from the powers when they are chosen, before the boost.
func init() {
	Register(Spec{
		OracleID:     "04cc9177-1ffc-4ac8-a64d-7995b4f33d23",
		Name:         "Spry and Mighty",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Answers: game.AnswerPump},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) < 2 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Spry and Mighty — choose exactly two creatures you control",
				Cards:    ids,
				Min:      2,
				Max:      2,
				LiveZone: game.ZoneBattlefield,
				ThenKey:  spryAndMightyChosen,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

var spryAndMightyChosen = askedPick("choose-cards/spry-and-mighty-two", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 2 {
		return nil
	}
	a, okA := ctx.Game.LookupCardForEffect(r.Picked[0])
	b, okB := ctx.Game.LookupCardForEffect(r.Picked[1])
	if !okA || !okB {
		return nil
	}
	x := a.CurrentPower() - b.CurrentPower()
	if x < 0 {
		x = -x
	}
	if err := (DrawCards{Player: ctx.Controller(), N: x}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range r.Picked {
		if err := (BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Spry and Mighty — +X/+X until end of turn"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Spry and Mighty — trample until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
