package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rot Wolf — Creature — Phyrexian Wolf {2}{G}, 2/2 (Mirrodin Besieged, S82 OLD
// b466, #5576):
//
//	"Infect
//	 Whenever a creature dealt damage by Rot Wolf this turn dies, you may draw
//	 a card."
//
// Infect is the engine's damage tail (ADR 0056): a creature that took
// infect damage was "dealt damage" all the same, so it counts. The watcher is
// Abattoir Ghoul's over every creature this Wolf damaged earlier in the turn,
// with the CR 603.5 "you may" prompt before the draw.
//
// No simplification.
func init() {
	t := effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		_, died := effects.DiedCreature(ev, g)
		return died && ev.CardID != source.InstanceID && effects.DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
	}, "Rot Wolf — you may draw a card", effects.Do(effects.DrawCards{N: 1}))
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: "Rot Wolf: draw a card?"}
	effects.Register(effects.Spec{
		OracleID:        "7409906d-7111-48fa-bfea-8b95e0fdb7dc",
		Name:            "Rot Wolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
