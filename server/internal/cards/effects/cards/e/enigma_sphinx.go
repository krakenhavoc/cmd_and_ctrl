package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enigma Sphinx — Artifact Creature — Sphinx {4}{W}{U}{B}, 5/4 (Commander
// 2018, S82 OLD batch 193, #5526):
//
//	"Flying
//	 When this creature is put into your graveyard from the battlefield, put
//	 it into your library third from the top.
//	 Cascade"
//
// Flying and cascade are the keywords. The tuck is Unquenchable Fury's
// "put into your graveyard" reading: any trip from the battlefield to the
// graveyard of the Sphinx's controller (its owner, when it is also its
// controller; a stolen one goes to its owner's graveyard, which is not
// "your" graveyard). It is mandatory, and it only moves the card while it
// is still in that graveyard as the trigger resolves, third from the top (or
// on the bottom of a library of fewer than two cards).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6fb84ecd-ab60-4c2c-987d-8a0749db5777",
		Name:            "Enigma Sphinx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "cascade"},
		Triggered: []game.TriggeredAbility{
			effects.Cascade(),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source) && source.Owner == source.Controller
			}, "Enigma Sphinx — put it into your library third from the top", enigmaSphinxTuck),
		},
	})
}

// enigmaSphinxTuck puts the Sphinx third from the top if it is still in the
// graveyard it was put into.
func enigmaSphinxTuck(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.B22TuckThirdFromTop(g, item.SourceCardID)
}
