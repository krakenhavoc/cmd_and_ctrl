package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howling Galefang — Creature — Beast {2}{G}{G}, 3/3 (Wilds of Eldraine,
// #4471):
//
//	"Vigilance
//	 This creature has haste as long as you own a card in exile that has
//	 an Adventure."
//
// A layer 6 haste grant, on while any face-up adventurer card you OWN is
// in exile (cast as an Adventure or exiled by anything else; ownership,
// not control). It switches off the moment the last one leaves exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f12c2b65-4a04-4fab-ae77-973fe5cc823d",
		Name:            "Howling Galefang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{func() game.StaticAbility {
			ab := KeywordGrant(whileYouHave(ownsExiledAdventure, selfCreature), "haste")
			ab.Label = "Has haste as long as you own a card in exile that has an Adventure."
			ab.DependsOnEvents = []game.EventKind{game.EventZoneMove}
			return ab
		}()},
	})
}
