package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bolg's Company — Creature — Goblin Soldier {B}{R}, 2/2 (The Hobbit):
//
//	"This creature has haste as long as you control another Goblin.
//	 {T}, Sacrifice another Goblin: Add {B}{R}."
//
// The haste is a self-only layer-6 keyword grant gated on the board, read
// live: it comes with the other Goblin and goes when that one does. It
// counts only OTHER Goblins (the Company is one itself, so selfWhileYouControlA
// would always say yes). The mana ability is a tap plus a sacrifice of
// another Goblin (Another sets ExcludeSource, so the Company cannot feed
// itself, though a second Bolg's Company can); it adds one black and one red
// mana, and it is a mana ability, so it does not use the stack. As a tap
// ability of a creature it obeys summoning sickness (CR 302.6), which the
// haste it grants itself with another Goblin lifts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa602f8f-1d80-4f6d-8b8f-d1a1f36037bd",
		Name:         "Bolg's Company",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.HobA9WhileYouControlAnother("Goblin"), "haste"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				SacrificeOther: effects.Another(effects.SacrificeSpec("another Goblin", effects.Subtype("Goblin"))),
			},
			Produced: "{B}{R}",
			Label:    "{T}, Sacrifice another Goblin: Add {B}{R}",
			Answers:  game.AnswerSacOutlet,
		}},
	})
}
