package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tomb Blade — Artifact Creature — Necron {4}{B}{B}, 5/4 (Warhammer 40,000
// Commander, #5948):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player loses
//	 life equal to the number of creatures they control unless they sacrifice a
//	 creature of their choice.
//	 Unearth {6}{B}{B}"
//
// A punisher in the shape of Bellowing Mauler's, asked of the player the damage
// event names: losing the life is always offered, the sacrifice only with a
// creature to give, and the count is of creatures that player controls as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c857de66-f4e0-4e97-ad01-fac623af8225",
		Name:            "Tomb Blade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Tomb Blade — that player loses life equal to the number of creatures they control unless they sacrifice a creature",
				effects.GapB189TombBladeLoseOrSacrifice),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{6}{B}{B}")},
	})
}
