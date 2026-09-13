package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arbor Elf — Creature — Elf Druid {G}, 1/1 (EDHREC rank 601):
//
//	"{T}: Untap target Forest."
//
// The one-drop that is a mana dork only in a deck full of Forests —
// and a two-mana dork under Utopia Sprawl. It is NOT a mana ability
// (it targets, CR 605.1a), so it is an ordinary CR 602 activated
// ability: it uses the stack, it can be responded to, and it is gated
// by summoning sickness like every tap ability on a creature.
//
// "Forest" is the land TYPE, read post-layer (#354): a Forest made by
// Yavimaya, or a nonbasic with the type, is a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4567a528-75f0-4ea6-b927-3a500caf76ac",
		Name:         "Arbor Elf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap target Forest.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target Forest", b05HasSubtype("Forest")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return UntapTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
