package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sovereigns of Lost Alara — Creature — Spirit {4}{W}{U}, 4/5 (Alara Reborn,
// #5586):
//
//	"Exalted
//	 Whenever a creature you control attacks alone, you may search your
//	 library for an Aura card that could enchant that creature, put it onto
//	 the battlefield attached to that creature, then shuffle."
//
// Exalted is a printed keyword the engine derives. The search ability is
// Squall's "attacks alone" (game.AttackedAlone, read once the declaration is
// stamped) with Auratouched Mage's Aura search: the Aura must be able to
// enchant the lone attacker (CR 303.4f) and enters attached to it. The search
// is optional and may find nothing; the library is shuffled either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86e13e19-45df-4d73-8600-51b7caa3e6ab",
		Name:            "Sovereigns of Lost Alara",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "Sovereigns of Lost Alara — search your library for an Aura card that could enchant that creature", effects.S82OldB501SovereignsSearch),
		},
	})
}
