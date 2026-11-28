package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agonizing Memories — Sorcery {2}{B}{B} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Look at target player's hand and choose two cards from it. Put them on top of that player's
//	 library in any order."
//
// Only the caster looks (ChooseFromRevealedHand Look); they choose two cards (or the whole hand, if it is
// smaller), which stay in the hand until the pile is ordered, and the caster then puts them on top of the owner's
// library in an order of their choice (ADR 0088). A hand with one card gives that card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ac9d397-10ca-43ce-bff9-45d631553486",
		Name:         "Agonizing Memories",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := effects.TargetedPlayer(ctx)
			if player == uuid.Nil {
				return nil
			}
			return effects.ChooseFromRevealedHand{
				Player: player,
				Label:  "card",
				Count:  2,
				Look:   true,
				Keep:   true,
				Then:   agonizingMemoriesPut,
			}.Apply(ctx)
		},
	})
}

// agonizingMemoriesPut puts the chosen cards on top of their owner's library in the order the caster
// picks. An on-disk key.
var agonizingMemoriesPut = effects.RevealedPickThen("revealed-pick/s82-old-b779-agonizing-memories-put",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		ids := make([]uuid.UUID, 0, len(pick.Chosen))
		for _, c := range pick.Chosen {
			ids = append(ids, c.InstanceID)
		}
		return effects.PutInLibraryInAnyOrder{
			Chooser:   pick.Chooser,
			Cards:     ids,
			From:      game.ZoneHand,
			Placement: game.LibraryPlaceTop,
			Label:     "Agonizing Memories — put them on top of that player's library in any order",
		}.Apply(ctx)
	})
