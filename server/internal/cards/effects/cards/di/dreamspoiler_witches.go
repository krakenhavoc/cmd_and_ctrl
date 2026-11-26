package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreamspoiler Witches — Creature — Faerie Wizard {3}{B}, 2/2 (Lorwyn, S82
// OLD b545, #5593):
//
//	"Flying
//	 Whenever you cast a spell during an opponent's turn, you may have
//	 target creature get -1/-1 until end of turn."
//
// Blightcaster's optional targeted shrink on a cast trigger, with "during
// an opponent's turn" as the condition: the caster is the Witches'
// controller and is not the active player. The target is chosen as the
// trigger goes on the stack and the "may" is asked as it resolves.
//
// No simplification.
func init() {
	cast := effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Actor == source.Controller && !effects.IsActivePlayer(g, source.Controller)
	}, "Dreamspoiler Witches — target creature gets -1/-1 until end of turn",
		effects.BoostTargetUntilEOT("Dreamspoiler Witches — -1/-1 until end of turn", -1, -1))
	effects.Register(effects.Spec{
		OracleID:        "822eb600-cfaf-4ea1-a016-0c1f0cfa6436",
		Name:            "Dreamspoiler Witches",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Optional(effects.Targeting(cast, effects.TargetCreature("target creature")),
				"Dreamspoiler Witches — have target creature get -1/-1 until end of turn?"),
				effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
		},
	})
}
