package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gabriel Lorca, Mirror Menace — Legendary Creature — Human Rogue
// Officer {4}{B}, 5/4 (Star Trek, #4973):
//
//	"Menace
//	 Whenever you face a dilemma, you may pay {B}. If you do, return this
//	 card from your graveyard to your hand. (You face a dilemma as you
//	 choose one or more modes for a spell or ability.)"
//
// The trigger watches from the graveyard only (InGraveyard, Bloodghast's
// and Silvan Reveler's shape, CR 113.6): Lorca on the battlefield does
// nothing when you face a dilemma. "You" is the card's owner while it is
// there (CR 108.4), so it is its owner's dilemmas that count
// (EventFacedDilemma, batch 7). The {B} is asked as the ability resolves
// (Forsaken Miner's offer, with Silvan Reveler's hand twin): the return
// happens only if the mana was really paid, and only for the very object
// that triggered (CR 400.7), so a Lorca that left the graveyard in
// response stays where it went.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6349eecb-dabd-4dcb-a8fb-c8ee66f61461",
		Name:            "Gabriel Lorca, Mirror Menace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventFacedDilemma, YouFacedADilemma,
				"Gabriel Lorca, Mirror Menace — you may pay {B} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return mayPayForGraveyardReturnToHand(g, item, "{B}", "Gabriel Lorca — pay {B} to return it from your graveyard to your hand?")
				})),
		},
	})
}
