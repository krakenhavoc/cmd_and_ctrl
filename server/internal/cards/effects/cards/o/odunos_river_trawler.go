package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Odunos River Trawler — Creature — Zombie {2}{B}, 2/2 (Born of the Gods, S82
// OLD b379, #5559):
//
//	"When this creature enters, return target enchantment creature card from
//	 your graveyard to your hand.
//	 {W}, Sacrifice this creature: Return target enchantment creature card from
//	 your graveyard to your hand."
//
// Midnight Recovery's graveyard target (printed types: enchantment and
// creature) for both halves; the enter trigger and the sacrifice ability each
// return the one target through ReturnTargetedCardToHand. The sacrifice is a
// cost, so the Trawler is gone before the ability resolves.
//
// No simplification.
func init() {
	const label = "target enchantment creature card from your graveyard"
	effects.Register(effects.Spec{
		OracleID:     "0282b5c0-e792-4099-a3fa-329eda11fb44",
		Name:         "Odunos River Trawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Odunos River Trawler — return target enchantment creature card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard(label, effects.YouOwn(), effects.Creature(), effects.Enchantment())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, Sacrifice this creature: Return target enchantment creature card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard(label, effects.YouOwn(), effects.Creature(), effects.Enchantment()),
			Effect:  effects.ReturnTargetedCardToHand,
		}},
	})
}
