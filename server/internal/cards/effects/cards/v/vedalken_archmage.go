package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vedalken Archmage — Creature — Vedalken Wizard {2}{U}{U}, 0/2
// (EDHREC rank 3113):
//
//	"Whenever you cast an artifact spell, draw a card."
//
// The artifact deck's Beast Whisperer. Sai's condition
// (b29ArtifactSpellCastByYou — the spell is read off the stack,
// where its type line is intact) and a draw of one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "568cf486-0261-4634-ac36-a6507101b2d0",
		Name:         "Vedalken Archmage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B29ArtifactSpellCastByYou(ev, source, g)
			}, "Vedalken Archmage — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
