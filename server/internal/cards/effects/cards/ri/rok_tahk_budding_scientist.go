package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rok-Tahk, Budding Scientist — {4}{G}{G} Legendary Creature — Brikar
// Scientist, 5/5 (Star Trek, #4973):
//
//	"Creatures you control have "{T}: Add {G}{G}."
//	 When Rok-Tahk enters, create Murf, a legendary 0/1 green Ooze Slug
//	 creature token with indestructible."
//
// Cryptolith Rite's grant (ADR 0093) with a fixed {G}{G}: every creature
// its controller controls, Rok-Tahk included, has the mana ability while
// Rok-Tahk is on the battlefield, and a creature still under summoning
// sickness can't use it (CR 302.6). Murf is the printed token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef919269-043f-45ab-a5da-403cdaa4da23",
		Name:         "Rok-Tahk, Budding Scientist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Grants:       []effects.AbilityGrant{effects.TapForManaGrant(rokTahkGrant, "{G}{G}", "Add {G}{G}", "{T}: Add {G}{G}.")},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(effects.B16CreaturesYouControl, rokTahkGrant), "Creatures you control have \"{T}: Add {G}{G}.\""),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Rok-Tahk, Budding Scientist — create Murf, a legendary 0/1 green Ooze Slug token with indestructible",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 green Legendary Ooze Slug named Murf with indestructible"), N: 1})),
		},
	})
}

const rokTahkGrant = "rok-tahk-budding-scientist/gg"
