package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neyam Shai Murad — Legendary Creature — Human Rogue {2}{W}{B}, 3/3:
//
//	"Rogue Trader — Whenever Neyam Shai Murad deals combat damage to a
//	 player, you may have that player return target permanent card from their
//	 graveyard to their hand. If you do, that player chooses a permanent card
//	 in your graveyard, then you put it onto the battlefield under your
//	 control."
//
// The target is a permanent card in the damaged player's graveyard (the
// clause is bound to the event's player, owner-judged for a graveyard card),
// chosen when the trigger is put on the stack; "you may" is asked as it
// resolves. "If you do" is the return actually happening: a card that left the
// graveyard in response is a fizzled clause and nothing follows. Then THAT
// player picks among the permanent cards in YOUR graveyard (the card just
// returned is theirs, so it is never a candidate), and you put the pick
// onto the battlefield under your control. With no candidate nothing is asked.
func init() {
	t := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		Key:     "Neyam Shai Murad — you may have that player return target permanent card from their graveyard to their hand",
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
		},
		Targets: effects.ControlledByThatPlayer(effects.TargetCardInGraveyard("target permanent card from that player's graveyard",
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return effects.IsPermanentCard(c) })),
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, "Neyam Shai Murad — that player returns a permanent card; then you take one of yours")
			item.Params.Player = ev.Target
			return item
		},
		Effect: neyamShaiMuradResolve,
	}
	effects.Register(effects.Spec{
		OracleID:     "9466cfc1-a8fe-44a0-abd2-5d2dfe7110e0",
		Name:         "Neyam Shai Murad",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Optional(t, "Have that player return a permanent card from their graveyard to their hand?")},
	})
}

func neyamShaiMuradResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	// "If you do": the card is now in a hand.
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	var candidates []uuid.UUID
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if effects.IsPermanentCard(c) {
				candidates = append(candidates, c.InstanceID)
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Params.Player,
		FromPlayer: item.Controller,
		Source:     ctx.Source(),
		Question:   "Neyam Shai Murad — choose a permanent card in their graveyard",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    neyamShaiMuradPick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

var neyamShaiMuradPick = effects.AskedPick("choose-cards/neyam-shai-murad-pick", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
	}
	return nil
})
