package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discordant Dirge — Enchantment {3}{B}{B} (Urza's Saga, S82 OLD b724):
//
//	At the beginning of your upkeep, you may put a verse counter on this enchantment.
//	{B}, Sacrifice this enchantment: Look at target opponent's hand and choose up to X cards from it, where X is the number of verse counters on this enchantment. That player discards those cards.
//
// The upkeep counter is a "you may" trigger (Barrin's Codex's shape). The cost is paid with the enchantment sacrificed, so X is the verse counters it had as it last existed on the battlefield (CR 608.2h). Only you see the hand; the cards you choose are discarded by the opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edd4741b-3ee6-421a-8bac-d60736755356",
		Name:         "Discordant Dirge",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB724VerseUpkeep("Discordant Dirge")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice this enchantment: Look at target opponent's hand and choose up to X cards from it, where X is the number of verse counters on this enchantment. That player discards those cards.",
			Cost:    effects.S82OldB724SacrificeCost("{B}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:  effects.S82OldB724DiscardUpToXFromTheirHand,
		}},
	})
}
