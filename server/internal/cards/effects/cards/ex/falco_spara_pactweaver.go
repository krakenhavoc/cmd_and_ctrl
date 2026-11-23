package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// falcoSparaAltCostKey is the offer a cast from the top of the library
// claims: the spell's mana cost and one counter removed from a creature the
// caster controls. An on-disk identity: never renamed, never reused.
const FalcoSparaAltCostKey = "falco_spara_remove_counter"

// Falco Spara, Pactweaver — Legendary Creature — Bird Demon {1}{G}{W}{U},
// 3/3:
//
//	"Flying, trample
//	 Falco Spara enters with a shield counter on it.
//	 You may look at the top card of your library any time.
//	 You may cast spells from the top of your library by removing a counter
//	 from a creature you control in addition to paying their other costs."
//
// Into the Pit's top-of-library window (LibraryTopVisible and
// PlayFromTopOfYourLibrary, cast-only because a land is not a spell) with
// Dawnhand Dissident's price: the spell's mana cost plus one counter of any
// kind removed from among creatures you control (RemoveCountersInAddition,
// N = 1). A caster whose creatures hold no counter cannot claim it
// (CR 118.3). Falco Spara may remove its own shield counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "b9c1ca24-5696-4777-8465-13a8f78e4fb5",
		Name:              "Falco Spara, Pactweaver",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flying", "trample"},
		LibraryTopVisible: game.LibraryTopOwner,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterShield, 1, "Falco Spara enters with a shield counter on it."),
		},
		CastPermissions: []game.CastPermission{func() game.CastPermission {
			p := effects.PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Cast a spell from the top of your library by removing a counter from a creature you control in addition to paying its other costs (Falco Spara)")
			p.CastOnly = true
			p.AltCostKey = FalcoSparaAltCostKey
			p.RemoveCountersInAddition = 1
			return p
		}()},
	})
}
