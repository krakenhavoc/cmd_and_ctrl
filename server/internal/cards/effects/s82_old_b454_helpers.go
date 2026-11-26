package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b454_helpers.go — Commander 2011 counters (S82 older sets, batch
// 454, #5574): the shared shapes of Crescendo of War, Sigil Captain, Vish Kal
// and Szadek.

// S82OldB454CombatAnthem is "<Attacking | Blocking> creatures [you control] get
// +1/+0 for each <kind> counter on this <permanent>": a layer-7c static whose
// size is read from the source's counters on every layer pass. A creature's
// attacking or blocking status changes the set, so the layer cache is dropped
// on a declared attack or a creature leaving combat (DependsOnAttackingStatus)
// and, for blockers, on a declared block (EventBlock).
func S82OldB454CombatAnthem(label, kind string, attacking, yoursOnly bool) game.StaticAbility {
	s := game.StaticAbility{
		Label:    label,
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			if !target.IsCreature() || (yoursOnly && target.Controller != source.Controller) {
				return false
			}
			if attacking {
				return target.AttackingTarget != uuid.Nil
			}
			return target.BlockingTarget != uuid.Nil
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			c.Power += source.Counters[kind]
		},
	}
	// Leaving combat (CR 506.4) is announced as an attack change, so both
	// statics listen for it; a block is its own event.
	s.DependsOnAttackingStatus = true
	if !attacking {
		s.DependsOnEvents = []game.EventKind{game.EventBlock}
	}
	return s
}

// S82OldB454OneOneEntered is Sigil Captain's "whenever a creature you control
// enters, if that creature is 1/1": the entering creature is yours (the Captain
// itself counts) and is 1/1 as the event happens.
func S82OldB454OneOneEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := EnteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.PowerForComparison() == 1 && c.CurrentToughness() == 1
}

// S82OldB454SigilCaptain is the resolution: two +1/+1 counters on the creature
// that entered, if it is still 1/1 and is still the object that entered
// (CR 603.4, 400.7).
func S82OldB454SigilCaptain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left || info.Power != 1 || info.Toughness != 1 {
		return nil
	}
	return AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
}

// S82OldB454VishKalFeed is "Sacrifice a creature: Put X +1/+1 counters on Vish
// Kal, where X is the sacrificed creature's power."
func S82OldB454VishKalFeed(g *game.Game, item *game.StackItem) error {
	fed, ok := B17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	n := DepartedCreaturePower(g, fed)
	ctx := NewContext(g, item)
	if n <= 0 || fed == item.SourceCardID || !OnBattlefield(g, item.SourceCardID) || ctx.IsNewSourceObject(item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// S82OldB454VishKalShrink is "Remove all +1/+1 counters from Vish Kal: Target
// creature gets -1/-1 until end of turn for each +1/+1 counter removed this
// way."
func S82OldB454VishKalShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := ctx.CountersRemoved()
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || n <= 0 {
		return nil
	}
	return BoostUntilEOT{Target: t.ID, Power: -n, Toughness: -n, Label: "Vish Kal — -1/-1 for each counter removed"}.Apply(ctx)
}
