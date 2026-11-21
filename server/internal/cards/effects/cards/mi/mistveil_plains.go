package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistveil Plains — Land — Plains (EDHREC rank 3229):
//
//	"({T}: Add {W}.)
//	 This land enters tapped.
//	 {W}, {T}: Put target card from your graveyard on the bottom of
//	 your library. Activate only if you control two or more white
//	 permanents."
//
// Shadowmoor's white hybrid land: a repeatable tuck of your own
// graveyard card, for a commander deck the way to get a key spell back
// into the library. The {W} is the Plains type's intrinsic ability.
// "Activate only if you control two or more white permanents" is the
// activation condition (CR 602.1b, #743), ControlsAtLeast(2, white).
// The target is re-checked on resolution, so a card exiled from the
// graveyard in response is left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb5c1817-ac22-4779-9005-251bc354f181",
		Name:         "Mistveil Plains",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{W}, {T}: Put target card from your graveyard on the bottom of your library. Activate only if you control two or more white permanents.",
			Cost:      effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets:   effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
			Condition: effects.ControlsAtLeast(2, effects.MatchColor("W")),
			Effect:    effects.S82One13TuckTargetToLibraryBottom,
		}},
	})
}
