package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b05_helpers.go — shared bodies for The Brothers' War batch 5
// (#5482): the Soldier cards.

// broB05SoldierToken is the printed 1/1 colorless Soldier artifact
// creature token's tokens_table.go key.
const BroB05SoldierToken = "1/1 colorless Soldier artifact"

// broB05NontokenSoldierYouControlDied is "another nontoken Soldier you
// control dies": the dead permanent, read as it last existed on the
// battlefield (CR 603.10a), is another Soldier, not a token, and left
// under the source's controller's control.
func BroB05NontokenSoldierYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	return ok && dead.InstanceID != source.InstanceID && !IsToken(dead) &&
		LeftUnderControlOf(ev, dead) == source.Controller &&
		LeftAsSubtype(ev, dead, source.TextSubtype("Soldier"))
}

// broB05EntersTappedUnlessRevealOrControl is "As this land enters, you may
// reveal a <subtype> card from your hand. This land enters tapped unless
// you revealed a <subtype> card this way or you control a <subtype>": the
// reveal-land replacement that does not apply at all, and so asks nothing,
// while the entering permanent's controller controls one.
func BroB05EntersTappedUnlessRevealOrControl(name, subtype string) game.ReplacementEffect {
	r := EntersTappedUnlessYouRevealFromHand(name, "a "+subtype+" card", func(c game.Card) bool {
		return c.HasSubtypeWord(subtype)
	})
	r.RowLabel = "Enters tapped unless you reveal a " + subtype + " card from your hand or you control a " + subtype + "."
	base := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !base(ev, g, src) {
			return false
		}
		controller := EnteringPermanentChooser(ev, g, src)
		word := subtype
		if src != nil {
			word = src.TextSubtype(subtype)
		}
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.IsCreature() && c.HasSubtype(word) {
				return false
			}
		}
		return true
	}
	return r
}

// broB05PreventDamageToOtherAttackingSoldiers is "Prevent all damage that
// would be dealt to other attacking Soldiers you control" (Rescue
// Retriever): a CR 615 prevention over every damage event, combat or not,
// aimed at an attacking Soldier the source's controller controls that is
// not the source.
func BroB05PreventDamageToOtherAttackingSoldiers(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == src.InstanceID {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && c.Controller == src.Controller && c.AttackingTarget != uuid.Nil &&
				c.IsCreature() && game.ThroughTextOf(src, c).HasSubtypeWord("Soldier")
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// broB05ZephyrSentinelEnters is "return up to one other target creature
// you control to its owner's hand. If it was a Soldier, put a +1/+1
// counter on this creature." The Soldier test reads the creature as it is
// on the battlefield, before it leaves, and rides the bounce's keyed tail
// as N; the tail runs once the card has reached a hand (a move the window
// cancelled or redirected puts no counter), and the counter lands on the
// Sentinel only while it is still there.
func BroB05ZephyrSentinelEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok || !OnBattlefield(g, id) {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	wasSoldier := 0
	if c, found := g.LookupCardForEffect(id); found {
		if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && game.ThroughTextOf(&src, c).HasSubtypeWord("Soldier") {
			wasSoldier = 1
		}
	}
	return BounceToHand{Target: id, ThenTail: WithN(ctx.ItemTail(broB05ZephyrSentinelCounter), wasSoldier)}.Apply(ctx)
}

// broB05ZephyrSentinelCounter is the "if it was a Soldier, put a +1/+1
// counter on this creature" half of Zephyr Sentinel's trigger. An on-disk
// key.
var broB05ZephyrSentinelCounter = CardTail("bro-b05/zephyr-sentinel-counter", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N == 0 || len(o.IDs) == 0 || !OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
