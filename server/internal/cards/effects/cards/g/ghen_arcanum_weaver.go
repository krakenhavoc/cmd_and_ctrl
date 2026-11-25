package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghen, Arcanum Weaver — Legendary Creature — Human Wizard {R}{W}{B}, 2/3
// (Commander Legends):
//
//	"{R}{W}{B}, {T}, Sacrifice an enchantment: Return target enchantment card
//	 from your graveyard to the battlefield."
//
// Altar of Bhaal's return, paid with {T} (Ghen's own, so summoning sickness
// applies, CR 302.6) and the sacrifice of an enchantment as costs
// (CR 118.3). The target is chosen as the ability is activated, so the
// enchantment sacrificed as a cost is not a legal target; it comes back
// under its owner's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ac59552-9562-492c-aea4-4d64bad58e3e",
		Name:         "Ghen, Arcanum Weaver",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{W}{B}, {T}, Sacrifice an enchantment: Return target enchantment card from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{R}{W}{B}"), effects.TapCost(), effects.SacrificeN(1, "an enchantment", effects.Enchantment())),
			Targets: effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.Enchantment(), effects.YouOwn()),
			Effect:  effects.ReturnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}
