package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectral Snatcher — Creature — Spirit {4}{B}{B}, 6/5 (Duskmourn: House
// of Horror, #4466):
//
//	"Ward—Discard a card. (Whenever this creature becomes the target of
//	 a spell or ability an opponent controls, counter it unless that
//	 player discards a card.)
//	 Swampcycling {2} ({2}, Discard this card: Search your library for a
//	 Swamp card, reveal it, put it into your hand, then shuffle.)"
//
// Ward—Discard a card is Ward(WardDiscard(1)) (ward.go, CR 702.21a), as
// on Forum Necroscribe: the payer chooses the card. Swampcycling is the
// shared Typecycling constructor (CR 702.29e): any card with the Swamp
// subtype, basic or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "371762f6-e07b-4c0e-a4ee-6c35a3809eb6",
		Name:         "Spectral Snatcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscard(1), "Spectral Snatcher — ward, discard a card"),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
