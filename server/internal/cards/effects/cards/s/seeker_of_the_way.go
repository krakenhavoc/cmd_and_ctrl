package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seeker of the Way — Creature — Human Warrior {1}{W}, 2/2 (S82 GAP, #5948):
//
//	"Prowess
//	 Whenever you cast a noncreature spell, this creature gains lifelink
//	 until end of turn."
func init() {
	const label = "Seeker of the Way — this creature gains lifelink until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "e206842f-139b-490d-bc10-d365705ac720",
		Name:            "Seeker of the Way",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, label,
				effects.ThisGainsKeywordUntilEndOfTurn("lifelink", label)),
		},
	})
}
