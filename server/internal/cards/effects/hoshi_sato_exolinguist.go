package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoshi Sato, Exolinguist — Legendary Creature — Human Scientist {4}{U},
// 3/3 (Star Trek, #4973):
//
//	"Federation — When Hoshi Sato enters, look at the top X cards of
//	 your library, where X is the number of creature types among
//	 non-Borg creatures you control. Put one of them into your hand and
//	 the rest on the bottom of your library in a random order."
//
// "Federation" is a flavour word (CR 207.2d). X is counted as the
// trigger resolves (CR 608.2h), over the creatures its controller
// controls then, Hoshi included while she is still there: each creature
// type counts once however many creatures share it, a changeling counts
// every creature type (CR 702.73a), and a creature that is a Borg —
// a changeling among them — adds nothing. Types are read through the
// layered view. Geometer's Arthropod's look-and-take finishes it: the
// one card is mandatory, and an X of 0 looks at nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39426cd1-8c0d-4ec6-8747-f0693080f8f2",
		Name:         "Hoshi Sato, Exolinguist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hoshi Sato, Exolinguist — look at the top X cards of your library; put one into your hand and the rest on the bottom in a random order",
				hoshiSatoLook),
		},
	})
}

func hoshiSatoLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := trekB7CreatureTypesAmongNonBorgYouControl(g, item)
	if x <= 0 {
		return nil
	}
	player := item.Controller
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  g.LookAtTopOfLibraryForEffect(player, x),
		Max:    1,
		Label:  "Hoshi Sato, Exolinguist — put one of them into your hand",
		Key:    takeRestBottomRandom,
	}.Apply(ctx)
}
