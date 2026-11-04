package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garrison Excavator — Creature — Orc Sorcerer {3}{R}, 3/4 (Secrets of
// Strixhaven, #4456):
//
//	"Menace
//	 Whenever one or more cards leave your graveyard, create a 2/2 red and
//	 white Spirit creature token."
//
// A batch trigger over cards of yours that leave a graveyard, by a zone move
// or by being cast from it (socB26CardsLeaveYourGraveyard): cards that leave
// as one event make one Spirit, cards that leave as separate events make one
// each. Cards leaving an opponent's graveyard do nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d808e64-a648-4f1b-a3a9-fd09568db281",
		Name:            "Garrison Excavator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			socB26CardsLeaveYourGraveyard("Garrison Excavator — create a 2/2 red and white Spirit creature token",
				Do(CreateToken{Template: TokenCard("2/2 red and white Spirit"), N: 1})),
		},
	})
}
