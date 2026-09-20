package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const uginEyeOfTheStormsOracle = "5c58353a-fd60-4528-bf0d-669626cda0b2"

// pushColoredCreature seats a creature with an explicit colour so
// "one or more colors" has something real to read.
func pushUginVictim(g *game.Game, owner uuid.UUID, name string, colors []string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Colors:     colors,
		Owner:      owner,
		Controller: owner,
	})
}

// castUginEyeOfTheStorms puts Ugin in the active seat's hand with his
// PRINTED loyalty stamped on the card — deck import is what supplies
// it in a real game (ADR 0032 §1), and a fixture that skips it seats
// a walker with zero loyalty that the CR 704.5i state-based action
// bins on arrival.
func castUginEyeOfTheStorms(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID:      id,
		Name:            "Ugin, Eye of the Storms",
		TypeLine:        "Legendary Planeswalker — Ugin",
		OracleID:        uginEyeOfTheStormsOracle,
		StartingLoyalty: 7,
		Owner:           active.ID,
		Controller:      active.ID,
	})
	toMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Ugin: %v", err)
	}
	return id
}

// TestUginEyeOfTheStormsCastTriggerExilesAColoredPermanent: the cast
// trigger fires as the spell is announced and resolves above Ugin.
func TestUginEyeOfTheStormsCastTriggerExilesAColoredPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
	ugin := castUginEyeOfTheStorms(t, g)

	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	if triggerOnStack(g, ugin) == nil {
		t.Fatal("the cast trigger sits above Ugin on the stack")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("nothing is exiled until the trigger resolves")
	}
	passPriorityAroundTable(t, g)

	if !inExile(g, victim) {
		t.Error("the targeted coloured permanent is exiled")
	}
	if !g.Battlefield.Contains(ugin) {
		t.Error("Ugin resolves after his cast trigger")
	}
}

// TestUginEyeOfTheStormsCastTriggerSkipsColorlessPermanents: only a
// permanent that's one or more colors is a legal target.
func TestUginEyeOfTheStormsCastTriggerSkipsColorlessPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colorless := pushUginVictim(g, opp.ID, "A Colorless Golem", nil)
	colored := pushUginVictim(g, opp.ID, "A Blue Drake", []string{"U"})
	castUginEyeOfTheStorms(t, g)

	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, colorless) {
		t.Error("a colorless permanent is not a legal target")
	}
	if !hasID(p.PickTargetCards, colored) {
		t.Error("a coloured permanent is a legal target")
	}
}

// TestUginEyeOfTheStormsTriggersOnAColorlessSpell is the second
// trigger, and its complement: a coloured spell does nothing.
func TestUginEyeOfTheStormsTriggersOnAColorlessSpell(t *testing.T) {
	for _, tc := range []struct {
		name    string
		colors  []string
		trigger bool
	}{
		{"a colorless spell", nil, true},
		{"a red spell", []string{"R"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ugin := pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
			victim := pushUginVictim(g, opp.ID, "A Green Bear", []string{"G"})
			toMain(t, g)

			id := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: id, Name: "A Spell", TypeLine: "Sorcery",
				Colors: tc.colors, Owner: me.ID, Controller: me.ID,
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			if !tc.trigger {
				if latestPickTarget(g, me.ID) != nil {
					t.Error("a coloured spell must not trigger Ugin")
				}
				return
			}
			b04WaitForPick(t, g, me.ID)
			pickCard(t, g, me.ID, victim)
			passPriorityAroundTable(t, g)
			if !inExile(g, victim) {
				t.Error("the coloured permanent is exiled")
			}
			if loyaltyCount(g, ugin) != 7 {
				t.Errorf("the trigger costs no loyalty: %d", loyaltyCount(g, ugin))
			}
		})
	}
}

// TestUginEyeOfTheStormsPlusTwo: gain 3 life and draw a card.
func TestUginEyeOfTheStormsPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ugin := pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
	toMain(t, g)
	life, hand := me.Life, me.Hand.Size()

	if err := g.ActivateCatalogAbility(me.ID, ugin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	if loyaltyCount(g, ugin) != 9 {
		t.Errorf("loyalty = %d, want 9 — the cost is paid at announce", loyaltyCount(g, ugin))
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3", life, me.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand delta %d, want +1", me.Hand.Size()-hand)
	}
}

// TestUginEyeOfTheStormsZeroAddsThreeColorless: the 0 is a loyalty
// ability, so it USES THE STACK (CR 605.1a) — the mana arrives when
// it resolves, not when it is activated.
func TestUginEyeOfTheStormsZeroAddsThreeColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ugin := pushCatalogWalker(g, me.ID, "Ugin, Eye of the Storms", uginEyeOfTheStormsOracle, 7)
	toMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, ugin, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("0: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %+v, want empty until the ability resolves", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool = %+v, want {C}{C}{C}", me.ManaPool)
	}
	for _, m := range me.ManaPool {
		if m.Color != "C" {
			t.Errorf("pool = %+v, want colorless", me.ManaPool)
		}
	}
	if loyaltyCount(g, ugin) != 7 {
		t.Errorf("loyalty = %d, want 7 — a 0 costs nothing", loyaltyCount(g, ugin))
	}
}

// TestUginEyeOfTheStormsDeclaresItsMissingUltimate: the −11 is not
// registered, and the card says so rather than shipping two
// abilities and pretending the third was never printed.
func TestUginEyeOfTheStormsDeclaresItsMissingUltimate(t *testing.T) {
	spec, ok := Lookup(uginEyeOfTheStormsOracle)
	if !ok {
		t.Fatal("Ugin, Eye of the Storms is not registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("the missing ultimate must be declared: %+v", spec.Caveats)
	}
	if len(spec.Activated) != 2 {
		t.Errorf("%d loyalty abilities, want the +2 and the 0 only", len(spec.Activated))
	}
	if spec.StartingLoyalty != 0 {
		t.Error("starting loyalty is printed card data (ADR 0032) and must not be set here")
	}
}
