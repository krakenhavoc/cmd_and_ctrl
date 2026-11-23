package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wizards of Thay — Creature — Human Wizard {3}{U}, 3/3 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and attacking
//	 that player or a planeswalker they control. Exile the tokens at end of
//	 combat.)
//	 Instant and sorcery spells you cast cost {1} less to cast.
//	 You may cast sorcery spells as though they had flash."
//
// Warchief Giant's myriad, Mizzix of the Izmagnus's instant-and-sorcery cost
// reduction (a flat {1} here) and Hypersonic Dragon's sorcery flash.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cda1e3b5-ea80-4e83-b305-88993d307170",
		Name:         "Wizards of Thay",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.GrantedMyriad("Wizards of Thay")},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{SorceryOnly: true}, "You may cast sorcery spells as though they had flash."),
		},
	})
}
