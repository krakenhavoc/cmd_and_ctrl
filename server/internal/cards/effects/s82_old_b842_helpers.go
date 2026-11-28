package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b842_helpers.go — Ice Age removal cards (S82 OLD batch 842, #5636).

// S82OldB842DestroyArtifactsCreaturesAndLands is "Destroy all artifacts, creatures, and
// lands. They can't be regenerated." (Obliterate, Jokulhaups).
func S82OldB842DestroyArtifactsCreaturesAndLands(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{
		Match:             Or(Artifact(), Creature(), Land()),
		CantBeRegenerated: true,
	}.Apply(ctx)
}

// s82OldB842StenchDamage is the decline branch of Stench of Evil's per-land "unless
// they pay {2}": the land's controller, who is the one asked, takes 1 damage from the sorcery.
// An on-disk key: never renamed, never reused.
var s82OldB842StenchDamage = PayUnlessElse("pay-unless/s82-old-b842-stench-of-evil-damage", func(ctx *Context, r game.PayUnlessAnswer) error {
	return DealDamage{Source: r.Source, Target: r.Chooser, Amount: 1}.Apply(ctx)
})

// s82OldB842StenchTail is Stench of Evil's "for each land destroyed this way": each destroyed
// land's controller (as the land last existed, CR 603.10) is asked in turn.
var s82OldB842StenchTail = MassTail("mass/s82-old-b842-stench-of-evil", func(ctx *Context, swept []game.Card, _ int) error {
	for _, c := range swept {
		if !c.IsLand() {
			continue
		}
		if err := (PayUnless{
			Chooser:  c.Controller,
			Cost:     "{2}",
			Question: "Stench of Evil — pay {2}? (Otherwise it deals 1 damage to you)",
			Else:     s82OldB842StenchDamage,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB842StenchOfEvil is "Destroy all Plains. For each land destroyed this way, this spell
// deals 1 damage to that land's controller unless they pay {2}."
func S82OldB842StenchOfEvil(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{Match: Subtype("Plains"), ThenTail: ctx.Tail(s82OldB842StenchTail)}.Apply(ctx)
}

// S82OldB842DestroyTargetCantRegenerate is "Destroy target <permanent>. It can't be regenerated."
func S82OldB842DestroyTargetCantRegenerate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return DestroyTarget{Target: t.ID, CantBeRegenerated: true}.Apply(ctx)
		}
	}
	return nil
}

// S82OldB842DestroyCombatPartners is Kjeldoran Frostbeast's end-of-combat trigger: "destroy all
// creatures blocking or blocked by this creature" (regeneration and indestructible apply).
func S82OldB842DestroyCombatPartners(g *game.Game, item *game.StackItem) error {
	ids := OldB822CombatPartners(NewContext(g, item))
	if len(ids) == 0 {
		return nil
	}
	g.DestroyPermanentsForEffect(ids)
	return nil
}

// S82OldB842TotalWarDestroy is Total War's trigger body: the attacking player's untapped
// non-Wall creatures that didn't attack are destroyed, except those they haven't controlled
// continuously since the beginning of the turn. Read as the trigger resolves.
func S82OldB842TotalWarDestroy(g *game.Game, item *game.StackItem) error {
	who := item.Trigger.Event.Actor
	if who == uuid.Nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != who || !c.IsCreature() || c.Tapped || c.SummonedThisTurn {
			continue
		}
		if c.HasSubtypeWord("Wall") || g.AttackedThisTurn(c.InstanceID) {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	g.DestroyPermanentsForEffect(ids)
	return nil
}

// s82OldB842DestroyLandThen is "Destroy target land. If that land was a snow land, <then>":
// the land is read before it is destroyed, and the rider is not conditional on the
// destruction succeeding. A target that is gone fizzles (CR 608.2b).
func s82OldB842DestroyLandThen(ctx *Context, then func(ctx *Context, controller uuid.UUID) error) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		land, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		wasSnow, controller := land.HasSupertype("snow"), land.Controller
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		if wasSnow {
			return then(ctx, controller)
		}
		return nil
	}
	return nil
}

// S82OldB842Icequake is Icequake's resolution: destroy the land; if it was snow, 1 damage to its controller.
func S82OldB842Icequake(_ *game.StackItem, ctx *Context) error {
	return s82OldB842DestroyLandThen(ctx, func(ctx *Context, controller uuid.UUID) error {
		return DealDamage{Source: ctx.Source(), Target: controller, Amount: 1}.Apply(ctx)
	})
}

// S82OldB842Thermokarst is Thermokarst's resolution: destroy the land; if it was snow, you gain 1 life.
func S82OldB842Thermokarst(_ *game.StackItem, ctx *Context) error {
	return s82OldB842DestroyLandThen(ctx, func(ctx *Context, _ uuid.UUID) error {
		return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	})
}

// s82OldB842EssenceVortexDestroy is the decline branch of Essence Vortex: the creature is
// destroyed and can't be regenerated.
var s82OldB842EssenceVortexDestroy = PayUnlessElse("pay-unless/s82-old-b842-essence-vortex-destroy", func(ctx *Context, r game.PayUnlessAnswer) error {
	if len(r.Carry) < 1 {
		return nil
	}
	creature := r.Carry[0]
	if !OnBattlefield(ctx.Game, creature) {
		return nil
	}
	return DestroyTarget{Target: creature, CantBeRegenerated: true}.Apply(ctx)
})

// S82OldB842EssenceVortex is "Destroy target creature unless its controller pays life equal to its
// toughness. A creature destroyed this way can't be regenerated." The toughness is read as the
// spell resolves; a controller who can't pay the life loses the creature (CR 119.4).
func S82OldB842EssenceVortex(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		g.RecomputeLayersIfStaleLocked()
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		return B23PayLifeUnless{
			Chooser:  c.Controller,
			Life:     max(c.CurrentToughness(), 0),
			Question: "Essence Vortex — pay life equal to " + c.Name + "'s toughness? (Otherwise it is destroyed)",
			Else:     s82OldB842EssenceVortexDestroy,
			Carry:    []uuid.UUID{t.ID},
		}.Apply(ctx)
	}
	return nil
}

// S82OldB842MinionOfLeshracUpkeep is "At the beginning of your upkeep, this creature deals 5 damage
// to you unless you sacrifice a creature other than this creature. If this creature deals damage to
// you this way, tap it." The sacrifice is offered only when there is another creature to give.
func S82OldB842MinionOfLeshracUpkeep(g *game.Game, item *game.StackItem) error {
	if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	pay := SacrificePayment(1, "creature other than this creature", "creatures other than this creature", QueryType("creature"))
	except := item.SourceCardID
	pay.Action.Except = &except
	return UpkeepPayUnless{
		Chooser:  item.Controller,
		Cost:     pay.Words,
		Action:   pay.Action,
		Question: "Minion of Leshrac — sacrifice another creature, or take 5 damage?",
		Else:     s82OldB842MinionDamage,
	}.Apply(ctx)
}

// s82OldB842MinionDamage is the decline branch: 5 damage to you, then tap the Minion if it was dealt.
var s82OldB842MinionDamage = PayUnlessElse("pay-unless/s82-old-b842-minion-of-leshrac-damage", func(ctx *Context, r game.PayUnlessAnswer) error {
	g := ctx.Game
	ref, ok := g.PermanentRefForEffect(r.Source)
	if !ok {
		// The Minion left: it dealt its damage as it last existed, and has nothing to tap.
		return DealDamage{Source: r.Source, Target: r.Chooser, Amount: 5}.Apply(ctx)
	}
	before := g.DamageDealtToPlayerThisTurnByObjectForEffect(ref, r.Chooser)
	if err := (DealDamage{Source: r.Source, Target: r.Chooser, Amount: 5}).Apply(ctx); err != nil {
		return err
	}
	if g.DamageDealtToPlayerThisTurnByObjectForEffect(ref, r.Chooser) > before && OnBattlefield(g, r.Source) {
		return g.TapTargetForEffect(r.Source)
	}
	return nil
})

// S82OldB842EssenceVortexResolve adapts the effect body to a spell's resolution.
func S82OldB842EssenceVortexResolve(item *game.StackItem, ctx *Context) error {
	return S82OldB842EssenceVortex(ctx.Game, item)
}

// S82OldB842DestroyAllWhitePermanents is Anarchy: "Destroy all white permanents."
func S82OldB842DestroyAllWhitePermanents(_ *game.StackItem, ctx *Context) error {
	return DestroyAllMatching{Match: OfColor("W")}.Apply(ctx)
}

// S82OldB842DestroyAllEnchantments is "Destroy all enchantments." (Essence Filter's first choice).
func S82OldB842DestroyAllEnchantments(_ *game.StackItem, ctx *Context, _ int) error {
	return DestroyAllMatching{Match: Enchantment()}.Apply(ctx)
}

// S82OldB842DestroyAllNonwhiteEnchantments is "Destroy all nonwhite enchantments." (Essence Filter's second choice).
func S82OldB842DestroyAllNonwhiteEnchantments(_ *game.StackItem, ctx *Context, _ int) error {
	return DestroyAllMatching{Match: And(Enchantment(), Not(OfColor("W")))}.Apply(ctx)
}
