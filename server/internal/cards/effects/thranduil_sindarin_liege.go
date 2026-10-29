package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thranduil, Sindarin Liege // Silvan Rally — an Adventure card (CR 715,
// The Hobbit, #4454), oracle f6771d32…:
//
//	Thranduil, Sindarin Liege — Legendary Creature — Elf Noble
//	  {2}{G/U}{G/U}, 2/3
//	  "Other Elves you control get +1/+1.
//	   Landfall — Whenever a land you control enters, create a 1/1 green
//	   Elf creature token."
//	Silvan Rally — Sorcery — Adventure {1}{G/U}{G/U}
//	  "Mill four cards, then put up to two land cards from among them
//	   into your hand. (Then exile this card. You may cast the creature
//	   later from exile.)"
//
// The anthem is the tribal lord (TribalAnthem, "Elf" read through the
// source's text, CR 612.2) and the landfall token is Thranduil, the
// Strategist's. Silvan Rally is Speak Secrets' mill-and-take with a floor
// of zero and a ceiling of two.
//
// No simplification.

// thranduilSindarinLiegeOracleID is shared by both faces of the card.
const thranduilSindarinLiegeOracleID = "f6771d32-395e-4832-b33d-cbc12dff1516"

func init() {
	Register(Spec{
		OracleID:     thranduilSindarinLiegeOracleID,
		Name:         "Thranduil, Sindarin Liege",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Thranduil, Sindarin Liege — create a 1/1 green Elf", game.LandfallPayoff{Tokens: 1},
				b34CreateTokens(thranduilElfToken, 1)),
		},
	})
	Register(Spec{
		OracleID:     thranduilSindarinLiegeOracleID + "#1",
		Name:         "Silvan Rally",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return hobB9MillFour(ctx, silvanRallyMilled)
		},
	})
}

// silvanRallyMilled is the clause after the mill: up to two land cards from
// among the milled cards.
var silvanRallyMilled = CardTail("silvan-rally/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return hobB9TakeFromMilled(ctx, o.IDs, Land(),
		"Silvan Rally — put up to two land cards from among them into your hand", 0, 2)
})
