package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brago, King Eternal — Legendary Creature — Spirit Noble {2}{W}{U}, 2/4:
//
//	"Flying
//	 Whenever Brago deals combat damage to a player, exile any number of
//	 target nonland permanents you control, then return those cards to
//	 the battlefield under their owner's control."
//
// The trigger targets (CR 603.3d), any number including none, and Brago
// may target itself. Every target still legal on resolution (CR 608.2b)
// leaves as ONE exile and the ones that reached exile come back as ONE
// entry (ReturnFromExileTogether), so each sees the others enter
// (CR 603.6a) and each is a new object (CR 400.7): untapped, no
// counters, auras that were on it gone, every ETB re-triggered. "Under
// their owner's control" is the printed point of the card in a table
// with control magic: a permanent you stole comes back to its owner. A
// token exiled this way ceases to exist (CR 111.8). A commander among
// the targets pauses on its owner's CR 903.9 question; the return waits
// on ExileCardsKeyedForEffect's continuation for the cards that really
// reached exile.
//
// No simplification. Roon deck request (#4128).
func init() {
	Register(Spec{
		OracleID:        "fcb0c7db-bd07-4c25-b7f8-b6c207e4f6be",
		Name:            "Brago, King Eternal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer("Brago, King Eternal — exile your permanents, then return them", bragoBlinkTargets),
				TargetPermanent("any number of target nonland permanents you control", Nonland(), YouControl()).WithCount(0, 0),
			),
		},
	})
}

// bragoBlinkTargets exiles every still-legal target at once and returns
// the cards that reached exile under their owners' control, together.
func bragoBlinkTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(bragoReturn))
}

// bragoReturn returns the cards that reached exile, together (#4253 slice 9).
var bragoReturn = CardTail("brago-king-eternal/return-exiled", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs

	return ReturnFromExileTogether{Targets: exiled}.Apply(NewContext(g, item))
})
