package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mana_pipeline_test.go — the four new seams on ManaAbility (#352),
// exercised through ActivateManaAbility and the auto-tapper:
//
//  1. a mana component in the cost (the Signet cycle);
//  2. an activation gate (Temple of the False God, Mox Opal);
//  3. a produced string computed at activation (derived and scaled);
//  4. spend restrictions stamped onto the produced tokens, including
//     across a PendingChoiceMana.

// --- 1. a mana component in the cost -----------------------------

func signetShape(a, b string) []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost:  true,
		ManaCost: "{1}",
		Produced: "{" + a + "}{" + b + "}",
		Label:    "{1}, {T}: Add {" + a + "}{" + b + "}",
	}}
}

func TestManaAbilityManaCostIsPaidFromThePool(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	signet := pushIntrinsicPermanent(g, me, "Azorius Signet", "Artifact", signetShape("W", "U"), nil)
	me.ManaPool.AddMana(ManaToken{Color: "G"})

	if err := g.ActivateManaAbility(me.ID, signet, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	// {G} in, {W}{U} out — net +1 and a colour fix.
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want two tokens", me.ManaPool)
	}
	if me.ManaPool[0].Color != "W" || me.ManaPool[1].Color != "U" {
		t.Errorf("pool = %v, want {W}{U}", me.ManaPool)
	}
	if !g.Battlefield.Cards[cardIndex(t, g, signet)].Tapped {
		t.Error("the Signet did not tap")
	}
}

// The whole point of a COST: it is validated before anything is paid,
// so an unaffordable activation leaves the permanent untouched.
func TestManaAbilityManaCostRejectsAnEmptyPoolWithoutTapping(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	signet := pushIntrinsicPermanent(g, me, "Dimir Signet", "Artifact", signetShape("U", "B"), nil)

	err := g.ActivateManaAbility(me.ID, signet, 0, ManaAbilityParams{})
	if err == nil {
		t.Fatal("a Signet activated on an empty pool succeeded")
	}
	var insufficient *InsufficientManaError
	if !errors.As(err, &insufficient) {
		t.Errorf("err = %v, want *InsufficientManaError", err)
	}
	if g.Battlefield.Cards[cardIndex(t, g, signet)].Tapped {
		t.Error("a rejected activation tapped the source anyway")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// Restricted mana cannot fund an activation it does not name — the
// spend context on the activation path, not just the cast path.
func TestManaAbilityManaCostHonoursSpendRestrictions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	signet := pushIntrinsicPermanent(g, me, "Boros Signet", "Artifact", signetShape("R", "W"), nil)
	me.ManaPool.AddMana(ManaToken{
		Color:        "C",
		Restrictions: []string{ManaRestrictColorless, ManaRestrictSubtype("Eldrazi")},
	})

	if err := g.ActivateManaAbility(me.ID, signet, 0, ManaAbilityParams{}); err == nil {
		t.Fatal("Eldrazi-only mana funded a Signet activation")
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the restricted token still there", me.ManaPool)
	}
}

// --- 2. the activation gate --------------------------------------

func gatedShape(n int) []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost:  true,
		Produced: "{C}{C}",
		Label:    "Add {C}{C}",
		Condition: func(g *Game, controller, _ uuid.UUID) bool {
			count := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == controller && c.IsLand() {
					count++
				}
			}
			return count >= n
		},
	}}
}

func TestGatedManaAbilityRefusesBelowThresholdAndPaysNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	temple := pushIntrinsicPermanent(g, me, "Temple of the False God", "Land", gatedShape(5), nil)

	err := g.ActivateManaAbility(me.ID, temple, 0, ManaAbilityParams{})
	if !errors.Is(err, ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet", err)
	}
	if g.Battlefield.Cards[cardIndex(t, g, temple)].Tapped {
		t.Error("a gated ability that failed its gate still tapped the source")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

func TestGatedManaAbilityFiresAtThreshold(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	temple := pushIntrinsicPermanent(g, me, "Temple of the False God", "Land", gatedShape(5), nil)
	// The Temple counts itself, so four more lands reaches five.
	for i := 0; i < 4; i++ {
		pushIntrinsicPermanent(g, me, "Wastes", "Basic Land — Wastes", nil, nil)
	}

	if err := g.ActivateManaAbility(me.ID, temple, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want {C}{C}", me.ManaPool)
	}
}

// A gated source below its threshold is not a source. Planning it
// would produce a plan the executor then refuses, stranding whatever
// it had already tapped.
func TestAutoTapperSkipsAGatedSourceBelowThreshold(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Temple of the False God", "Land", gatedShape(5), nil)
	cost, _ := ParseCost("{1}")

	if _, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Error("the auto-tapper planned a Temple of the False God on one land")
	}
}

// --- 3. a produced string computed at activation -----------------

func scaledShape(symbol string) []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost: true,
		Label:   "Add " + symbol + " for each creature you control",
		ProducedFunc: func(g *Game, controller, _ uuid.UUID) string {
			out := ""
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == controller && c.IsCreature() {
					out += symbol
				}
			}
			return out
		},
	}}
}

