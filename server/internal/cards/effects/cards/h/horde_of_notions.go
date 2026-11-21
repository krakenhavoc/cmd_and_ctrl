package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horde of Notions — Legendary Creature — Elemental {W}{U}{B}{R}{G}, 5/5
// (Lorwyn Eclipsed Commander, #5368):
//
//	"Vigilance, trample, haste
//	 {W}{U}{B}{R}{G}: You may play target Elemental card from your graveyard
//	 without paying its mana cost."
//
// Arcane Heist's grant: a free-cast permission over the one targeted card,
// taken during the resolution (CR 608.2g, #4250) through the resolution_cast
// offer, so a creature is cast at instant speed, as the ability's own
// resolution allows. Declining leaves the card in the graveyard. Only a
// target that is still an Elemental card in your graveyard as the ability
// resolves is offered (CR 608.2b). A cast from the graveyard goes to the
// graveyard again if it would leave the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7656948-4d7d-47f7-b8a1-fc9716155e37",
		Name:            "Horde of Notions",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: You may play target Elemental card from your graveyard without paying its mana cost.",
			Cost:    effects.ManaCost("{W}{U}{B}{R}{G}"),
			Targets: effects.TargetCardInGraveyard("target Elemental card from your graveyard", effects.YouOwn(), effects.Subtype("Elemental")),
			Effect:  effects.Cmdr2S12HordePlay,
		}},
	})
}
