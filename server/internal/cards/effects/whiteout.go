package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiteout — Snow Instant {1}{G}:
//
//	"All creatures lose flying until end of turn.
//	 Sacrifice a snow land: Return this card from your graveyard to your
//	 hand."
//
// The first line affects the creatures on the battlefield as it resolves (CR
// 611.2c); each is a layer 6 removal with its own timestamp, so a flying
// grant made later still takes hold. The second line is an activated ability
// that works from the graveyard (Blessed Ghoul's body) paid by sacrificing a
// snow land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e8de3e1-1159-460f-99b7-1043e5814597",
		Name:         "Whiteout",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !c.IsCreature() {
					continue
				}
				if err := untilEndOfTurn(ctx, c.InstanceID, nil, "Whiteout — loses flying", game.RemoveKeywordsMod("flying")); err != nil {
					return err
				}
			}
			return nil
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a snow land: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    SacrificeN(1, "a snow land", Land(), Snow()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
