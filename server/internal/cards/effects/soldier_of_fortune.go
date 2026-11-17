package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soldier of Fortune — Creature — Human Mercenary {R}, 1/1 (Alliances,
// #5634):
//
//	"{R}, {T}: Target player shuffles their library."
//
// The {T} keeps it off the turn it arrives (CR 302.6). The shuffle
// randomizes the target player's library (CR 701.24a); a player who has
// left the game by the time it resolves is not a legal target, and the
// ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b2367b3-3a8d-442f-8d58-6e4c03fa5a46",
		Name:         "Soldier of Fortune",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: Target player shuffles their library.",
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if err := (ShuffleLibrary{Player: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
