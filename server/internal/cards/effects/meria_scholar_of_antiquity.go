package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meria, Scholar of Antiquity — Legendary Creature — Elf Artificer {1}{R}{G},
// 3/3 (Dominaria United, #5487):
//
//	"Tap an untapped nontoken artifact you control: Add {G}.
//	 Tap two untapped nontoken artifacts you control: Exile the top card of
//	 your library. You may play it this turn."
//
// Both costs tap other permanents (CR 701.26a), not the {T} symbol, so an
// artifact that arrived this turn pays them. The first is a mana ability
// (CR 605.1a): it has no target, adds mana, and is not a loyalty ability. The
// second uses the stack; the exiled card, a land included, may be played
// until end of turn (the impulse shape, b12ImpulseExileForTurn).
//
// No simplification.
func init() {
	nontokenArtifact := func(label string) game.TapOthersCost {
		return game.TapOthersCost{
			Count:  1,
			Filter: TargetPermanent(label, Artifact(), Not(IsTokenPredicate())),
			Label:  label,
		}
	}
	one := nontokenArtifact("an untapped nontoken artifact you control")
	two := nontokenArtifact("two untapped nontoken artifacts you control")
	two.Count = 2
	Register(Spec{
		OracleID:     "24aaadda-ef1c-44eb-9b01-4847bf454cd5",
		Name:         "Meria, Scholar of Antiquity",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{TapOthers: &one},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped nontoken artifacts you control: Exile the top card of your library. You may play it this turn.",
			Purpose: game.Purpose{Impulse: 1, Answers: game.AnswerValue},
			Cost:    game.AbilityCost{TapOthers: &two},
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, err := b12ImpulseExileForTurn(g, item, 1)
				return err
			},
		}},
	})
}
