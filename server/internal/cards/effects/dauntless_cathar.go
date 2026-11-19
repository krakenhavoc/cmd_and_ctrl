package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dauntless Cathar — Creature — Human Soldier {2}{W}, 3/2 (Innistrad
// Remastered):
//
//	"{1}{W}, Exile this card from your graveyard: Create a 1/1 white Spirit
//	 creature token with flying. Activate only as a sorcery."
//
// Works from the graveyard only; the exile is part of the cost.
func init() {
	Register(Spec{
		OracleID:     "c14bd133-ff17-48b3-a1c8-56c93d8c6e43",
		Name:         "Dauntless Cathar",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}{W}, Exile this card from your graveyard: Create a 1/1 white Spirit creature token with flying. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{W}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Effect:       Do(CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1}),
		}},
	})
}
