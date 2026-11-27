package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Response — Instant {1}{U} (Invasion, S82 #5616):
//
//	"Counter target spell or ability an opponent controls that targets a land
//	 you control. If a permanent's ability is countered this way, destroy that
//	 permanent.
//	 Draw two cards."
//
// The clause is a stack-item predicate (an opponent's, with a land you control
// among its targets); a spell is countered and nothing is destroyed, an
// ability's source permanent is destroyed after it is countered. Mana abilities
// never use the stack, so they are not legal targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2faa8b6-e433-4171-9774-9170484530c4",
		Name:         "Teferi's Response",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		Targets:      effects.TargetSpellOrAbility("target spell or ability an opponent controls that targets a land you control", effects.SpellOrAbilityTargetingYourLand()),
		OnResolve:    effects.TeferisResponseResolve,
	})
}
