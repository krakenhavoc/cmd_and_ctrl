package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dreams of Steel and Oil — Sorcery {B} (The Brothers' War, #5482):
//
//	"Target opponent reveals their hand. You choose an artifact or creature
//	 card from it, then choose an artifact or creature card from their
//	 graveyard. Exile the chosen cards."
//
// Two picks, as printed: the hand pick (the hand is revealed to the table,
// CR 701.20a) and then a graveyard pick that offers only that player's
// artifact and creature cards. Each is made if anything qualifies, and
// neither depends on the other (an empty hand still leaves the graveyard
// pick). The hand card is exiled as it is picked and the graveyard card
// right after; nobody can act in between, because the prompt blocks the
// table, so it is the same as exiling both together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c308ebf-df5e-4f47-9b97-5c31944fe91c",
		Name:         "Dreams of Steel and Oil",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player: TargetedPlayer(ctx),
				Filter: Or(Artifact(), Creature()),
				Label:  "artifact or creature card",
				Exile:  true,
				Then:   dreamsOfSteelAndOilGraveyardPick,
			}.Apply(ctx)
		},
	})
}

// dreamsOfSteelAndOilGraveyardPick is "then choose an artifact or creature
// card from their graveyard", with the hand already revealed and nothing
// in it left to offer.
var dreamsOfSteelAndOilGraveyardPick = RevealedPickThen("revealed-pick/dreams-of-steel-and-oil-graveyard",
	func(ctx *Context, pick game.RevealedPick) error {
		return ChooseFromRevealedHand{
			Player:          pick.FromPlayer,
			Filter:          func(*game.Game, uuid.UUID, game.Card) bool { return false },
			Label:           "artifact or creature card from their graveyard",
			Exile:           true,
			Revealed:        true,
			FromGraveyard:   true,
			GraveyardFilter: Or(Artifact(), Creature()),
		}.Apply(ctx)
	})
