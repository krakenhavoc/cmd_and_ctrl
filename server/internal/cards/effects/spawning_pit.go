package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spawning Pit — Artifact {2} (EDHREC rank 4294):
//
//	"Sacrifice a creature: Put a charge counter on this artifact.
//	 {1}, Remove two charge counters from this artifact: Create a 2/2
//	 colorless Spawn artifact creature token."
//
// A two-mana free sacrifice outlet that also, eventually, pays you
// back. The aristocrats deck plays it for the first ability alone —
// no mana, no tap, activate as many times as you have creatures, which
// is the whole shopping list for a Blood Artist turn. The second
// ability is the tax that makes the first one fair.
//
// The sacrifice is a COST (CR 601.2h), so it is paid at announce: the
// creature is gone and its dies-trigger is on the stack before the
// counter ability resolves, and a response cannot save it. "A
// creature" is any creature its controller controls — the Pit is an
// artifact and never a candidate for its own cost, so no "another" is
// needed.
//
// The counter removal is also a cost and is paid at announce (#625),
// so two activations in the same window cannot spend the same two
// counters. A Pit with fewer than two charge counters simply does not
// offer the ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea7b0704-7715-4f8a-b32f-d2d2f3b36054",
		Name:         "Spawning Pit",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "Sacrifice a creature: Put a charge counter on this artifact",
				Cost:  SacrificeACreature(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddCounter{
						Target: item.SourceCardID,
						Kind:   game.CounterCharge,
						N:      1,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "{1}, Remove two charge counters from this artifact: Create a 2/2 colorless Spawn artifact creature token",
				Cost:  Plus(ManaCost("{1}"), RemoveCountersFromThis(game.CounterCharge, 2)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 colorless Spawn artifact"),
						N:          1,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
