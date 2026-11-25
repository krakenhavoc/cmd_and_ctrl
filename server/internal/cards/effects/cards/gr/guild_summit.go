package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guild Summit — Enchantment {2}{U} (Guilds of Ravnica):
//
//	"When this enchantment enters, you may tap any number of untapped Gates
//	 you control. Draw a card for each Gate tapped this way.
//	 Whenever a Gate you control enters, draw a card."
//
// The enters trigger is Minas Tirith Garrison's pick-and-tap over Gates
// (zero is a legal pick; tapping is not a cost, so a Gate that left or got
// tapped in the meantime is skipped and draws nothing), and the second
// ability is Gate Colossus's "a Gate you control enters" watcher.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38a3e144-93ad-497b-9b29-7040b6f337e1",
		Name:         "Guild Summit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Guild Summit — you may tap any number of untapped Gates you control; draw a card for each",
				guildSummitOffer),
			effects.On(game.EventETB, guildSummitGateEntered,
				"Guild Summit — a Gate you control entered; draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// guildSummitGateEntered is "whenever a Gate you control enters".
func guildSummitGateEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtypeWord("Gate")
}

// guildSummitOffer asks which untapped Gates to tap. With none to offer
// nothing is asked and nothing happens.
func guildSummitOffer(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && !c.Tapped && c.HasSubtypeWord("Gate") {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Guild Summit — tap any number of untapped Gates you control",
		Cards:    candidates,
		Min:      0,
		// Max 0 means "all of them", which is what "any number" is.
		Max:      0,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  guildSummitTapPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// guildSummitTapPick taps the picked Gates that are still untapped Gates its
// controller controls, then draws for each. An on-disk key.
var guildSummitTapPick = effects.AskedPick("choose-cards/guild-summit-tap-gates", func(ctx *effects.Context, r game.CardsChosen) error {
	tapped := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || c.Controller != ctx.Item.Controller || !c.HasSubtypeWord("Gate") {
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
	return effects.DrawCards{Player: ctx.Item.Controller, N: tapped}.Apply(ctx)
})
