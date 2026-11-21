package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transmogrant Altar — Artifact {3} (The Brothers' War, S82):
//
//	"{B}, {T}, Sacrifice a creature: Add {C}{C}{C}.
//	 {2}, {T}, Sacrifice a creature: Create a 3/3 colorless Zombie artifact
//	 creature token. Activate only as a sorcery."
//
// The first is a mana ability (CR 605.1a: no target, adds mana), so it
// resolves at once with no stack (Ashnod's Altar's shape) and pays {B}, the
// tap and the sacrifice as costs. The second uses the stack and is
// sorcery-speed. Either may sacrifice any creature, including an animated
// one that is also the Altar's only fuel.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "239fea8d-c49d-40bc-96e1-b84d328a0e1e",
		Name:         "Transmogrant Altar",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Mana:           "{B}",
				Tap:            true,
				SacrificeOther: effects.SacrificeACreature().SacrificeOther,
			},
			Produced: "{C}{C}{C}",
			Label:    "{B}, {T}, Sacrifice a creature: Add {C}{C}{C}.",
			Answers:  game.AnswerSacOutlet,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}, Sacrifice a creature: Create a 3/3 colorless Zombie artifact creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Zombie artifact"), N: 1}),
		}},
	})
}
