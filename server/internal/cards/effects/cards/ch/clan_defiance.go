package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clan Defiance — Sorcery {X}{R}{G} (Ravnica: Clue Edition):
//
//	"Choose one or more —
//	 • Clan Defiance deals X damage to target creature with flying.
//	 • Clan Defiance deals X damage to target creature without flying.
//	 • Clan Defiance deals X damage to target player or planeswalker."
//
// Each chosen bullet has its own target (CR 700.2) and is resolved in
// printed order (CR 608.2c). A bullet whose target has gone does nothing
// and the others still resolve. Flying is read when the spell resolves or is
// targeted through the layered view.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "abcf5e77-88fc-4b76-a1e8-907a5b27d612",
		Name:         "Clan Defiance",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOneOrMore(
			effects.ModeWithPurpose(effects.ModeDoing("Clan Defiance deals X damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")), clanDefianceDeal),
				effects.ForTargets(effects.DamageXToTarget(0))),
			effects.ModeWithPurpose(effects.ModeDoing("Clan Defiance deals X damage to target creature without flying.",
				effects.TargetCreature("target creature without flying", effects.WithoutKeyword("flying")), clanDefianceDeal),
				effects.ForTargets(effects.DamageXToTarget(0))),
			effects.ModeWithPurpose(effects.ModeDoing("Clan Defiance deals X damage to target player or planeswalker.",
				effects.TargetPlayerOrPlaneswalker(), clanDefianceDeal),
				effects.ForTargets(effects.DamageXToTarget(0))),
		),
	})
}

// clanDefianceDeal is one bullet: X damage, the announced X, to that
// bullet's target if it is still legal (CR 608.2b).
func clanDefianceDeal(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}.Apply(ctx)
}
