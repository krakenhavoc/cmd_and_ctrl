package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strands of Night — Enchantment {2}{B}{B} (Weatherlight):
//
//	"{B}{B}, Pay 2 life, Sacrifice a Swamp: Return target creature card from your
//	 graveyard to the battlefield."
//
// All three costs are paid as the ability is activated (CR 602.2b); the Swamp is
// any Swamp the activator controls, and the creature card returns under their
// control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e8ae1f7-8926-414a-8434-db0b0f877505",
		Name:         "Strands of Night",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, Pay 2 life, Sacrifice a Swamp: Return target creature card from your graveyard to the battlefield.",
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Reanimates: true}),
			Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.PayLife(2), effects.SacrificeN(1, "a Swamp", effects.Subtype("Swamp"))),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
			Effect:  effects.OldB775ReturnTargetCreatureCardToBattlefield,
		}},
	})
}
