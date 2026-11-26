package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Myr Reservoir — Artifact {3} (Scars of Mirrodin, S82 OLD batch 472, #5577):
//
//	"{T}: Add {C}{C}. Spend this mana only to cast Myr spells or activate
//	 abilities of Myr.
//	 {3}, {T}: Return target Myr card from your graveyard to your hand."
//
// A mana ability producing two colorless with Flamebraider's restriction over
// Myr: the mana pays for a Myr spell or for an activated ability of a Myr
// source, and for nothing else. The second ability returns a Myr card (a card
// with the creature type Myr, not one that merely says it) from your own
// graveyard; a card that has left by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82b327c6-5cd9-4bee-bef4-860522bded9c",
		Name:         "Myr Reservoir",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "{T}: Add {C}{C}. Spend this mana only to cast Myr spells or activate abilities of Myr.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Myr")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Myr")},
			)},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Return target Myr card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target Myr card from your graveyard", effects.YouOwn(), effects.HasSubtype("Myr")),
			Effect:  effects.ReturnTargetCardToHand,
		}},
	})
}
