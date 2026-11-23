package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sevinne's Reclamation — Sorcery {2}{W} (EDHREC rank 341):
//
//	"Return target permanent card with mana value 3 or less from your
//	 graveyard to the battlefield. If this spell was cast from a
//	 graveyard, you may copy this spell and may choose a new target
//	 for the copy.
//	 Flashback {4}{W}"
//
// Three already-built primitives: `Flashback` (alternative_cost.go)
// for the cast path, `StackItem.CastFromZone` (#761, wash_away.go's
// own field) to read "cast from a graveyard" the same way
// Increasing Vengeance's doubled branch does, and `copyThisSpellFor`
// (chain_of_vapor.go) — the shared "may copy this spell and may
// choose a new target for the copy" body #920 built and Chain of
// Smog already reuses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b9f9f5b-8712-4f00-90cb-1b7b9970eccc",
		Name:         "Sevinne's Reclamation",
		Completeness: effects.CompletenessFull,
		CastableZones: []game.ZoneKind{
			game.ZoneGraveyard,
		},
		AlternativeCosts: []game.AlternativeCost{
			effects.Flashback("{4}{W}"),
		},
		Targets: effects.TargetCardInGraveyard(
			"target permanent card with mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) != 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (effects.ReturnFromGraveyard{
					Target:     item.Targets[0].ID,
					Dest:       game.ZoneBattlefield,
					Controller: item.Controller,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if item.CastFromZone != game.ZoneGraveyard {
				return nil
			}
			return effects.MayChoice{
				Player:   item.Controller,
				Question: "Sevinne's Reclamation — copy it? (you may choose a new target for the copy)",
				Then:     effects.MayCopyThisSpell,
			}.Apply(ctx)
		},
	})
}
