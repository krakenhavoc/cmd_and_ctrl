package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Combustible Gearhulk — Artifact Creature — Construct {4}{R}{R}, 6/6:
//
//	"First strike
//	 When this creature enters, target opponent may have you draw
//	 three cards. If the player doesn't, you mill three cards, then
//	 this creature deals damage to that player equal to the total mana
//	 value of those cards."
//
// The cleanest statement of #568's yes/no half: the question belongs
// to an OPPONENT, is asked while the trigger resolves, and the branch
// the engine takes is the rest of the card. Until MayChoice gained
// Player addressing (#796) there was no prompt shape that could ask
// it — pay_unless speaks about mana, and a trigger's CR 603.5 "you
// may" belongs to the trigger's controller.
//
// The damage is read off the three cards that were MILLED, not off the
// library beforehand: MillToZoneForEffect hands back what actually
// moved, so a short library mills what it has and the damage is the
// total mana value of those cards (CR 701.13b — you mill as many as
// you can).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3494a575-f68e-4e78-9f6b-142cd8a0edea",
		Name:            "Combustible Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Combustible Gearhulk — target opponent may have you draw three cards", combustibleGearhulkAsk),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}

// combustibleGearhulkAsk is the trigger's body: the chosen opponent
// decides.
//
// Caller holds g.mu.
func combustibleGearhulkAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return MayChoice{
			Player:   t.ID,
			Question: "Combustible Gearhulk — let its controller draw three cards?",
			YesLabel: "They draw three",
			NoLabel:  "They mill three and it deals damage to you",
			OnYes:    combustibleGearhulkDrawThree,
			OnNo:     combustibleGearhulkMillThenBurn,
		}.Apply(ctx)
	}
	// CR 608.2b: the target is gone, so the trigger does nothing at
	// all — not the "doesn't" branch, which would burn a player who
	// was never asked.
	return nil
}

// combustibleGearhulkDrawThree is the yes branch.
//
// Caller holds g.mu.
func combustibleGearhulkDrawThree(ctx *Context) error {
	return DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
}

// combustibleGearhulkMillThenBurn is the "if the player doesn't"
// branch: mill three and burn the chooser for their total mana value.
//
// The damaged player is read back off the item's target rather than
// captured, so an undo that replays this branch reads the same ref the
// trigger was put on the stack with.
//
// Caller holds g.mu.
func combustibleGearhulkMillThenBurn(ctx *Context) error {
	milled, err := ctx.Game.MillToZoneForEffect(ctx.Controller(), 3, game.ZoneGraveyard, nil)
	if err != nil {
		return err
	}
	total := 0
	for _, id := range milled {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			total += c.ManaValue()
		}
	}
	if total <= 0 {
		return nil
	}
	for _, t := range ctx.Targets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), t.ID, total)
	}
	return nil
}
