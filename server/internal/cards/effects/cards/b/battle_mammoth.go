package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battle Mammoth — {3}{G}{G} Creature — Elephant 6/5
//
//	"Trample
//	 Whenever a permanent you control becomes the target of a spell or
//	 ability an opponent controls, you may draw a card.
//	 Foretell {2}{G}{G}"
//
// EventBecomesTarget fires once per target slot (CR 115.3), so a spell that
// targets two of your permanents draws twice. A player target is no
// permanent, and your own spells and abilities don't count. The Mammoth
// itself is a permanent you control.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a6db509-7ef4-483c-b8a0-62a4f3e16ef9",
		Name:            "Battle Mammoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Draws: 1},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{2}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventBecomesTarget,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && effects.OnBattlefield(g, ev.CardID) && c.Controller == source.Controller
				},
				"Battle Mammoth — you may draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Battle Mammoth — draw a card?", 1),
		},
	})
}
