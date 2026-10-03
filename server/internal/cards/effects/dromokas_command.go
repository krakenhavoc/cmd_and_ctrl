package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dromoka's Command — Instant {G}{W}:
//
//	"Choose two —
//	 • Prevent all damage target instant or sorcery spell would deal this turn.
//	 • Target player sacrifices an enchantment of their choice.
//	 • Put a +1/+1 counter on target creature.
//	 • Target creature you control fights target creature you don't control."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the first bullet is a
// preventFromSource record whose source is the targeted spell, pinned on
// the stack as the Command resolves. Damage a fight or another object
// deals because of that spell is not the spell's, so it is not prevented
// (its ruling).
//
// The bullets run in printed order (CR 608.2c). The sacrifice is the
// target player's choice and they may have to answer a prompt, so the
// two bullets after it run once the enchantment has gone: an Aura they
// give up is off the creature before it fights.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3bd8ea71-cbcc-4659-b9a8-88cf27ee12d8",
		Name:         "Dromoka's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			Mode("Prevent all damage target instant or sorcery spell would deal this turn.",
				instantOrSorcerySpell("target instant or sorcery spell")),
			Mode("Target player sacrifices an enchantment of their choice.", TargetPlayer("target player")),
			Mode("Put a +1/+1 counter on target creature.", TargetCreature("target creature")),
			Mode("Target creature you control fights target creature you don't control.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
				)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range OptionTargets(ctx, 0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (PreventDamageFromSource{From: t.ID, Protect: ShieldAnything}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(1) {
				for _, t := range OptionTargets(ctx, 1) {
					if t.Kind == game.TargetPlayer {
						return dromokasCommandSacrifice(ctx, t.ID)
					}
				}
			}
			return dromokasCommandRest(ctx)
		},
	})
}

// dromokasCommandSacrifice is the second bullet: the player chooses one
// of their enchantments and sacrifices it, and the later bullets follow
// once it has gone. A player with no enchantment is asked nothing.
func dromokasCommandSacrifice(ctx *Context, player uuid.UUID) error {
	return ChoosePermanents{
		Player:   player,
		Of:       []uuid.UUID{player},
		Question: "Dromoka's Command — sacrifice an enchantment",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsEnchantment() {
					ids = append(ids, c.InstanceID)
				}
			}
			return ids, 1, 1
		},
		Sacrifice: true,
		Then: func(ctx *Context, _ game.PromptedPicks) error {
			return dromokasCommandRest(ctx)
		},
	}.Apply(ctx)
}

// dromokasCommandRest is the third and fourth bullets, in printed order.
func dromokasCommandRest(ctx *Context) error {
	if ctx.HasMode(2) {
		for _, t := range OptionTargets(ctx, 2) {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	for occ, m := range ctx.Modes() {
		if m == 3 {
			return FightTheModesTargets(ctx.Item, ctx, occ)
		}
	}
	return nil
}
