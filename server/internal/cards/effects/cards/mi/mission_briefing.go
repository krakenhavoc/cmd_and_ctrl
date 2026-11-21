package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mission Briefing — Instant {U}{U}:
//
//	"Surveil 2, then choose an instant or sorcery card in your graveyard.
//	 You may cast it this turn. If that spell would be put into your
//	 graveyard, exile it instead."
//
// The choice is made after the surveil settles (Surveil.ThenTail), so a card
// the surveil just binned is a candidate; the chosen card is not cast by the
// resolution and not for free: a permission lasting the turn keeps the card's
// own timing (a sorcery waits for your main phase) and charges its mana cost,
// and sends the spell to exile instead of a graveyard however it leaves the
// stack (ExileInsteadOfGraveyard). The choice is not optional while a
// candidate exists ("choose"), and the cast is.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3a68018-9eff-47e6-a612-182886d28fe3",
		Name:         "Mission Briefing",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Surveil{Player: item.Controller, N: 2, ThenTail: game.KeyedTail{Ref: missionBriefingChoose}}.Apply(ctx)
		},
	})
}

// missionBriefingChoose is the clause after the surveil. An on-disk key.
var missionBriefingChoose = effects.CardTail("mission-briefing/choose", func(ctx *effects.Context, _ game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	cards := effects.GraveyardCardIDs(ctx, you, func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    you,
		FromPlayer: you,
		Source:     ctx.Source(),
		Question:   "Mission Briefing — choose an instant or sorcery card in your graveyard; you may cast it this turn",
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    missionBriefingChosen,
	})
	return nil
})

// missionBriefingChosen opens the cast over the chosen card, if it is still in
// the graveyard. An on-disk key.
var missionBriefingChosen = game.RegisterChooseCardsThen("choose-cards/mission-briefing-chosen", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	cmdr2S42GrantGraveyardCastThisTurn(g, r.Chooser, r.Source, "Mission Briefing", r.Picked[0])
	return nil
})

// cmdr2S42GrantGraveyardCastThisTurn is "you may cast it this turn" over a
// card in the chooser's graveyard: paid at its mana cost, at its own timing,
// and exiled instead of going to a graveyard.
func cmdr2S42GrantGraveyardCastThisTurn(g *game.Game, player, source uuid.UUID, name string, cardID uuid.UUID) {
	z := g.FindCardZoneForEffect(cardID)
	c, ok := g.LookupCardForEffect(cardID)
	if !ok || z == nil || z.Kind != game.ZoneGraveyard {
		return
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:                  player,
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		AltCostKey:              game.AltCostKeyGrantedCost,
		CastOnly:                true,
		ExileInsteadOfGraveyard: true,
		Source:                  source,
		SourceName:              name,
		Label:                   name + " — cast " + c.Name + " from your graveyard this turn",
	}, []game.Card{c})
}
