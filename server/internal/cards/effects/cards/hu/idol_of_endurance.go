package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Idol of Endurance — Artifact {2}{W} (Core Set 2021, #5511):
//
//	"When this artifact enters, exile all creature cards with mana value 3 or
//	 less from your graveyard until this artifact leaves the battlefield.
//	 {1}{W}, {T}: Until end of turn, you may cast a creature spell from among
//	 cards exiled with this artifact without paying its mana cost."
//
// The exile is "until": each card is exiled linked to this object (CR 607.2a)
// and goes back to its owner's graveyard, the zone it came from, when the
// Idol leaves (CR 610.3, UntilReturn.ToGraveyard). An Idol already gone when
// the trigger resolves exiles nothing (CR 610.3b). The ability opens ONE
// free cast, until end of turn, over the creature cards exiled with this very
// object: a second Idol's exiles are not in it, and one cast spends it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b946317-8a66-41b7-b902-308784d69b90",
		Name:         "Idol of Endurance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Idol of Endurance — exile all creature cards with mana value 3 or less from your graveyard until this artifact leaves the battlefield",
				idolOfEnduranceExile),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Until end of turn, you may cast a creature spell from among cards exiled with this artifact without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Effect:  idolOfEnduranceGrant,
		}},
	})
}

// idolOfEnduranceExile exiles every creature card with mana value 3 or less in
// the controller's graveyard, linked to this Idol, and records their return.
func idolOfEnduranceExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok || g.UntilEventHappenedForEffect(game.UntilReturn{Leaves: ref}, item) {
		return nil
	}
	ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.IsCreature() && !c.IsToken() && c.ManaValue() <= 3
	})
	for _, id := range ids {
		if err := (effects.ExileTarget{Target: id, Linked: true, ThenTail: ctx.ItemTail(idolOfEnduranceHold)}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// idolOfEnduranceHold records that the card now in exile returns to its
// owner's graveyard when the Idol leaves the battlefield. An on-disk key:
// never renamed, never reused.
var idolOfEnduranceHold = effects.CardTail("tail/s82-old-b110-idol-of-endurance-hold", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.ExiledWith.ID == uuid.Nil {
			continue
		}
		ctx.Game.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   ctx.Controller(),
			SourceCardID: ctx.Source(),
			Label:        "Idol of Endurance — the exiled card returns to its owner's graveyard when Idol of Endurance leaves the battlefield",
			Leaves:       game.ObjectRef(c.ExiledWith),
			ToGraveyard:  true,
		})
	}
	return nil
})

// idolOfEnduranceGrant is the activated ability: one free cast of a creature
// spell from among the cards exiled with this Idol, until end of turn.
func idolOfEnduranceGrant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cards []game.Card
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			cards = append(cards, *c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Cost:       "{0}",
		CastOnly:   true,
		CastsLeft:  1,
		Duration:   g.UntilEndOfTurnDuration(),
		Source:     ctx.Source(),
		SourceName: "Idol of Endurance",
		Label:      "Cast a creature spell from among the cards exiled with Idol of Endurance without paying its mana cost",
	}, cards)
	return nil
}
