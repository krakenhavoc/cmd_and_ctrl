package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Green Goblin, Nemesis — Legendary Creature — Goblin Human Villain
// {2}{B}{R}, 3/3 (Marvel's Spider-Man, #4460):
//
//	"Flying
//	 Whenever you discard a nonland card, put a +1/+1 counter on target
//	 Goblin you control.
//	 Whenever you discard a land card, create a tapped Treasure token."
//
// Two triggers on EventDiscardCard, one per printed sentence and one per
// card discarded (Bone Miser's shape). The card is read where it landed,
// so a replaced destination still counts as a discard (CR 701.9a). A
// nonland card names a Goblin target as the trigger goes on the stack
// (Green Goblin is a Goblin himself); a land card makes a tapped Treasure.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d1f1e48f-f1f4-497f-9112-79159d2f443b",
		Name:            "Green Goblin, Nemesis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventDiscardCard, youDiscardedA(func(c game.Card) bool { return !c.IsLand() }),
				"Green Goblin, Nemesis — put a +1/+1 counter on target Goblin you control",
				PlusOneCounterOnTheTarget),
				TargetCreature("target Goblin you control", YouControl(),
					func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasSubtypeWord("Goblin") })),
			On(game.EventDiscardCard, youDiscardedA(game.Card.IsLand),
				"Green Goblin, Nemesis — create a tapped Treasure token",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
