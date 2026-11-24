package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amy Rose — Legendary Creature — Hedgehog Warrior {2}{R}{W}, 3/3 (Secret
// Lair Drop, #5462):
//
//	"Haste
//	 Whenever Amy Rose attacks, attach up to one target Equipment to her.
//	 Then up to one other target attacking creature gets +X/+0 until end of
//	 turn, where X is Amy Rose's power."
//
// Two "up to one" target clauses chosen as the trigger goes on the stack: any
// Equipment (Sokka and Suki's attach), and another attacking creature. As it
// resolves the attach happens first, so X is Amy Rose's power AFTER any
// Equipment is on her, read as the pump is applied (CR 608.2c); the pump is a
// fixed +X/+0 until end of turn with X settled then. Either half does nothing
// if its target has become illegal (CR 608.2b), and neither depends on the
// other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2d597b4-f9ba-44b0-a947-0c2ecda2d308",
		Name:            "Amy Rose",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Amy Rose — attach up to one target Equipment to her; then up to one other target attacking creature gets +X/+0 until end of turn, where X is her power",
					amyRoseAttack),
				effects.Clauses(
					effects.TargetPermanent("up to one target Equipment", effects.HasSubtype("Equipment")).WithCount(0, 1),
					effects.Another(effects.TargetCreature("up to one other target attacking creature", effects.IsAttackingCreature())).WithCount(0, 1),
				)),
		},
	})
}

func amyRoseAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self := item.SourceCardID
	if equip, ok := ctx.ClauseTarget(0); ok && equip.Kind == game.TargetCard && effects.OnBattlefield(g, self) {
		if err := g.AttachForEffect(equip.ID, game.TargetRef{Kind: game.TargetCard, ID: self}); err != nil {
			return err
		}
	}
	pumped, ok := ctx.ClauseTarget(1)
	if !ok || pumped.Kind != game.TargetCard {
		return nil
	}
	amy, ok := g.LookupCardForEffect(self)
	if !ok || amy.CurrentPower() <= 0 {
		return nil
	}
	return effects.BoostUntilEOT{
		Target: pumped.ID,
		Power:  amy.CurrentPower(),
		Label:  "Amy Rose — +X/+0 until end of turn",
	}.Apply(ctx)
}
