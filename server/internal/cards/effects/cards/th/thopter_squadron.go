package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thopter Squadron — Artifact Creature — Thopter {5}, 0/0 (Exodus, S82 OLD batch 741, #5626):
//
//	"Flying
//	 This creature enters with three +1/+1 counters on it.
//	 {1}, Remove a +1/+1 counter from this creature: Create a 1/1 colorless
//	 Thopter artifact creature token with flying. Activate only as a sorcery.
//	 {1}, Sacrifice another Thopter: Put a +1/+1 counter on this creature.
//	 Activate only as a sorcery."
//
// The Spike cycle's entry replacement (spike_counters.go). Both abilities are
// sorcery-speed; the counter is a cost (CR 118.3); the sacrifice names any
// other Thopter you control, token or card. The counter goes on the Squadron
// only if it is still the same permanent as the ability resolves (CR 400.7).
//
// No simplification.
func init() {
	const name = "Thopter Squadron"
	effects.Register(effects.Spec{
		OracleID:     "f1482ca4-610b-48fe-ac23-4526637bb72a",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "three", 3)},
		Activated: []effects.ActivatedAbility{
			{
				Label: "{1}, Remove a +1/+1 counter from this creature: Create a 1/1 colorless Thopter artifact creature token with flying. " +
					"Activate only as a sorcery.",
				Purpose:      game.Purpose{Tokens: 1},
				Cost:         effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1}.Apply(ctx)
				},
			},
			{
				Label:        "{1}, Sacrifice another Thopter: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another Thopter", effects.HasSubtype("Thopter"))),
				SorcerySpeed: true,
				Effect:       effects.ThisStillHere(effects.PlusOneCountersOnThis(1)),
			},
		},
		PrintedKeywords: []string{"flying"},
	})
}
