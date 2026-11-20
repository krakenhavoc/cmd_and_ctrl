package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "3bd8ea71-cbcc-4659-b9a8-88cf27ee12d8",
		Name:         "Dromoka's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.Mode("Prevent all damage target instant or sorcery spell would deal this turn.",
				effects.TargetInstantOrSorcerySpell("target instant or sorcery spell")),
			effects.Mode("Target player sacrifices an enchantment of their choice.", effects.TargetPlayer("target player")),
			effects.Mode("Put a +1/+1 counter on target creature.", effects.TargetCreature("target creature")),
			effects.Mode("Target creature you control fights target creature you don't control.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				)),
		),
		// Each bullet is a step, in printed order: the sacrifice's question, and the
		// counter and the fight after it, wait for the prompt before (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				for _, t := range effects.OptionTargets(ctx, 0) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.PreventDamageFromSource{From: t.ID, Protect: effects.ShieldAnything}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				for _, t := range effects.OptionTargets(ctx, 1) {
					if t.Kind == game.TargetPlayer {
						return dromokasCommandSacrifice(ctx, t.ID)
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(2) {
					return nil
				}
				for _, t := range effects.OptionTargets(ctx, 2) {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				for occ, m := range ctx.Modes() {
					if m == 3 {
						return effects.FightTheModesTargets(ctx.Item, ctx, occ)
					}
				}
				return nil
			}),
		),
	})
}

// dromokasCommandSacrifice is the second bullet: the player chooses one
// of their enchantments and sacrifices it, and the later bullets are the
// steps after this one, so they follow once it has gone. A player with no enchantment is asked nothing.
func dromokasCommandSacrifice(ctx *effects.Context, player uuid.UUID) error {
	return effects.ChoosePermanents{
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
		Keyed: dromokasCommandSacrificed,
	}.Apply(ctx)
}

// dromokasCommandSacrificed is the chosen enchantment sacrificed, keyed
// (#4290).
var dromokasCommandSacrificed = effects.RegisterPermanentsPick("prompt-run/dromokas-command-sacrifice", effects.PermanentsDef{Sacrifice: true})
