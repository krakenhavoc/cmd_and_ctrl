package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dwarven Warriors — Creature — Dwarf Warrior {2}{R}, 1/1 (The Hobbit
// Eternal, HOC):
//
//	"{T}: Target creature with power 2 or less can't be blocked this
//	 turn."
//
// Access Tunnel's ability with the cost a bare tap (#4454). The power cap
// rides PowerLE(2) on the target clause and is checked again when the
// ability resolves (CR 608.2b); the creature it names may be the Dwarf
// itself. "Can't be blocked" is a restriction until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfc553cd-3b4c-47a9-bffb-e5790befb32c",
		Name:         "Dwarven Warriors",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with power 2 or less can't be blocked this turn",
			Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with power 2 or less", PowerLE(2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Dwarven Warriors — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
