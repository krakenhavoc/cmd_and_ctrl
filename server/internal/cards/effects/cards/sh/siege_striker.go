package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siege Striker — Creature — Human Soldier {2}{W}, 1/1 (Core Set 2021,
// #5511):
//
//	"Double strike
//	 Whenever this creature attacks, you may tap any number of untapped
//	 creatures you control. This creature gets +1/+1 until end of turn for
//	 each creature tapped this way."
//
// The pick is "any number" over the controller's untapped creatures, floor
// zero, which is the "you may". Each pick is re-read as it is tapped (the
// board can move under the prompt): only a creature that really became tapped
// counts. The boost goes on the Striker if it is still the permanent that
// attacked (CR 400.7), and nothing happens for zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12dd664c-a279-4adc-b6cb-bec22eedca5c",
		Name:            "Siege Striker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Siege Striker — you may tap any number of untapped creatures you control; +1/+1 for each", siegeStrikerAsk),
		},
	})
}

// siegeStrikerAsk offers the controller's untapped creatures.
func siegeStrikerAsk(g *game.Game, item *game.StackItem) error {
	cands := effects.UntappedCreaturesControlledBy(g, item.Controller)
	if len(cands) == 0 || !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Siege Striker — you may tap any number of untapped creatures you control",
		Cards:    cands,
		Min:      0,
		Max:      0, // zero means every candidate: "any number"
		LiveZone: game.ZoneBattlefield,
		ThenKey:  siegeStrikerTap,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// siegeStrikerTap taps the picks and gives the Striker +1/+1 for each one
// that really became tapped. An on-disk key: never renamed, never reused.
var siegeStrikerTap = effects.AskedPick("choose-cards/s82-old-b106-siege-striker-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	tapped := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		tapped++
	}
	if tapped == 0 || !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.BoostUntilEOT{Target: ctx.Source(), Power: tapped, Toughness: tapped, Label: "Siege Striker — +1/+1 for each creature tapped this way"}.Apply(ctx)
})
