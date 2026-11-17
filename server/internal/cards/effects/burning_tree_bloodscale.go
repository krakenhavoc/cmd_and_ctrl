package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burning-Tree Bloodscale — Creature — Lizard Berserker {2}{R}{G}, 2/2
// (Guildpact, #5599):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 {2}{R}: Target creature can't block this creature this turn.
//	 {2}{G}: Target creature blocks this creature this turn if able."
//
// Bloodthirst is its own entry replacement (Bloodthirst). The {2}{R} row
// is Wrinkly Monkey Shenanigans' "can't block this creature" record on
// the Bloodscale; the {2}{G} row is Turntimber Basilisk's requirement,
// which waits for the Bloodscale to attack this turn and asks nothing if
// it never does (CR 509.1c). Both read their target again at resolution
// (CR 608.2b). The {2}{R} row declares the can't-block harm for the bot
// (#4977); the engine's record bars only blocks of the Bloodscale.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9be4be0-71fe-41d9-8f30-351385753fc9",
		Name:         "Burning-Tree Bloodscale",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Bloodthirst(1, "Burning-Tree Bloodscale")},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{R}: Target creature can't block this creature this turn.",
				Purpose: ForTargets(CantBlockUntilEOT(0, 0, 0)),
				Cost:    ManaCost("{2}{R}"),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return CantBlockThisCreatureThisTurn{Blocker: id, Attacker: item.SourceCardID, Label: "Burning-Tree Bloodscale — can't block it"}.Apply(ctx)
				},
			},
			{
				Label:   "{2}{G}: Target creature blocks this creature this turn if able.",
				Cost:    ManaCost("{2}{G}"),
				Targets: TargetCreature("target creature"),
				Effect:  TargetBlocksThisCreature,
			},
		},
	})
}
