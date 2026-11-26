package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB378PeregrinationHandTail is the hand half of Peregrination's two-land
// search: one more basic land into your hand, then the shuffle (a registered
// tail, so the first land's entry question is a restore point). An on-disk key.
var S82OldB378PeregrinationHandTail = CultivateHandTail("peregrination/basic-land-to-hand", "Peregrination")

// S82OldB378DevotionOf reads "your devotion to <colour>" (CR 700.5) for the
// spell or ability being resolved, the colour read through the card's text
// (CR 612.2).
func S82OldB378DevotionOf(ctx *Context, colorLetter string) int {
	return DevotionTo(ctx.Game, ctx.Controller(), ctx.TextColor(colorLetter))
}

// S82OldB378CantBlockThisCombat is "That creature can't block this combat"
// for one creature: a CR 509.1b restriction that ends with combat (Flummoxed
// Cyclops' shape). A creature that has left the battlefield is left alone.
func S82OldB378CantBlockThisCombat(ctx *Context, creature uuid.UUID, label string) error {
	d, ok := DurationUntilEndOfCombat(ctx)
	if !ok || !OnBattlefield(ctx.Game, creature) {
		return nil
	}
	return ScopedEffectFor{
		Target:   creature,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBlock)},
		Duration: d,
		Label:    label,
	}.Apply(ctx)
}

// S82OldB378ForgestokerPing is Forgestoker Dragon's ability: one damage to
// the target creature, which then can't block this combat. The restriction
// does not depend on the damage being dealt, so a prevented ping still stops
// the block; a target that is no longer legal is left alone (CR 608.2b).
func S82OldB378ForgestokerPing(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
		}
	}
	if target == uuid.Nil {
		return nil
	}
	if err := SourceDealsDamageToEachLegalTarget(1)(g, item); err != nil {
		return err
	}
	return S82OldB378CantBlockThisCombat(ctx, target, "Forgestoker Dragon — can't block this combat")
}

// S82OldB378SpellDamagedAnOpponent is the condition of "whenever an instant
// or sorcery spell you control deals damage to an opponent" (Satyr
// Firedancer): the damage source is a spell its controller controls, read
// while it is mid-resolution, and the damage went to another player.
func S82OldB378SpellDamagedAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target == uuid.Nil || source == nil {
		return false
	}
	spell, item, ok := g.ResolvingSpellForEffect(ev.Source)
	if !ok || item.Kind != game.StackItemSpell || item.Controller != source.Controller {
		return false
	}
	if !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	return ev.Target != source.Controller && g.PlayerByIDForEffect(ev.Target) != nil
}

// S82OldB378ThatMuchToTarget is "this creature deals that much damage to
// target creature": the trigger's total (ThatMuchDamage) from the source
// permanent to the one target, if it is still legal.
func S82OldB378ThatMuchToTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if item.Trigger == nil || len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: item.Trigger.Event.Amount}.Apply(ctx)
}

// S82OldB378CombatPartners lists the creatures blocking or blocked by the
// source creature, read from its battlefield object (Vortex Elemental).
func S82OldB378CombatPartners(ctx *Context) []uuid.UUID {
	self, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	me := ctx.Source()
	var out []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == me || !c.IsCreature() {
			continue
		}
		blocking := self.Attacking && c.IsBlockingAttacker(me)
		blocked := c.AttackingTarget != uuid.Nil && slices.Contains(self.Blocking, c.InstanceID)
		if blocking || blocked {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// s82OldB378VortexShuffle is the tail of Vortex Elemental's first ability:
// "then those players shuffle". Carry names the owners, read before the move,
// so a commander that went to the command zone instead (CR 903.9) still has
// its owner shuffle.
var s82OldB378VortexShuffle = game.RegisterTail("then/s82-old-b378-vortex-shuffle", func(g *game.Game, o game.TailOutcome) error {
	for _, owner := range o.Tail.Carry {
		if err := g.ShuffleLibraryForEffect(owner); err != nil {
			return err
		}
	}
	return nil
})

// S82OldB378VortexTuck puts Vortex Elemental and each creature blocking or
// blocked by it on top of their owners' libraries, then those players shuffle.
func S82OldB378VortexTuck(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ids := append([]uuid.UUID{ctx.Source()}, S82OldB378CombatPartners(ctx)...)
	var owners []uuid.UUID
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if ok && !slices.Contains(owners, c.Owner) {
			owners = append(owners, c.Owner)
		}
	}
	return g.TuckCardsToLibraryKeyedForEffect(ids, game.TuckOptions{},
		game.KeyedTail{Ref: s82OldB378VortexShuffle, Player: item.Controller, Source: item.SourceCardID, Carry: owners})
}

// S82OldB378TargetBlocksThis is "Target creature blocks this creature this
// turn if able" (Vortex Elemental): the requirement names this creature as
// the object it is now (CR 509.1c).
func S82OldB378TargetBlocksThis(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(ctx.Source())
	if !ok {
		return nil
	}
	return BlocksAttackerUntilEOT{Blocker: t.ID, Attacker: ref, Label: "Vortex Elemental — blocks it this turn if able"}.Apply(ctx)
}
