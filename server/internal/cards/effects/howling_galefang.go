package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howling Galefang — Creature — Beast {2}{G}{G}, 4/4 (Wilds of Eldraine,
// #4471):
//
//	"Vigilance
//	 This creature has haste as long as you own a card in exile that has
//	 an Adventure."
//
// The haste is Grand Ball Guest's conditional layer 6 grant (KeywordGrant
// over whileYouHave, CR 613.1f) with youOwnAnAdventureCardInExile as the
// condition: any face-up card you own in exile that has an Adventure
// (CR 715.1), whether it went on an adventure or was exiled some other
// way, and whoever exiled it.
//
// No simplification.
func init() {
	haste := KeywordGrant(whileYouHave(youOwnAnAdventureCardInExile, selfCreature), "haste")
	haste.Label = "Has haste as long as you own a card in exile that has an Adventure."
	Register(Spec{
		OracleID:        "f12c2b65-4a04-4fab-ae77-973fe5cc823d",
		Name:            "Howling Galefang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{haste},
	})
}
