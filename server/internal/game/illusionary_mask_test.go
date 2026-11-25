package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// illusionary_mask_test.go — #6219, Illusionary Mask: "{X}: You may
// choose a creature card in your hand whose mana cost could be paid by
// some amount of, or all of, the mana you spent on {X}. If you do, you may
// cast that card face down as a 2/2 creature spell without paying its
// mana cost. If the creature that spell becomes as it resolves has not
// been turned face up and would assign or deal damage, be dealt damage,
// or become tapped, instead it's turned face up and …" CR 708.2: the
// face-down object is a 2/2 with no text; CR 608.2g: the cast happens
// during the ability's resolution.

// pushMask puts an artifact with the Mask's ability on `me`'s battlefield.
func pushMask(g *Game, me *Player) uuid.UUID {
	c := NewCard("Illusionary Mask", me.ID)
	c.TypeLine = "Artifact"
	c.Controller = me.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "{X}: You may cast a creature card face down.",
		Cost:  AbilityCost{Mana: "{X}"},
		Effect: func(g *Game, item *StackItem) error {
			cands := g.MaskCandidatesForEffect(item.Controller, item.Paid.Mana)
			g.OfferMaskCastForEffect(item.Controller, item.SourceCardID, "Illusionary Mask", cands)
			return nil
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// bigInHand puts a {2}{G}{G} 5/5 creature card in `me`'s hand.
func bigInHand(me *Player) uuid.UUID {
	c := NewCard("Craw Giant", me.ID)
	c.TypeLine = "Creature — Giant"
	c.ManaCost = "{2}{G}{G}"
	c.Power, c.Toughness = 5, 5
	me.Hand.PushTop(c)
	return c.InstanceID
}

// maskedOnBattlefield activates the Mask for X=4 with {G}{G}{C}{C}, casts
// the creature face down through the offer and resolves it. It returns the
// permanent.
func maskedOnBattlefield(t *testing.T, g *Game, me *Player) uuid.UUID {
	t.Helper()
	mask := pushMask(g, me)
	big := bigInHand(me)
	me.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "G"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, mask, 0, ActivateAbilityParams{XValue: 4, Strict: true}); err != nil {
		t.Fatalf("activate the Mask for X=4: %v", err)
	}
	var offer *PendingChoice
	for i := 0; i < 8 && offer == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
		offer = pendingChoiceOfKind(g, PendingChoiceResolutionCast)
	}
	if offer == nil {
		t.Fatal("the Mask's resolution offered no face-down cast")
	}
	if !slices.Contains(offer.DiscardOptions, big) {
		t.Fatalf("the offer = %v, want the {2}{G}{G} creature", offer.DiscardOptions)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, big, CastSpellParams{}); err != nil {
		t.Fatalf("cast it face down: %v", err)
	}
	if c, ok := g.LookupCardForEffect(big); !ok || !c.FaceDown || c.FaceDownKind != FaceDownMasked {
		t.Fatalf("the spell is not a masked face-down spell: %+v", c.FaceDownKind)
	}
	settleStack(t, g)
	c := findBattlefieldCard(g, big)
	if c == nil {
		t.Fatal("the face-down spell did not resolve onto the battlefield")
	}
	if !c.FaceDown || c.FaceDownKind != FaceDownMasked {
		t.Fatalf("the permanent is not face down and masked: %q", c.FaceDownKind)
	}
	return big
}

// TestMaskCastsAFourDropFaceDownAndItTurnsUpWhenDealtDamage — the
// acceptance test: a 4-drop cast with X=4 enters as a 2/2 face down and
// turns face up as it would be dealt damage, then is dealt it as the 5/5.
func TestMaskCastsAFourDropFaceDownAndItTurnsUpWhenDealtDamage(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := maskedOnBattlefield(t, g, me)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCard(g, big)
	if p, tg := c.CurrentPower(), c.CurrentToughness(); p != 2 || tg != 2 {
		t.Fatalf("face down it is %d/%d, want the CR 708.2 2/2", p, tg)
	}

	var err error
	g.WithWriteLock(func() { err = g.dealDamageToPermanentLocked(uuid.Nil, nil, big, 3, DamageMarks{}) })
	if err != nil {
		t.Fatalf("deal 3: %v", err)
	}
	g.RunStateChecksForTest()
	c = findBattlefieldCard(g, big)
	if c == nil {
		t.Fatal("it died: it was dealt the damage as the 2/2, not turned face up first")
	}
	if c.FaceDown {
		t.Fatal("it is still face down after it would have been dealt damage")
	}
	if c.DamageMarked != 3 {
		t.Errorf("damage = %d, want the 3 it was dealt as the face-up creature", c.DamageMarked)
	}
}

