package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Price of Betrayal — Sorcery {B} (War of the Spark, #5520):
//
//	"Remove up to five counters from target artifact, creature, planeswalker,
//	 or opponent."
//
// Etched Slith's counter removal (#6100), a permanent's or a player's counters
// of any kind, up to five in all, chosen kind by kind as the spell resolves
// (CR 608.2c). A target without counters is still targeted and nothing comes
// off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43a2bfe4-9705-4d0f-901a-21f9803b50f1",
		Name:         "Price of Betrayal",
		Completeness: effects.CompletenessFull,
		Targets:      priceOfBetrayalTargets(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard || t.Kind == game.TargetPlayer {
					return effects.S82OldB170RemoveCounters.Choose(ctx, t.ID, nil, 5, "Price of Betrayal — remove up to five counters")
				}
			}
			return nil
		},
	})
}

// priceOfBetrayalTargets is "target artifact, creature, planeswalker, or
// opponent".
func priceOfBetrayalTargets() *game.TargetSpec {
	opponent := effects.Opponent()
	return &game.TargetSpec{
		Mode:     "any",
		Label:    "target artifact, creature, planeswalker, or opponent",
		Players:  true,
		PlayerOK: func(g *game.Game, caster uuid.UUID, p *game.Player) bool { return opponent(g, caster, p) },
		Zones:    []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsArtifact() || c.IsCreature() || c.IsPlaneswalker()
		},
		Min: 1, Max: 1,
	}
}
