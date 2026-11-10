package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	kirolAttentiveFirstYearOracle = "3a272fa9-5f28-43a4-95de-10cb511f1b43"
	kirolElvishVisionaryOracle    = "c6a3a882-a127-4590-93d7-679ef4313efe"
)

func TestKirolAttentiveFirstYearIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(kirolAttentiveFirstYearOracle)
	if !ok || spec.Name != "Kirol, Attentive First-Year" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// kirolCastVisionary casts an Elvish Visionary and passes priority until
// its enters trigger is on the stack, returning that trigger's item id.
func kirolCastVisionary(t *testing.T, g *game.Game, me *game.Player) (creature, trigger uuid.UUID) {
	t.Helper()
	creature = rfdHandCard(me, "Elvish Visionary", "Creature — Elf Shaman", "", kirolElvishVisionaryOracle, 1, 1)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 16; i++ {
		if it := acItemOnStack(g, "Elvish Visionary — draw a card"); it != nil {
			return creature, it.ID
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the Visionary's enters trigger never reached the stack")
	return
}

// Tapping Kirol and another creature copies a triggered ability you
// control: the Visionary's draw happens twice. A second activation in
// the same turn is refused, and the next trigger draws once.
func TestKirolCopiesATriggeredAbilityOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	kirol := pushCatalogPermanent(g, me.ID, "Kirol, Attentive First-Year", "Legendary Creature — Vampire Cleric", kirolAttentiveFirstYearOracle, true)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "Other Bear", 2, 2)

	_, trig := kirolCastVisionary(t, g, me)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, kirol, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trig}},
		TapIDs:  []uuid.UUID{kirol, bear}}); err != nil {
		t.Fatalf("activate (a summoning-sick Kirol may tap for it): %v", err)
	}
	if !b16Tapped(t, g, kirol) || !b16Tapped(t, g, bear) || b16Tapped(t, g, other) {
		t.Error("the cost taps exactly the two named creatures")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Fatalf("hand %d after the trigger and its copy, want %d", got, hand+2)
	}

	visionary, trig2 := kirolCastVisionary(t, g, me)
	hand = me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, kirol, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trig2}},
		TapIDs:  []uuid.UUID{other, visionary}}); err == nil {
		t.Fatal("activated a second time this turn")
	}
	if b16Tapped(t, g, other) {
		t.Error("a refused activation tapped a creature")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after an uncopied trigger, want %d", got, hand+1)
	}
}

// The cost takes two untapped creatures you control (one, or an
// opponent's, doesn't pay it), and an opponent's trigger is not a legal
// target.
func TestKirolNeedsTwoOfYourCreaturesAndYourOwnTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMainForCost(t, g)
	kirol := pushCatalogPermanent(g, me.ID, "Kirol, Attentive First-Year", "Legendary Creature — Vampire Cleric", kirolAttentiveFirstYearOracle, false)
	theirBear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Artifact", "", false)

	visionary, trig := kirolCastVisionary(t, g, me)
	if err := g.AnnounceTrigger(opp.ID, theirs, game.AbilityParams{Label: "their trigger"}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	offered := acLegalAbilityTargets(g, me.ID, kirolAttentiveFirstYearOracle, 0)
	if offered[acItemLabelled(g, "their trigger")] {
		t.Error("an opponent's trigger is not 'you control'")
	}
	if !offered[trig] {
		t.Error("your own trigger is offered")
	}

	for _, pay := range [][]uuid.UUID{{kirol}, {kirol, theirBear}} {
		if err := g.ActivateCatalogAbility(me.ID, kirol, 0, game.ActivateAbilityParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: trig}},
			TapIDs:  pay}); err == nil {
			t.Errorf("paid the cost with %v", pay)
		}
	}
	if b16Tapped(t, g, kirol) || b16Tapped(t, g, visionary) || b16Tapped(t, g, theirBear) {
		t.Error("a refused activation tapped a creature")
	}
}
