package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commander's Sphere — Artifact {3}:
//
//	"{T}: Add one mana of any color in your commander's color identity.
//	 Sacrifice this artifact: Draw a card."
//
// A three-mana rock that is never a dead draw late, which is the
// whole reason it out-plays the strictly-cheaper Signets in casual
// Commander.
//
// Both halves already had machinery: the identity-narrowed pipe is
// Arcane Signet's (NarrowToCommanderIdentity filters "{W|U|B|R|G}"
// against the controller's commander identity when the ability fires), and the
// cash-in is an ordinary CR 602 activated ability whose entire cost
// is sacrificing the source. No mana, no tap — the Sphere can be
// cracked with its mana ability already used, tapped, this turn, at
// instant speed, which is exactly the printed card.
func init() {
	Register(Spec{
		OracleID:     "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b",
		Name:         "Commander's Sphere",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
			// The printed text asks for the narrowing (manaPickOptionsFor).
			NarrowToCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice this artifact: Draw a card.",
			Cost:  SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
