package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minas Tirith Garrison — Creature — Human Soldier {3}{U}, */5 (The Hobbit):
//
//	"Minas Tirith Garrison's power is equal to the number of cards in
//	 your hand.
//	 Whenever this creature attacks, you may tap any number of untapped
//	 Humans you control. Draw a card for each Human tapped this way."
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a),
// Psychosis Crawler's shape with the toughness left at the printed 5:
// DependsOnHandSize drops the layer cache whenever a hand changes, so the
// number is read live.
//
// The attack trigger is Teferi Akosa's "tap any number" prompt: the
// controller picks from the untapped Humans (a floor of zero, since "any
// number" includes none; Max 0 means all of them) and the continuation
// re-reads every pick before tapping it, because the prompt outlives the
// resolution. A Human that was tapped, removed or stolen by then was not
// "tapped this way" and draws nothing. The Garrison itself is tapped by
// attacking (no vigilance), so it is not offered; a vigilant one could tap
// itself, as printed. The cards are drawn once, for the count that really
// got tapped. Tapping is not a cost: nothing here depends on the draws
// being possible.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1cc49cd-5039-4c31-a895-4ebf869fc5e4",
		Name:         "Minas Tirith Garrison",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(powerIsCardsInYourHand(), "Minas Tirith Garrison's power is equal to the number of cards in your hand."),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Minas Tirith Garrison — you may tap any number of untapped Humans you control; draw a card for each",
				minasTirithGarrisonOffer),
		},
	})
}

// minasTirithGarrisonOffer asks which untapped Humans to tap. With none
// to offer nothing is asked and nothing happens.
func minasTirithGarrisonOffer(g *game.Game, item *game.StackItem) error {
	human := Subtype("Human")
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !c.Tapped && human(g, item.Controller, c) {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Minas Tirith Garrison — tap any number of untapped Humans you control",
		Cards:    candidates,
		Min:      0,
		// Max 0 means "all of them", which is what "any number" is.
		Max:      0,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  minasTirithGarrisonTapPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// minasTirithGarrisonTapPick taps the picked Humans that are still
// untapped Humans its controller controls, then draws for each.
var minasTirithGarrisonTapPick = askedPick("choose-cards/minas-tirith-garrison-tap-humans", func(ctx *Context, r game.CardsChosen) error {
	human := Subtype("Human")
	tapped := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Item.Controller || !human(ctx.Game, ctx.Item.Controller, c) {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		tapped++
	}
	if tapped == 0 {
		return nil
	}
	return DrawCards{Player: ctx.Item.Controller, N: tapped}.Apply(ctx)
})
