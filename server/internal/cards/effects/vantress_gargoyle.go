package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vantress Gargoyle — Artifact Creature — Gargoyle {1}{U}, 5/4:
//
//	"Flying
//	 This creature can't attack unless defending player has seven or
//	 more cards in their graveyard.
//	 This creature can't block unless you have four or more cards in
//	 hand.
//	 {T}: Each player mills a card."
//
// The attack half is #1879's restriction (ADR 0107 §2, CR 508.1c), asked
// of each target's defending player (CR 508.5, 508.5a). The block half
// is a CR 509.1b block rule read as blockers are declared. The mill is
// every player still in the game, you included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc7c0197-bfa1-43b8-87c8-f134192b14c0",
		Name:            "Vantress Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{CantAttackUnlessDefendingPlayerHasCardsInGraveyard(7)},
		BlockRules: []game.BlockRule{CantBlockUnless(YouHaveCardsInHandAtLeast(4),
			"it can't block unless its controller has four or more cards in hand")},
		Activated: []ActivatedAbility{{
			Label: "{T}: Each player mills a card.",
			Cost:  game.AbilityCost{Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b02EachPlayerMills(g, item, 1)
			},
		}},
	})
}