func TestScaledManaAbilityCountsTheBoard(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cradle := pushIntrinsicPermanent(g, me, "Gaea's Cradle", "Legendary Land", scaledShape("{G}"), nil)
	for i := 0; i < 3; i++ {
		pushIntrinsicPermanent(g, me, "Bear", "Creature — Bear", nil, nil)
	}
	// An opponent's creatures do not count.
	pushIntrinsicPermanent(g, g.Seats[1], "Bear", "Creature — Bear", nil, nil)

	if err := g.ActivateManaAbility(me.ID, cradle, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool = %v, want three {G}", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "G" {
			t.Errorf("token %v, want {G}", tok)
		}
	}
}

// Nothing to count is not an error: the land taps and adds nothing,
// which is what the printed card does.
func TestScaledManaAbilityWithNothingToCountStillTaps(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	cradle := pushIntrinsicPermanent(g, me, "Gaea's Cradle", "Legendary Land", scaledShape("{G}"), nil)

	if err := g.ActivateManaAbility(me.ID, cradle, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
	if !g.Battlefield.Cards[cardIndex(t, g, cradle)].Tapped {
		t.Error("the land did not tap")
	}
}

// --- 4. spend restrictions on produced tokens --------------------

func restrictedShape(produced string, restrictions []string) []ManaAbilityShape {
	return []ManaAbilityShape{{
		TapCost:      true,
		Produced:     produced,
		Label:        "restricted",
		Restrictions: restrictions,
	}}
}

func TestRestrictionsAreStampedOntoProducedTokens(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	tags := []string{ManaRestrictColorless, ManaRestrictSubtype("Eldrazi")}
	temple := pushIntrinsicPermanent(g, me, "Eldrazi Temple", "Land", restrictedShape("{C}{C}", tags), nil)

	if err := g.ActivateManaAbility(me.ID, temple, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 2 {
		t.Fatalf("pool = %v, want two tokens", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if len(tok.Restrictions) != 2 {
			t.Fatalf("token %v carries %d restrictions, want 2", tok, len(tok.Restrictions))
		}
	}
	// And the backing array is not shared with the catalog shape —
	// mutating one token must not reach the other.
	me.ManaPool[0].Restrictions[0] = "mutated"
	if me.ManaPool[1].Restrictions[0] == "mutated" {
		t.Error("two tokens share one restriction backing array")
	}
}

// The hardest leak to spot: a pipe slot mints its token LATER, in
// ResolveManaChoice, long after the ability shape is out of scope.
// Restrictions have to ride the PendingChoice to get there.
func TestRestrictionsSurviveAPendingManaChoice(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	tags := []string{ManaRestrictCast, ManaRestrictSupertype("Legendary")}
	halfling := pushIntrinsicPermanent(g, me, "Delighted Halfling", "Creature — Halfling Citizen",
		restrictedShape("{W|U|B|R|G}", tags), []string{"haste"})

	if err := g.ActivateManaAbility(me.ID, halfling, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	var choice *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMana {
			choice = c
		}
	}
	if choice == nil {
		t.Fatal("no mana pick was queued")
	}
	if len(choice.ManaRestrictions) != 2 {
		t.Fatalf("choice carries %v, want the ability's two restrictions", choice.ManaRestrictions)
	}
	if err := g.ResolveManaChoice(choice.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("pool = %v, want one token", me.ManaPool)
	}
	if len(me.ManaPool[0].Restrictions) != 2 {
		t.Errorf("the minted token carries %v — a pipe slot leaked unrestricted mana",
			me.ManaPool[0].Restrictions)
	}
}

// New game state must survive an undo snapshot intact and unaliased.
func TestCloneDeepCopiesPendingManaRestrictions(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceMana,
		Chooser:          me.ID,
		FromPlayer:       me.ID,
		Count:            1,
		ColorOptions:     []string{"W", "U"},
		ManaRestrictions: []string{ManaRestrictCast},
	})
	clone := g.Clone()
	if len(clone.PendingChoices) != 1 {
		t.Fatalf("clone has %d pending choices, want 1", len(clone.PendingChoices))
	}
	if len(clone.PendingChoices[0].ManaRestrictions) != 1 {
		t.Fatal("the clone lost the choice's mana restrictions")
	}
	clone.PendingChoices[0].ManaRestrictions[0] = "mutated"
	if g.PendingChoices[0].ManaRestrictions[0] == "mutated" {
		t.Error("clone and original share one ManaRestrictions backing array — undo would corrupt the live game")
	}
}

// The auto-tapper refuses both kinds of source it cannot reason
// about: one with a mana cost (recursive) and one with restricted
// output (a decision, not a resource).
func TestAutoTapperSkipsManaCostAndRestrictedSources(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Azorius Signet", "Artifact", signetShape("W", "U"), nil)
	pushIntrinsicPermanent(g, me, "Ancient Ziggurat", "Land",
		restrictedShape("{G}", []string{ManaRestrictCast, ManaRestrictType("Creature")}), nil)
	cost, _ := ParseCost("{1}")

	if plan, ok := g.AutoTapForCost(me.ID, cost, 0); ok {
		t.Errorf("the auto-tapper planned %v; both sources should be invisible to it", plan)
	}
}

// cardIndex finds a battlefield card by instance ID.
func cardIndex(t *testing.T, g *Game, id uuid.UUID) int {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return i
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return -1
}
