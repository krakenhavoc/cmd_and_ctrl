package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Herald — Creature — Human Shaman {2}{R}, 3/2:
//
//	"Haste
//	 When this creature enters, return any number of Aura cards from your
//	 graveyard to the battlefield attached to creatures you control. Exile
//	 those Auras at the beginning of your next end step. If those Auras would
//	 leave the battlefield, exile them instead of putting them anywhere else."
//
// The Auras are chosen on resolution (no targets), return together, and each
// chooses a host among the creatures you control as it enters (CR 303.4f,
// ZoneEntryOptions.AuraHostsYouControlOnly); one with no such creature stays
// in the graveyard (CR 303.4g). Both exile clauses apply to the Auras that
// did enter.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1692602a-0c44-485c-aadc-c9c70ed70ea8",
		Name:            "Storm Herald",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Storm Herald — return any number of Aura cards from your graveyard to the battlefield attached to creatures you control", stormHeraldAsk),
		},
	})
}

// stormHeraldAsk offers the Aura cards in the controller's graveyard.
func stormHeraldAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := ctx.Controller()
	cands := effects.GraveyardCardIDs(ctx, you, func(c game.Card) bool { return c.IsAura() })
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Storm Herald — return any number of Aura cards from your graveyard to the battlefield attached to creatures you control",
		Cards:      cands,
		Min:        0,
		Max:        len(cands),
		Zone:       game.ZoneGraveyard,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    stormHeraldPick,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// stormHeraldPick is the chosen set, an on-disk key: never renamed, never
// reused.
var stormHeraldPick = effects.AskedPick("choose-cards/storm-herald-auras", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, r.Picked, game.ZoneEntryOptions{
		AuraHostsCreaturesOnly:  true,
		AuraHostsYouControlOnly: true,
	}, stormHeraldEntered, 0)
})

// stormHeraldEntered exiles the Auras that entered at the beginning of your
// next end step, and exiles them instead of letting them go anywhere else. An
// on-disk key.
var stormHeraldEntered = effects.CardTail("storm-herald/entered", func(ctx *effects.Context, o game.TailOutcome) error {
	var ids []uuid.UUID
	for _, id := range o.IDs {
		if effects.OnBattlefield(ctx.Game, id) {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	if err := (effects.ScheduleDelayedTrigger{
		Label:              "Storm Herald — exile those Auras",
		Cards:              ids,
		Body:               effects.ExileListedCardsBody,
		ControllerTurnOnly: true,
	}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range ids {
		effects.ExileInsteadOfLeavingBattlefield(ctx.Game, id, ctx.Controller(),
			"Storm Herald: if that Aura would leave the battlefield, exile it instead")
	}
	return nil
})
