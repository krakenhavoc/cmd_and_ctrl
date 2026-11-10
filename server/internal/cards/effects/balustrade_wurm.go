package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Balustrade Wurm — Creature — Wurm {3}{G}{G}, 5/5 (Duskmourn: House of
// Horror, #4466):
//
//	"This spell can't be countered.
//	 Trample, haste
//	 Delirium — {2}{G}{G}: Return this card from your graveyard to the
//	 battlefield with a finality counter on it. Activate only if there are
//	 four or more card types among cards in your graveyard and only as a
//	 sorcery. (If a creature with a finality counter on it would die, exile
//	 it instead.)"
//
// "Can't be countered" is the spec's CantBeCountered (Surrak, Elusive
// Hunter's), working on the stack (CR 113.6g); trample and haste are printed
// keywords. The delirium ability is Resurrected Cultist's graveyard return
// (returnThisFromGraveyardWithAFinalityCounter) with {2}{G}{G}: delirium is an
// ability word (CR 207.2c), the "activate only if" is checked as it is
// activated (CR 602.5b) and the Wurm itself counts as a creature card in that
// graveyard, sorcery timing is CR 602.5d, and the counter is on the creature
// as it enters so it is exiled instead of dying (CR 122.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2dd32fec-5bc3-4cc0-a9b4-814229a64914",
		Name:            "Balustrade Wurm",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample", "haste"},
		Activated: []ActivatedAbility{{
			Label:        "Delirium — {2}{G}{G}: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only if there are four or more card types among cards in your graveyard and only as a sorcery.",
			Cost:         ManaCost("{2}{G}{G}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return b16CardTypesInGraveyard(g, controller) >= 4
			},
			Effect: returnThisFromGraveyardWithAFinalityCounter,
		}},
	})
}
