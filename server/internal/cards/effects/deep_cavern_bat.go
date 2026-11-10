package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deep-Cavern Bat — Creature — Bat {1}{B}, 1/1 (Lost Caverns of Ixalan,
// #4470):
//
//	"Flying, lifelink
//	 When this creature enters, look at target opponent's hand. You may
//	 exile a nonland card from it until this creature leaves the battlefield."
//
// Word of Command's look (only the controller learns the hand, CR 701.20e),
// then Cloak and Dagger, Entwined's linked hand exile and return: the card is
// exiled linked to the Bat, and the Bat leaving the battlefield returns it to
// its owner's hand (CR 610.3). If the Bat left before the trigger resolved the
// hand is still looked at and nothing is exiled (CR 610.3b, the ruling).
//
// No simplification.
const deepCavernBatHeldLabel = "Deep-Cavern Bat — the exiled card returns when it leaves the battlefield"

func init() {
	enters := WhenThisEnters(
		"Deep-Cavern Bat — look at target opponent's hand; you may exile a nonland card from it until Deep-Cavern Bat leaves the battlefield",
		deepCavernBatEnters)
	enters.Targets = TargetPlayer("target opponent", Opponent())
	Register(Spec{
		OracleID:        "e46f6644-6bf1-47b1-817c-906bd3eca99f",
		Name:            "Deep-Cavern Bat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

func deepCavernBatEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := TargetedPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	filter := Nonland()
	// CR 610.3b: the Bat left before this resolved, so only the look happens.
	if ref, ok := ctx.SourceRef(); !ok || g.UntilEventHappenedForEffect(game.UntilReturn{Leaves: ref}, item) {
		filter = func(*game.Game, uuid.UUID, game.Card) bool { return false }
	}
	return ChooseFromRevealedHand{
		Player:   player,
		Filter:   filter,
		Label:    "nonland card",
		Look:     true,
		Optional: true,
		Exile:    true,
		Linked:   true,
		Then:     deepCavernBatHandReturn,
	}.Apply(ctx)
}

// deepCavernBatHandReturn records that the exiled hand card goes back to its
// owner's hand when the Bat leaves the battlefield: the exile linked it to the
// Bat object (CR 607.2a), the object the return waits on.
var deepCavernBatHandReturn = RevealedPickThen("revealed-pick/deep-cavern-bat-return",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		id := pick.Chosen[0].InstanceID
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.ExiledWith.ID == uuid.Nil {
			return nil
		}
		ctx.Game.ScheduleUntilReturnForEffect(id, game.UntilReturn{
			Controller:   pick.Chooser,
			SourceCardID: pick.Source,
			Label:        deepCavernBatHeldLabel,
			Leaves:       game.ObjectRef(c.ExiledWith),
			ToHand:       true,
		})
		return nil
	})
