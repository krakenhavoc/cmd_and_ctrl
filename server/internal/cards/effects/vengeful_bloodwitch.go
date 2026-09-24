package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Bloodwitch — Creature — Vampire Warlock {1}{B}, 1/1
// (EDHREC rank 1775):
//
//	"Whenever this creature or another creature you control dies,
//	 target opponent loses 1 life and you gain 1 life."
//
// A Zulaport Cutthroat that picks its victim. The condition is
// b16SelfOrAnotherCreatureYouControlDied — the Bloodwitch's own
// death by ID (cardDied, read from the graveyard), or any creature
// its controller controlled dying; the drain is a targeted trigger
// with an opponent clause, so the controller picks the opponent when
// it goes on the stack and the drain fizzles on a player who left.
// One trigger per creature, so a board wipe drains once per body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4cdbc466-42fc-471f-beab-397caec18101",
		Name:         "Vengeful Bloodwitch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16SelfOrAnotherCreatureYouControlDied(ev, source, g)
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Vengeful Bloodwitch — target opponent loses 1 life and you gain 1 life",
					drainTargetOpponentOne)
			},
		}},
	})
}
