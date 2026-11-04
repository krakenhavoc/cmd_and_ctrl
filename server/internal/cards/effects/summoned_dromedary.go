package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summoned Dromedary — Creature — Spirit Camel {3}{W}, 4/3 (Secrets of
// Strixhaven, #4456):
//
//	"Vigilance
//	 {1}{W}: Return this card from your graveyard to your hand. Activate only
//	 as a sorcery."
//
// Gollum, Patient Plotter's graveyard ability with a mana cost: it works
// only from the graveyard and only at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "697b14c6-fe6a-4880-9cf2-b6602a52ca0b",
		Name:            "Summoned Dromedary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:        "{1}{W}: Return this card from your graveyard to your hand. Activate only as a sorcery.",
			Cost:         ManaCost("{1}{W}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       returnThisCardFromYourGraveyardToHand,
		}},
	})
}
