package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shiko and Narset, Unified — Legendary Creature — Human Spirit Dragon
// {1}{U}{R}{W}, 4/4 (Tarkir: Dragonstorm Commander):
//
//	"Flying, vigilance
//	 Flurry — Whenever you cast your second spell each turn, copy that spell
//	 if it targets a permanent or player, and you may choose new targets for
//	 the copy. If you don't copy a spell this way, draw a card."
//
// The trigger is Cori Mountain Stalwart's flurry. Whether the spell targets a
// permanent or player is read when it is cast and kept on the item
// (Params.Amount, a fill-in Build), because by resolution the target may have
// died or the spell may be gone. A spell that does is copied from last-known
// information (CR 608.2h), naming the spell without targeting it, and the copy
// is not cast (CR 707.10); it resolves before the original. Any other spell, a
// graveyard-targeting one or one with no target, draws instead.
func init() {
	const label = "Shiko and Narset, Unified — flurry: copy that spell if it targets a permanent or player, otherwise draw a card"
	t := effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		if item.Params.Amount > 0 {
			return effects.CopySpell{StackID: item.Trigger.Event.CardID, Controller: item.Controller,
				ChooseNewTargets: true, FromLastKnown: true}.Apply(ctx)
		}
		return effects.DrawCards{N: 1}.Apply(ctx)
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		if effects.CmdrB13SpellTargetsPermanentOrPlayer(g, ev.CardID) {
			item.Params.Amount = 1
		}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "40ed32a6-ad56-48c7-aecf-b4238c34c212",
		Name:            "Shiko and Narset, Unified",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
