package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edifice of Authority — Artifact {3} (Jumpstart 2022, #5948):
//
//	"{1}, {T}: Target creature can't attack this turn. Put a brick counter
//	 on this artifact.
//	 {1}, {T}: Until your next turn, target creature can't attack or block
//	 and its activated abilities can't be activated. Activate only if there
//	 are three or more brick counters on this artifact."
//
// The counter is put on as the first ability resolves, so a target that has
// become illegal stops both halves (CR 608.2b). The second ability's
// condition is checked when it is activated (CR 602.5b) and its effect lasts
// until the start of your next turn; "activated abilities" includes mana
// abilities (CR 605.1a), so both restriction bits are set.
//
// No simplification.
func init() {
	const brick = "brick"
	effects.Register(effects.Spec{
		OracleID:     "7c16707e-24da-4a7e-9d24-d3731987f675",
		Name:         "Edifice of Authority",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Target creature can't attack this turn. Put a brick counter on this artifact.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantAttack,
						Label: "Edifice of Authority — can't attack this turn"}).Apply(ctx); err != nil {
						return err
					}
					return effects.AddCounter{Target: ctx.Source(), Kind: brick, N: 1}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, {T}: Until your next turn, target creature can't attack or block and its activated abilities can't be activated. Activate only if there are three or more brick counters on this artifact.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && c.Counters[brick] >= 3
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana)},
						Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
						Label:    "Edifice of Authority — can't attack or block, and its activated abilities can't be activated, until your next turn",
					}.Apply(ctx)
				},
			},
		},
	})
}
