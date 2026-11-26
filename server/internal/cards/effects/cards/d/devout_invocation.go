package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devout Invocation — Sorcery {6}{W} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"Tap any number of untapped creatures you control. Create a 4/4 white Angel
//	 creature token with flying for each creature tapped this way."
//
// Yotia Declares War's pick: the creatures are chosen as the spell resolves, any
// number, none included; each pick is re-read before it is tapped, so a creature
// that left, was tapped or was stolen in the meantime is not "tapped this way"
// and makes no Angel.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fab491b0-b820-4027-996f-390c97df1e96",
		Name:         "Devout Invocation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var candidates []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() && !c.Tapped {
					candidates = append(candidates, c.InstanceID)
				}
			}
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Devout Invocation — tap any number of untapped creatures you control",
				Cards:    candidates,
				Min:      0,
				Max:      0, // "any number"
				LiveZone: game.ZoneBattlefield,
				ThenKey:  devoutInvocationPick,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// devoutInvocationPick taps the picks that are still untapped creatures its
// controller controls and creates one Angel for each one tapped.
var devoutInvocationPick = effects.AskedPick("choose-cards/devout-invocation-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	tapped := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Item.Controller {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		tapped++
	}
	if tapped == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("4/4 white Angel with flying"), N: tapped}.Apply(ctx)
})
