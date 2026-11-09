package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightmare Sower — Creature — Faerie Assassin {3}{B}, 2/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Flying, lifelink
//	 Whenever you cast a spell during an opponent's turn, put a -1/-1
//	 counter on up to one target creature."
//
// Katara, Waterbending Master's cast condition (the caster is the Sower's
// controller and is not the active player). The target is "up to one", any
// creature, chosen as the trigger goes on the stack; with none chosen, or
// one that left, it does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4121b7d9-5fbe-40e2-a2a6-d7633f423b5f",
		Name:            "Nightmare Sower",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventCast, stdB64YouCastDuringAnOpponentsTurn,
					"Nightmare Sower — put a -1/-1 counter on up to one target creature", nightmareSowerCounter),
				UpToOneTargetCreature("up to one target creature")),
		},
	})
}

func nightmareSowerCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
}
