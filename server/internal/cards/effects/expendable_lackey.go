package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expendable Lackey — Creature — Human Citizen {U}, 1/1 (Streets of New
// Capenna, #5491):
//
//	"{1}{U}, Exile this card from your graveyard: Create a 1/1 blue Fish
//	 creature token with "This token can't be blocked." Activate only as
//	 a sorcery."
//
// Goldmeadow Nomad's graveyard activation: exiling the card is part of the
// cost (CR 602.2b), so it can be activated once, and only at sorcery timing
// (CR 602.5d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f6cbcd4-7a4b-4a38-bf29-995e82e658ce",
		Name:         "Expendable Lackey",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}{U}, Exile this card from your graveyard: Create a 1/1 blue Fish creature token with \"This token can't be blocked.\" Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{U}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: S82SncB07FishToken(), N: 1}),
		}},
	})
}