// TestMaskCandidatesAreWhatTheSpentManaCouldPay — "whose mana cost could
// be paid by some amount of, or all of, the mana you spent on {X}".
func TestMaskCandidatesAreWhatTheSpentManaCouldPay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	big := bigInHand(me)
	var got []uuid.UUID
	g.WithWriteLock(func() {
		got = g.MaskCandidatesForEffect(me.ID, []ManaToken{{Color: "C"}, {Color: "C"}, {Color: "C"}, {Color: "C"}})
	})
	if slices.Contains(got, big) {
		t.Error("four colourless mana pays {2}{G}{G}")
	}
	g.WithWriteLock(func() {
		got = g.MaskCandidatesForEffect(me.ID, []ManaToken{{Color: "G"}, {Color: "G"}, {Color: "R"}, {Color: "C"}, {Color: "C"}})
	})
	if !slices.Contains(got, big) {
		t.Error("{G}{G}{R}{C}{C} does not pay {2}{G}{G} with some of it")
	}
}

// TestMaskedPermanentTurnsUpAsItBecomesTapped — "or become tapped,
// instead it's turned face up and becomes tapped".
func TestMaskedPermanentTurnsUpAsItBecomesTapped(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := maskedOnBattlefield(t, g, me)
	var err error
	g.WithWriteLock(func() { err = g.TapTargetForEffect(big) })
	if err != nil {
		t.Fatal(err)
	}
	c := findBattlefieldCard(g, big)
	if c.FaceDown || !c.Tapped {
		t.Fatalf("face down = %v, tapped = %v; want it face up and tapped", c.FaceDown, c.Tapped)
	}
}

// TestMaskedBlockerTurnsUpBeforeItAssignsCombatDamage — "would assign …
// combat damage": the masked blocker is turned face up first and assigns
// its face-up power.
func TestMaskedBlockerTurnsUpBeforeItAssignsCombatDamage(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := maskedOnBattlefield(t, g, me)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, big)
		c.BlockingTarget = uuid.New() // a stand-in attacker that is gone: nothing to deal damage to
	})
	g.WithWriteLock(func() { g.unmaskCombatantsLocked(false) })
	if c := findBattlefieldCard(g, big); c.FaceDown {
		t.Fatal("a masked blocker that would assign combat damage stayed face down")
	}
}

// TestMaskedPermanentStaysDownOtherwise — nothing else turns it up, and
// it offers no way up of its own (CR 708.7): its card has no morph.
func TestMaskedPermanentStaysDownOtherwise(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := maskedOnBattlefield(t, g, me)
	c := findBattlefieldCard(g, big)
	if offers := TurnFaceUpOffers(*c); len(offers) != 0 {
		t.Errorf("a masked permanent with no morph offers %v to turn it up", offers)
	}
	if !c.FaceDown {
		t.Error("it was turned up with nothing happening to it")
	}
}

// TestMaskedAttackerTurnsUpAsTheDeclarationTapsIt — declaring it as an
// attacker taps it (CR 508.1f), so it is turned face up as the
// declaration is locked in, and attacks as the face-up creature.
func TestMaskedAttackerTurnsUpAsTheDeclarationTapsIt(t *testing.T) {
	g := newActiveGame(t)
	toMainPhase(t, g)
	mySeat := g.Turn.ActiveSeat
	me, them := g.Seats[mySeat], g.Seats[1-mySeat]
	big := maskedOnBattlefield(t, g, me)
	g.WithWriteLock(func() { findBattlefieldCard(g, big).SummonedThisTurn = false })
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(big, them.ID); err != nil {
		t.Fatalf("declare: %v", err)
	}
	g.WithWriteLock(func() { g.commitAttackDeclarationLocked() })
	c := findBattlefieldCard(g, big)
	if c.FaceDown || !c.Tapped || c.AttackingTarget != them.ID {
		t.Fatalf("face down = %v, tapped = %v, attacking = %v; want it face up, tapped and attacking", c.FaceDown, c.Tapped, c.AttackingTarget)
	}
}
