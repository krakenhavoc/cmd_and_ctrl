package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_hand_source_test.go — #1228. The auto-tapper can now plan a
// source that is NOT A PERMANENT: a Spirit Guide in the controller's
// hand whose mana ability functions there (CR 113.6) and pays for
// itself by exiling the card.
//
// Two claims, and the second is the one that needed arguing. The
// first is that it works at all — before this the tapper's candidate
// set was the battlefield and a hand full of Spirit Guides read as
// "missing {R}" to the cast preview, the strict gate and every bot.
// The second is the ORDER: a card in hand is worth more than a
// Treasure, which is already the last-resort tier, so the hand tier
// sits below even that. Every ordering test below is really one
// sentence — the planner spends the cheapest thing on the board, and
// a card still in your hand is not on the board.

// planContains reports whether an auto-tap plan names this card.
func planContains(plan []uuid.UUID, want uuid.UUID) bool {
	for _, id := range plan {
		if id == want {
			return true
		}
	}
	return false
}

// The headline: a Spirit Guide in hand and nothing else pays {R}.
func TestAutoTapPlansASpiritGuideOutOfHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0)
	if !ok {
		t.Fatalf("a Spirit Guide in hand should pay {R}")
	}
	if len(plan) != 1 || plan[0] != guide {
		t.Fatalf("plan = %v, want [%v]", plan, guide)
	}
}

// …and the executor really spends it: the card leaves the hand for
// exile and the mana lands in the pool.
func TestAutoTapExecutorExilesTheSpiritGuide(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	plan, ok := g.autoTapLocked(me.ID, costFor(t, "{R}"), 0, nil)
	if !ok {
		t.Fatalf("no plan")
	}
	g.materializePlanLocked(me, plan, costFor(t, "{R}"))

	if me.Hand.Contains(guide) {
		t.Error("the planned Spirit Guide is still in hand")
	}
	exiledCard(t, g, guide)
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "R" {
		t.Fatalf("mana pool = %v, want one {R}", me.ManaPool)
	}
	// #1212 / #1216: the snapshot survives the exile, on the auto-tap
	// route exactly as on the hand-clicked one.
	if !me.ManaPool[0].SourceKinds.Has(ManaSourceCreature) {
		t.Error("the auto-tapped Spirit Guide's mana records no creature source")
	}
}

// The ordering claim, spelled out. A Mountain pays {R} and untaps next
// turn; the Spirit Guide is gone for good. The planner takes the land.
func TestAutoTapPrefersALandOverASpiritGuide(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0)
	if !ok {
		t.Fatalf("no plan for {R}")
	}
	if !planContains(plan, mountain) {
		t.Errorf("plan = %v, want the Mountain", plan)
	}
	if planContains(plan, guide) {
		t.Errorf("plan = %v spent a card out of hand while a land was untapped", plan)
	}
}

// And below the sacrifice tier: a Treasure is already the last thing
// the planner reaches for, and a card in hand is worth more than one.
func TestAutoTapPrefersATreasureOverASpiritGuide(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	treasure := pushTreasures(g, me, 1)[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0)
	if !ok {
		t.Fatalf("no plan for {R}")
	}
	if !planContains(plan, treasure) {
		t.Errorf("plan = %v, want the Treasure cracked before the hand card", plan)
	}
	if planContains(plan, guide) {
		t.Errorf("plan = %v spent a card out of hand while a Treasure was on the table", plan)
	}
}

// The same order on the GENERIC half, which is where the wrong answer
// is cheapest to reach: orderUnusedByGenericPreference has its own
// comparator and the tier has to be in both.
func TestAutoTapPrefersABoardOverASpiritGuideForGeneric(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	treasure := pushTreasures(g, me, 1)[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{2}"), 0)
	if !ok {
		t.Fatalf("no plan for {2}")
	}
	if !planContains(plan, mountain) || !planContains(plan, treasure) {
		t.Errorf("plan = %v, want the Mountain and the Treasure", plan)
	}
	if planContains(plan, guide) {
		t.Errorf("plan = %v reached into hand with two sources still untapped", plan)
	}
}

// …and when the board genuinely cannot pay, the hand source is taken.
// Weaker-than-printed is not the posture here: the card really is a
// mana source, so a cast it can fund must not read as unpayable.
func TestAutoTapReachesForTheSpiritGuideWhenNothingElseCanPay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	mountain := pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}{R}"), 0)
	if !ok {
		t.Fatalf("a Mountain plus a Spirit Guide should pay {R}{R}")
	}
	if !planContains(plan, mountain) || !planContains(plan, guide) {
		t.Errorf("plan = %v, want both sources", plan)
	}
}

// An OPPONENT's Spirit Guide is not a source of mine (CR 108.4), and
// neither is one in a hand I do not own.
func TestAutoTapNeverPlansAnotherSeatsHand(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	seedSpiritGuide(them, "Simian Spirit Guide", "{R}")

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0); ok {
		t.Fatalf("planned %v out of another seat's hand", plan)
	}
}

// The lock-tap exclusion reaches the hand pile too: a card the player
// has reserved is not a candidate, exactly as a reserved land is not.
func TestAutoTapExcludesAReservedSpiritGuide(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	guide := seedSpiritGuide(me, "Simian Spirit Guide", "{R}")

	if plan, ok := g.AutoTapForCostExcluding(me.ID, costFor(t, "{R}"), 0,
		map[uuid.UUID]bool{guide: true}); ok {
		t.Fatalf("planned %v, but the Spirit Guide was excluded", plan)
	}
}

// The picker's demands, from the other side. A mana ability that
// functions from a hand but does NOT pay by exiling itself would be a
// free repeatable mana source; effects.Register refuses one at boot,
// and the planner refuses it here too rather than trusting a check two
// packages away.
func TestAutoTapRefusesACostlessHandManaAbility(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	c := NewCard("Free Mana", me.ID)
	c.TypeLine = "Creature — Ape Spirit"
	c.ManaAbilities = []ManaAbilityShape{{
		Zones:    []ZoneKind{ZoneHand},
		Produced: "{R}",
		Label:    "Add {R}",
	}}
	me.Hand.PushTop(c)

	if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0); ok {
		t.Fatalf("planned %v — a hand mana ability with no cost is not a source", plan)
	}
}

// And the exclusions the battlefield picker makes for a decision the
// planner may not take are made here too — a life cost, restricted
// output, a discard. The hand-clicked row still offers all of them.
func TestAutoTapRefusesAHandSourceThatAsksAQuestion(t *testing.T) {
	base := func() ManaAbilityShape {
		return spiritGuideAbility("{R}")[0]
	}
	cases := map[string]func(*ManaAbilityShape){
		"a life cost":         func(a *ManaAbilityShape) { a.LifeCost = 1 },
		"restricted output":   func(a *ManaAbilityShape) { a.Restrictions = []string{ManaRestrictCast} },
		"a discard cost":      func(a *ManaAbilityShape) { a.DiscardCards = &DiscardCost{N: 1, Label: "a card"} },
		"an unpaid mana cost": func(a *ManaAbilityShape) { a.ManaCost = "{1}" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			ab := base()
			mutate(&ab)
			c := NewCard("Awkward Guide", me.ID)
			c.TypeLine = "Creature — Ape Spirit"
			c.ManaAbilities = []ManaAbilityShape{ab}
			me.Hand.PushTop(c)

			if plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{R}"), 0); ok {
				t.Fatalf("planned %v despite %s", plan, name)
			}
		})
	}
}
