package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cytoshape — Instant {1}{G}{U}:
//
//	"Choose a nonlegendary creature on the battlefield. Target creature
//	 becomes a copy of that creature until end of turn."
//
// Two different selections, and the difference is printed: the
// creature that CHANGES is a target, chosen on cast and re-checked at
// resolution; the creature it copies is CHOSEN as the spell resolves
// (CR 608.2), so it does not target — hexproof and shroud do not stop
// it — and it may be any nonlegendary creature, the target included.
// The choice is a ChooseCards prompt, answered as the spell resolves.
//
// The copy is a duration copy (#1593, become_copy.go). On a Clone it
// ends by giving the Clone back its own entry copy, not by turning it
// into Clone (ADR 0043's 2026-09-28 amendment).
func init() {
	Register(Spec{
		OracleID:     "14221c18-7801-49c9-a2c8-53b8c5181d63",
		Name:         "Cytoshape",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			target := targets[0].ID
			var candidates []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && !c.IsLegendary() {
					candidates = append(candidates, c.InstanceID)
				}
			}
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Cytoshape — choose a nonlegendary creature on the battlefield to copy",
				Cards:    candidates,
				Min:      1,
				Max:      1,
				Zone:     game.ZoneBattlefield,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					if len(picked) == 0 {
						return nil
					}
					return BecomeCopy{
						Targets: []uuid.UUID{target},
						Of:      picked[0],
						Label:   "Cytoshape — becomes a copy until end of turn",
					}.Apply(NewContext(g, item))
				},
			})
			return nil
		},
	})
}
