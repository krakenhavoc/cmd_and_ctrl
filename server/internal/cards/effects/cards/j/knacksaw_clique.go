package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knacksaw Clique — Creature — Faerie Rogue {3}{U}, 1/4 (Shadowmoor, S82 OLD
// b531, #5591):
//
//	"Flying
//	 {1}{U}, {Q}: Target opponent exiles the top card of their library. Until end
//	 of turn, you may play that card. ({Q} is the untap symbol.)"
//
// {Q} is AbilityCost.UntapSelf (CR 107.6). The opponent's top card is exiled and
// you, not they, get the play permission until end of turn (Bloodsoaked Insight's
// impulse exile from another library): a land is playable as well as a spell, and
// the mana to cast it is paid as normal. A target that has left the game does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06af8e20-6f60-400d-9120-c99311232dde",
		Name:            "Knacksaw Clique",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {Q}: Target opponent exiles the top card of their library. Until end of turn, you may play that card.",
			Cost:    effects.OldB531UntapCost("{1}{U}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.TargetOpponentExilesTopYouMayPlay,
		}},
	})
}
