package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b824_helpers.go — Homelands, "Damage" (S82 OLD batch 824, #5635). A
// file of its own so concurrent batches never collide.

// OldB824DamageEachCreatureForEachAura is "deals 2 damage to each creature for
// each Aura attached to that creature" (Baki's Curse): one damage instance
// (CR 120.1), each creature taking twice the Auras attached to it as the spell
// resolves. A creature with none takes none.
func OldB824DamageEachCreatureForEachAura(perAura int) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return ctx.Game.DamageInstanceForEffect(func() error {
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if !c.IsCreature() {
					continue
				}
				n := perAura * B18AurasAttachedTo(ctx.Game, c.InstanceID)
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
}

// OldB824ControllerHasAnIsland judges "whose controller controls an Island"
// for a target clause (Dwarven Sea Clan), through the layers.
func OldB824ControllerHasAnIsland() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, o := range g.BattlefieldCardsForEffect() {
			if o.Controller == c.Controller && o.IsLand() && o.HasSubtypeWord("Island") {
				return true
			}
		}
		return false
	}
}

// OldB824BeforeTheEndOfCombatStep is "Activate only before the end of combat
// step": the combat phase of any turn up to, and not including, its end of
// combat step (CR 506.1). With the attacking-or-blocking clause on the target,
// that is declare attackers through combat damage.
func OldB824BeforeTheEndOfCombatStep() ActivationCondition {
	return func(g *game.Game, _, _ uuid.UUID) bool {
		return game.PhaseOf(g.Turn.Step) == game.PhaseCombat && g.Turn.Step != game.StepEndCombat
	}
}

// oldB824SourceDealsTwoToTheCreatureBody is the delayed trigger of Dwarven Sea
// Clan: "This creature deals 2 damage to that creature at end of combat". The
// Clan is the params' object, read as it last existed if it has left (CR
// 608.2h); the listed card is the creature, which is dealt nothing if it is no
// longer on the battlefield (CR 400.7). A registered key (ADR 0041 phase 3).
var oldB824SourceDealsTwoToTheCreatureBody = game.DelayedBody("s82-old-b824/source-deals-two-to-the-creature",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if len(item.Targets) == 0 {
			return nil
		}
		victim := item.Targets[0].ID
		if z := g.FindCardZoneForEffect(victim); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		return DealDamage{Target: victim, Amount: 2, SourceObject: &p.Object}.Apply(NewContext(g, item))
	})

// OldB824DealsTwoAtEndOfCombat is "Choose target attacking or blocking creature
// whose controller controls an Island. This creature deals 2 damage to that
// creature at end of combat": the clause is re-judged as the ability resolves,
// and the damage is a delayed trigger at the end of combat step of this combat.
func OldB824DealsTwoAtEndOfCombat(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		ref := game.ObjectRef{ID: item.SourceCardID, Epoch: g.ObjectEpochForEffect(item.SourceCardID)}
		return ScheduleDelayedTrigger{
			At:     game.StepEndCombat,
			Label:  name + " — it deals 2 damage to that creature",
			Cards:  []uuid.UUID{id},
			Body:   oldB824SourceDealsTwoToTheCreatureBody,
			Params: game.EffectParams{Object: ref},
		}.Apply(ctx)
	}
}
