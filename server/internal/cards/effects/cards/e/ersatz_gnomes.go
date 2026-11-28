package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ersatz Gnomes — Artifact Creature — Gnome {3}, 1/1 (Mirage, S82 OLD b808):
//
//	"{T}: Target spell becomes colorless.
//	 {T}: Target permanent becomes colorless until end of turn."
//
// Two separate {T} abilities, each a layer 5 change to no colour (CR 613.1e)
// over the one target, still legal as it resolves (CR 608.2b). The spell's has
// no stated duration (CR 611.2a), so it is a record pinned to the spell on
// the stack that lasts indefinitely and, for a permanent spell, carries over
// to the permanent it becomes (CR 400.7a), as Blind Seer's does for the length
// of its own duration. The permanent's lasts until end of turn and is an
// ordinary colour record locked to that object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea016b84-7f77-42e7-bd14-2e62a1e35f3b",
		Name:         "Ersatz Gnomes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target spell becomes colorless.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetSpell("target spell"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.OldB808TargetSpellBecomesColorless(g, item)
				},
			},
			{
				Label:   "{T}: Target permanent becomes colorless until end of turn.",
				Cost:    effects.TapCost(),
				Targets: effects.TargetPermanent("target permanent"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Gap2B05TargetBecomesColor(effects.NewContext(g, item), nil, "Ersatz Gnomes — colorless until end of turn")
				},
			},
		},
	})
}
