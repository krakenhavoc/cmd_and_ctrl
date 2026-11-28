package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB794DamageEachCreature is "<this> deals A damage to each creature … and
// an additional B damage to each creature …" (Kaervek's Hex, Tropical Storm):
// every creature takes the sum of the amounts it qualifies for, dealt together
// as one damage event by one source (CR 120.2). `amount` reads each creature
// through the layers as the spell resolves; a zero sum deals nothing to it.
func OldB794DamageEachCreature(ctx *Context, amount func(c game.Card) int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, c := range MatchingBattlefield(ctx, Creature()) {
			n := amount(c)
			if n <= 0 {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: c.InstanceID, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// OldB794SourceIsAttackingOrBlocking is "Activate only if this creature is
// attacking or blocking" (Sawback Manticore).
func OldB794SourceIsAttackingOrBlocking() ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && (c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil)
	}
}

// OldB794DamageEqualToAttackers is "deals damage to target attacking creature
// equal to the number of attacking creatures" (Divine Retribution): the count
// is taken as the spell resolves, the target's own included.
func OldB794DamageEqualToAttackers(_ *game.StackItem, ctx *Context) error {
	t, ok := FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: len(AttackingCreatures(ctx.Game))}.Apply(ctx)
}

// OldB794BuildersBane is the resolution of Builder's Bane: destroy the X
// targets as one event, then each player takes damage equal to the number of
// artifacts they controlled that went to a graveyard this way (CR 603.10a: the
// controller is read from the pre-move copy; a token counts, since it is put
// into the graveyard for a moment, CR 111.7; one a replacement sent elsewhere
// does not).
func OldB794BuildersBane(_ *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || ctx.IsNewSourceObject(t.ID) {
			continue
		}
		dup := false
		for _, id := range ids {
			dup = dup || id == t.ID
		}
		if !dup {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DestroyPermanentsKeyedForEffect(ids, ctx.ItemTail(oldB794BuildersBaneDamage))
}

var oldB794BuildersBaneDamage = MassTail("mass/old-b794-builders-bane-damage", func(ctx *Context, swept []game.Card, _ int) error {
	var inGraveyard []game.Card
	for _, c := range swept {
		z := ctx.Game.FindCardZoneForEffect(c.InstanceID)
		if IsToken(c) || (z != nil && z.Kind == game.ZoneGraveyard) {
			inGraveyard = append(inGraveyard, c)
		}
	}
	counts := ControllersOf(inGraveyard)
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, p := range ctx.Game.Seats {
			if p == nil || p.Eliminated || counts[p.ID] == 0 {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: p.ID, Amount: counts[p.ID]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
})

// OldB794DamageMinusOneFromSpells is Benevolent Unicorn's "If a spell would
// deal damage to a permanent or player, it deals that much damage minus 1 to
// that permanent or player instead": a replacement (CR 614.1a), not a
// prevention, on any spell's damage to any player or permanent. The source is
// read as the damage event snapshotted it (CR 608.2h).
func OldB794DamageMinusOneFromSpells(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: false,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			ch, ok := DamageSourceCharacteristics(ev, g)
			if !ok || !ch.SourceIsSpell {
				return false
			}
			if g.PlayerByIDForEffect(ev.DamageTarget) != nil {
				return true
			}
			_, onBattlefield := g.LookupCardForEffect(ev.DamageTarget)
			return onBattlefield && g.Battlefield.Contains(ev.DamageTarget)
		},
		Replace:    S82OldB317PreventOneOfThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      name + " — damage from a spell is reduced by 1",
	}
}
