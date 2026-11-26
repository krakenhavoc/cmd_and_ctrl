package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB432AttachOnEnterEquipment is the shape of "Equipped creature gets
// +1/+1. Whenever a creature you control enters, you may attach this Equipment
// to it. Equip {N}" (Sai of the Shinobi): the pump, Stormrider Rig's optional
// attach trigger, and an equip ability.
func S82OldB432AttachOnEnterEquipment(oracleID, name, equipCost string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered:    []game.TriggeredAbility{AttachThisToTheEnteringCreature(name)},
		Activated:    []ActivatedAbility{EquipAbility(equipCost)},
	}
}

// S82OldB432SourceIsEnchanted is "if this creature is enchanted": an Aura is
// attached to the source as the trigger fires and again as it resolves
// (CR 603.4).
var S82OldB432SourceIsEnchanted = CondBoard{
	Name: "this creature is enchanted",
	Test: func(ctx *Context) bool { return IsEnchantedByAura(ctx.Game, ctx.Source()) },
}

// S82OldB432InterveningIfEnchanted gives a trigger the intervening "if this
// creature is enchanted" (CR 603.4): it does not trigger unless an Aura is
// attached as the event happens, and does nothing unless one still is as it
// resolves.
func S82OldB432InterveningIfEnchanted(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && IsEnchantedByAura(g, source.InstanceID) && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(S82OldB432SourceIsEnchanted, t)
}

// S82OldB432DamageEqualToCounters is Preyseizer Dragon's trigger body: it deals
// damage to the target equal to the number of +1/+1 counters on it as the
// ability resolves, or as it last existed if it has left (CR 608.2h).
func S82OldB432DamageEqualToCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := CountersOnThisAsItResolves(game.CounterPlusOne)(ctx)
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// S82OldB432MutualDamage is Predatory Urge's granted ability as it resolves:
// "This creature deals damage equal to its power to target creature. That
// creature deals damage equal to its power to this creature." Both amounts are
// read first, then both are dealt as one damage event (CR 120.2); either
// creature gone leaves the other's half to happen alone, and the target gone
// leaves nothing to deal to (CR 608.2b).
func S82OldB432MutualDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstTargetCard(item)
	if !ok || !ctx.IsTargetLegal(t) {
		return nil
	}
	self, other := item.SourceCardID, t.ID
	selfCard, haveSelf := g.LookupCardForEffect(self)
	otherCard, haveOther := g.LookupCardForEffect(other)
	if !haveOther {
		return nil
	}
	var selfPower, otherPower int
	if haveSelf && OnBattlefield(g, self) {
		selfPower = selfCard.CurrentPower()
	}
	if OnBattlefield(g, other) {
		otherPower = otherCard.CurrentPower()
	}
	return g.DamageInstanceForEffect(func() error {
		if selfPower > 0 {
			if err := (DealDamage{Source: self, Target: other, Amount: selfPower}).Apply(ctx); err != nil {
				return err
			}
		}
		if otherPower > 0 && haveSelf {
			return DealDamage{Source: other, Target: self, Amount: otherPower}.Apply(ctx)
		}
		return nil
	})
}

// S82OldB432SunkenHopeBounce is Sunken Hope's trigger: the player whose upkeep
// it is picks a creature they control and returns it to its owner's hand. With
// none, nothing happens; a player with exactly one is not asked.
func S82OldB432SunkenHopeBounce(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Trigger().Event.Actor
	if player == uuid.Nil {
		return nil
	}
	return ChoosePermanents{
		Player:   player,
		Question: "Sunken Hope — return a creature you control to its owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := PermanentsControlledByMatching(g, of, Creature())
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: s82OldB432SunkenHopePick,
	}.Apply(ctx)
}

// s82OldB432SunkenHopePick returns the chosen creature to its owner's hand. An
// on-disk key: never renamed, never reused.
var s82OldB432SunkenHopePick = RegisterPermanentsPick("permanents-pick/s82-old-b432-sunken-hope-bounce", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		if ids := picked.Cards(); len(ids) > 0 {
			ctx.Game.BounceCardsToHandForEffect(ids)
		}
		return nil
	},
})

// S82OldB432NullmageReturnAndDestroy is Nullmage Advocate's ability as it
// resolves: each of the two target cards still legal returns to its owner's
// hand, then the artifact or enchantment, if still legal, is destroyed
// (CR 608.2b: a clause whose target is gone is skipped, the rest happens).
func S82OldB432NullmageReturnAndDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.ClauseTargets(0) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
		return DestroyTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}
