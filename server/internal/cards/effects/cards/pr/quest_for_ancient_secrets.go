package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quest for Ancient Secrets — Enchantment {U} (Zendikar, S82 OLD b493):
//
//	"Whenever a card is put into your graveyard from anywhere, you may put a quest
//	 counter on this enchantment.
//	 Remove five quest counters from this enchantment and sacrifice it: Target
//	 player shuffles their graveyard into their library."
//
// The counter trigger reads every route a card takes into its owner's graveyard
// (GraveyardArrivals{}, graveyard_arrival.go): discard, mill and every other
// zone move, a token never. The activation targets a player and shuffles that
// player's graveyard into their library (Psychic Spiral's move, Soul-Scar's
// shuffle through the engine's own seam).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7f7b882-90cb-4031-8a57-8e721bc905ac",
		Name:         "Quest for Ancient Secrets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.OnAny(effects.GraveyardArrivalKinds, effects.GraveyardArrivals{}.When(),
				"Quest for Ancient Secrets — put a quest counter",
				effects.Do(effects.CounterOnThis{Kind: "quest", N: 1})), "Quest for Ancient Secrets — put a quest counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove five quest counters from this enchantment and sacrifice it: Target player shuffles their graveyard into their library.",
			Cost:    effects.Plus(effects.RemoveCountersFromThis("quest", 5), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
					return nil
				}
				return effects.FinaleShuffleGraveyardIntoLibrary(ctx, item.Targets[0].ID)
			},
		}},
	})
}
