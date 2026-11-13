package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cabal Archon — Creature — Human Cleric {2}{B}, 2/2 (Onslaught):
//
//	"{B}, Sacrifice a Cleric: Target player loses 2 life and you gain 2
//	 life."
//
// Goblin Chirurgeon's sacrifice cost over Clerics, so the Archon may
// sacrifice itself. Any player is a legal target, you included. The
// gain happens only when the target is still legal as the ability
// resolves (CR 608.2b), the same body Vein Ripper's drain uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d80f016-5a4a-4843-9aec-0fe64f7f7ae8",
		Name:         "Cabal Archon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice a Cleric: Target player loses 2 life and you gain 2 life.",
			Cost:    Plus(ManaCost("{B}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Cleric", Subtype("Cleric"))}),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return targetOpponentLosesAndYouGain(g, item, 2)
			},
		}},
	})
}
