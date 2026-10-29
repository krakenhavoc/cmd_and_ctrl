package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Call Damage Control — {1}{G} Sorcery (Marvel Super Heroes, #4455):
//
//	"Choose up to two. Return those cards from your graveyard to your
//	 hand.
//	 • Target artifact card.
//	 • Target creature card.
//	 • Target enchantment card.
//	 • Target land card."
//
// A "choose up to two" spell (CR 700.2): zero, one or two different
// bullets, each with its own target in your graveyard. Each bullet
// returns its own target; one that left the graveyard by resolution is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	back := func(label, desc string, pred CardPredicate) game.ModeOption {
		return ModeDoing(label, TargetCardInGraveyard(desc, YouOwn(), pred), ReturnTheModesGraveyardTargetToHand)
	}
	Register(Spec{
		OracleID:     "49a32f34-01a9-4f25-bf38-b4d46ca67b2d",
		Name:         "Call Damage Control",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose up to two", 0, 2,
			back("Return target artifact card from your graveyard to your hand.", "target artifact card from your graveyard", Artifact()),
			back("Return target creature card from your graveyard to your hand.", "target creature card from your graveyard", Creature()),
			back("Return target enchantment card from your graveyard to your hand.", "target enchantment card from your graveyard", Enchantment()),
			back("Return target land card from your graveyard to your hand.", "target land card from your graveyard", Land()),
		),
	})
}
