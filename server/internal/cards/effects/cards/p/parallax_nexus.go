package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Parallax Nexus — Enchantment {2}{B} (Nemesis, S82 OLD, #5618):
//
//	"Fading 5
//	 Remove a fade counter from this enchantment: Target opponent exiles a card
//	 from their hand. Activate only as a sorcery.
//	 When this enchantment leaves the battlefield, each player returns to their
//	 hand all cards they own exiled with it."
//
// Fading is the shared pair. The targeted opponent picks the card (CR 406.3) and
// it is exiled linked to this object; the leave trigger returns every card still
// exiled with it to its owner's hand (Hallowed Spiritkeeper's twin,
// B36ReturnCardsExiledWithToOwnersHands). Sorcery speed is the activation's own
// timing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "40fbcf4b-d79e-4e6b-bcec-1d1544b3c4a2",
		Name:         "Parallax Nexus",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Fading(5)},
		Triggered: []game.TriggeredAbility{
			effects.FadingUpkeep(),
			effects.On(game.EventLTB, effects.Self, "Parallax Nexus — each player returns to their hand the cards they own exiled with it", effects.B36ReturnCardsExiledWithToOwnersHands),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "Remove a fade counter from this enchantment: Target opponent exiles a card from their hand. Activate only as a sorcery.",
			Cost:         effects.RemoveCountersFromThis(game.CounterFade, 1),
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			SorcerySpeed: true,
			Effect:       effects.OldB678TargetOpponentExilesFromHandLinked,
		}},
	})
}
