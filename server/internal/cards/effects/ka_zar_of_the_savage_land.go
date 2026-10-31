package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ka-Zar of the Savage Land — Legendary Creature — Human Barbarian Hero
// {4}{G}, 3/2 (Marvel Super Heroes, #4455):
//
//	"You may look at the top card of your library any time.
//	 You may play lands from the top of your library.
//	 When Ka-Zar enters, create Zabu, a legendary 2/2 green Cat creature
//	 token with "Landfall — Whenever a land you control enters, put a
//	 +1/+1 counter on Zabu.""
//
// Courser of Kruphix's land permission with the private look of Daredevil,
// Man Without Fear (LibraryTopOwner). Zabu is a catalog token (ADR 0083):
// his landfall trigger lives on the token, so it is the token's own ability
// and keeps working after Ka-Zar leaves, and it reads "Zabu" as the token
// itself. Legendary, so a second Zabu from a second Ka-Zar costs its
// controller one of them (CR 704.5j).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "88eeb519-5782-4732-bd34-81d1ee6dbc8f",
		Name:              "Ka-Zar of the Savage Land",
		Completeness:      CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			PlayFromTopOfYourLibrary(
				game.PermissionFilter{LandsOnly: true},
				"Play a land from the top of your library (Ka-Zar of the Savage Land)"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ka-Zar — create Zabu, a legendary 2/2 green Cat creature token", Do(CreateToken{Template: zabuToken(), N: 1})),
		},
	})
}

// zabuToken is Ka-Zar's Zabu.
func zabuToken() game.Card { return tokenFromCatalog(printedZabuToken) }

// printedZabuToken is Zabu as PRINTED: a legendary 2/2 green Cat whose
// landfall puts a +1/+1 counter on itself.
func printedZabuToken() tokenTemplate {
	return tokenTemplate{
		Slug: "zabu",
		Card: game.Card{
			Name:      "Zabu",
			TypeLine:  "Token Legendary Creature — Cat",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"G"},
		},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Zabu — landfall: put a +1/+1 counter on it", game.LandfallPayoff{Counters: 1}, plusOneCountersOnThis(1)),
		},
		Text: "Landfall — Whenever a land you control enters, put a +1/+1 counter on Zabu.",
	}
}
