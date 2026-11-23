package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sidisi, Regent of the Mire — Legendary Creature — Zombie Snake Warlock
// {1}{B}, 1/3 (Tarkir: Dragonstorm, #4463):
//
//	"{T}, Sacrifice a creature you control with mana value X other than
//	 Sidisi: Return target creature card with mana value X plus 1 from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// X is announced with the activation (CR 602.2b), exactly as Lazav's and
// Chthonian Nightmare's "mana value X" are. It is the one number both
// clauses read: the sacrifice cost must name a creature with mana value X
// (SacrificeAnotherWithManaValueX, game.SacrificeManaValueX), and the
// target is a creature card in your graveyard with mana value X + 1
// (TargetSpec.ManaValueEqualsXPlus). The target is judged at announce and
// again as the ability resolves (CR 608.2b); a creature with {X} in its
// cost counts it as 0 (the ruling). A token has mana value 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72074972-418d-4dce-83cf-ce8cf310fc42",
		Name:         "Sidisi, Regent of the Mire",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Sacrifice a creature you control with mana value X other than Sidisi: Return target creature card with mana value X plus 1 from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.SacrificeAnotherWithManaValueX("a creature you control with mana value X other than Sidisi", effects.Creature(), effects.YouControl())),
			SorcerySpeed: true,
			Targets: effects.TargetCardInGraveyard("target creature card with mana value X plus 1 from your graveyard",
				effects.YouOwn(), effects.Creature()).WithManaValueEqualsXPlus(1),
			Effect: effects.ReturnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}
