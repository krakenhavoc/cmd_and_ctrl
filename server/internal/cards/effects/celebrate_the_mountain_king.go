package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Celebrate the Mountain-king — Enchantment {3}{W}:
//
//	"When this enchantment enters, for each opponent, exile up to one
//	 target nonland permanent that player controls until this
//	 enchantment leaves the battlefield.
//	 When this enchantment enters, recruit. (Draw a card, then discard a
//	 card. If you discarded a nonland card, create a 1/1 white Human
//	 Soldier creature token.)"
//
// "For each opponent, up to one target … that player controls" is one
// up-to-one clause per opponent, each bound to that opponent
// (perOpponentClauses), so the CR 608.2b re-check asks whether THAT
// player still controls the pick. Each pick still legal is exiled
// through ExileUntil{ThisLeaves} (CR 610.3), one after another through
// each exile's keyed continuation (Abdel Adrian's chain), so a commander
// prompt an exile pauses on is a restore point. An enchantment that has
// already left by resolution exiles nothing (CR 610.3c).
//
// The two triggers are separate abilities; their controller orders them.
//
// No simplification. The Hobbit (#4475).
func init() {
	celebrateExiledOne = CardTail("then/celebrate-the-mountain-king-exiled-one", func(ctx *Context, o game.TailOutcome) error {
		return celebrateExileNext(ctx, o.Tail.Carry)
	})
	exile := WhenThisEnters("Celebrate the Mountain-king — exile a nonland permanent of each opponent until it leaves",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			var picks []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					picks = append(picks, t.ID)
				}
			}
			return celebrateExileNext(ctx, picks)
		})
	exile.Targets = perOpponentClauses(game.MaxPlayers-1, true, "nonland permanent", Nonland())
	Register(Spec{
		OracleID:     "d51136fa-3c13-48a5-83fd-51fe00010a4b",
		Name:         "Celebrate the Mountain-king",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			exile,
			WhenThisEnters("Celebrate the Mountain-king — recruit", Do(Recruit{})),
		},
	})
}

// celebrateExileNext exiles the first of `picks` until the enchantment
// leaves, the rest riding the exile's continuation as its Carry.
func celebrateExileNext(ctx *Context, picks []uuid.UUID) error {
	if len(picks) == 0 {
		return nil
	}
	id, rest := picks[0], picks[1:]
	t := ctx.ItemTail(celebrateExiledOne)
	t.Carry = append([]uuid.UUID(nil), rest...)
	return ExileUntil{
		Target:     id,
		ThisLeaves: true,
		Label:      "Celebrate the Mountain-king — returns when Celebrate the Mountain-king leaves the battlefield",
		ThenTail:   t,
	}.Apply(ctx)
}

// celebrateExiledOne is one exile settled, then the next pick. An on-disk
// key, registered in init (it refers to celebrateExileNext, which refers
// to it).
var celebrateExiledOne game.TailRef
