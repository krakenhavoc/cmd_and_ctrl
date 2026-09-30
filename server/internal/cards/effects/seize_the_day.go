package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seize the Day — Sorcery {3}{R}:
//
//	"Untap target creature. After this main phase, there is an
//	 additional combat phase followed by an additional main phase.
//	 Flashback {2}{R}"
//
// Relentless Assault's phases (ADR 0059 sub-PR 2b, #753) with a single
// untap, and flashback from the graveyard. The phases go on the turn
// the spell resolves in, and only from a main phase (the ruling).
// With its one target gone at resolution the spell does not resolve
// (CR 608.2b), so no phases are added either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2a6abd59-e448-46f1-9af8-bb9040645971",
		Name:             "Seize the Day",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return ExtraCombatAndMainAfterThisMain().Apply(ctx)
		},
	})
}
