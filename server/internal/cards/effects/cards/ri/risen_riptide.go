package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Risen Riptide — Creature — Elemental {2}{U}, 0/5 (Zendikar Rising,
// #5508):
//
//	"Whenever you cast a kicked spell, this creature has base power and
//	 toughness 5/5 until end of turn."
//
// The trigger reads the cast spell's paid kicker off its stack item
// (S82OldB097YouCastAKickedSpell). The effect is a layer 7b "base power and
// toughness become 5/5" until end of turn (Brokers Initiate's), so counters
// and pumps still apply on top; it does nothing for a Riptide that has left.
//
// No simplification.
func init() {
	const label = "Risen Riptide — base power and toughness 5/5 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "e682ad51-001c-4465-951e-43df96b7d3b9",
		Name:         "Risen Riptide",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB097YouCastAKickedSpell, label,
				effects.S82OldB097ThisHasBasePTUntilEOT(label, 5, 5)),
		},
	})
}
