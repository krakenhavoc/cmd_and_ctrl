package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ethereal Forager — Creature — Elemental Whale 3/3 {4}{U}{U}:
//
//	"Delve
//	 Flying
//	 Whenever this creature attacks, you may return an instant or
//	 sorcery card exiled with this creature to its owner's hand."
//
// ADR 0100 sub-PR 2. "Exiled with this creature" is CR 607.2q's link to
// the cards delve exiled to pay for the spell that became it. The
// trigger reads the link off the permanent as the ability RESOLVES,
// through ctx.SourcePermanent (CR 608.2h): the live Forager while it is
// still there, and its last-known information once it has left — the
// 2020-04-17 ruling that "that ability can still find the cards exiled
// with Ethereal Forager's delve ability" if the Forager leaves while
// the ability waits. Game.DelvedCardsForEffect keeps only the cards
// still in exile as the objects delve put there, so a card an earlier
// attack returned, or one that left and came back, is not offered
// (CR 400.7).
//
// "You may return an instant or sorcery card" is a choice of up to one
// among those cards, made as the ability resolves: a pick with a floor
// of zero, so choosing none is the "may". Not a target.
//
// No simplification.

const etherealForagerLabel = "Ethereal Forager — return an instant or sorcery card exiled with it"

func init() {
	Register(Spec{
		OracleID:        "a3f3a5b5-b931-4961-828c-501a33e5a0a0",
		Name:            "Ethereal Forager",
		Completeness:    CompletenessFull,
		Delve:           true,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(etherealForagerLabel, etherealForagerReturn),
		},
	})
}

// etherealForagerReturn offers the instant and sorcery cards still
// linked to the Forager's delve, and returns the one chosen, if any,
// to its owner's hand.
func etherealForagerReturn(g *game.Game, item *game.StackItem) error {
	info, ok := NewContext(g, item).SourcePermanent()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.DelvedCardsForEffect(info.Delved) {
		if b08IsInstantOrSorceryCard(c) {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Ethereal Forager — you may return an instant or sorcery card exiled with it to its owner's hand",
		Cards:    ids,
		Zone:     game.ZoneExile,
		Min:      0,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return g.BounceToHandForEffect(picked[0])
		},
	})
	return nil
}
