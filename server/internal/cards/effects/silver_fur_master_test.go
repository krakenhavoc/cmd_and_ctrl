package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// silver_fur_master_test.go — #5672: the ninjutsu keyword bit a cost
// modifier reads (game.AbilityCostSubject.Ninjutsu, ANinjutsuAbilityCost)
// and Silver-Fur Master, the card that prints "Ninjutsu abilities you
// activate cost {1} less to activate."

const silverFurMasterOracle = "60ec735d-c9be-4776-beb7-f603ae22edbf"

func pushSilverFurMaster(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Silver-Fur Master", "Creature — Rat Ninja", silverFurMasterOracle, 2, 2)
}

// sfmPrice is what `activator` is charged in mana for row `idx` of the
// card `id`, activated from `zone` — the one function the payment, the
// menu and the enumerator price an activation through.
func sfmPrice(t *testing.T, g *game.Game, activator, id uuid.UUID, zone game.ZoneKind, idx int) string {
	t.Helper()
	var q game.ParsedCost
	var err error
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("card %s not found", id)
		}
		rows, _ := game.ActivatedAbilitiesWithOrigins(c)
		q, err = g.AbilityManaCostForEffect(activator, c, zone, rows[idx])
	})
	if err != nil {
		t.Fatalf("AbilityManaCostForEffect: %v", err)
	}
	return q.String()
}

// Both constructors set the keyword bit, on a printed ability and on one
// a static grants to a card in hand, and nothing else carries it.
func TestNinjutsuAbilitiesCarryTheKeywordBit(t *testing.T) {
	for name, oracle := range map[string]string{
		"Ninja of the Deep Hours (ninjutsu)":   ninjaOfTheDeepHoursOracle,
		"Silver-Fur Master (ninjutsu)":         silverFurMasterOracle,
		"Yuriko (commander ninjutsu, 702.49d)": yurikoOracle,
	} {
		if ab := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})[0]; !ab.Ninjutsu {
			t.Errorf("%s: the ability does not carry the ninjutsu bit", name)
		}
	}
	if ab := game.ActivatedAbilitiesForCard(game.Card{OracleID: fearlessLiberatorOracle})[0]; ab.Ninjutsu {
		t.Error("a boast ability carries the ninjutsu bit")
	}

	g := newCatalogGame(t)
	me := g.Seats[0]
	card := zgHandCard(me, "Plain Creature", "Creature — Human")
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	rows, _ := zgRows(g, card)
	if len(rows) != 1 || !rows[0].Ninjutsu {
		t.Errorf("the ninjutsu Satoru Umezawa grants does not carry the bit: %v", zgLabels(rows))
	}
}

// The label and the bit must agree in both directions, on printed and
// granted rows.
func TestRegisterRefusesANinjutsuLabelAndBitThatDisagree(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	hand := Ninjutsu("{1}{U}")
	hand.Ninjutsu = false
	cases := map[string]Spec{
		"bit without the keyword": {Activated: []ActivatedAbility{{Label: "{1}: Do a thing.", Ninjutsu: true, Cost: ManaCost("{1}"), Effect: noop}}},
		"keyword without the bit": {Activated: []ActivatedAbility{hand}},
		"granted keyword without the bit": {Grants: []AbilityGrant{{
			Key: "test/ninjutsu", Zone: game.ZoneHand, Activated: []ActivatedAbility{hand},
		}}},
	}
	for name, spec := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", name)
				}
			}()
			spec.Name = "Test " + name
			checkNinjutsu(spec)
		}()
	}
	checkNinjutsu(Spec{Name: "Test ok", Activated: []ActivatedAbility{Ninjutsu("{1}{U}"), CommanderNinjutsu("{U}{B}")}})
}

// The headline, through a real activation from the hand: with the Master
// out, Ninja of the Deep Hours' {1}{U} ninjutsu is paid with {U} alone.
func TestSilverFurMasterMakesNinjutsuCostOneLess(t *testing.T) {
	g := newCatalogGame(t)
	pushSilverFurMaster(g, g.Seats[0].ID)
	ninja, attacker, me, opp := ninjutsuBoard(t, g, "Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{U}")

	if got := sfmPrice(t, g, me.ID, ninja, game.ZoneHand, 0); got != "{U}" {
		t.Errorf("ninjutsu {1}{U} is priced %s, want {U}", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("a {1}{U} ninjutsu with Silver-Fur Master out should cost {U}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool holds %d mana, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(ninja)
	if !ok || !g.Battlefield.Contains(ninja) || c.AttackingTarget != opp.ID {
		t.Error("the ninja did not arrive attacking")
	}
	// It is a Ninja the Master now pumps: 2/2 → 3/3.
	if p, tt := ptOf(t, g, ninja); p != 3 || tt != 3 {
		t.Errorf("the ninja is %d/%d under the Master, want 3/3", p, tt)
	}
}

// Without the Master the same pool cannot pay: the price is the printed
// one and the activation is refused with nothing spent.
func TestNinjutsuCostsItsPrintedPriceWithoutSilverFurMaster(t *testing.T) {
	g := newCatalogGame(t)
	ninja, attacker, me, _ := ninjutsuBoard(t, g, "Ninja of the Deep Hours", ninjaOfTheDeepHoursOracle, "{U}")
	if got := sfmPrice(t, g, me.ID, ninja, game.ZoneHand, 0); got != "{1}{U}" {
		t.Errorf("ninjutsu is priced %s, want {1}{U}", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker}, Strict: true,
	}); err == nil {
		t.Fatal("a {1}{U} ninjutsu was paid with {U} and no discount")
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the refused activation returned the attacker anyway")
	}
}

// What the clause does not reach: coloured mana (CR 118.7a), an
// opponent's ninjutsu ("you activate"), an ability that is not ninjutsu,
// and anything at all while the Master is still a card in hand.
func TestSilverFurMasterDiscountIsScopedAsPrinted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deepHours := pushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	inHand := pushNinjaToHand(me, "Silver-Fur Master", "Creature — Rat Ninja", silverFurMasterOracle, 2, 2)

	// A Master in hand has no static ability working (CR 113.6).
	if got := sfmPrice(t, g, me.ID, deepHours, game.ZoneHand, 0); got != "{1}{U}" {
		t.Errorf("with the Master only in hand, ninjutsu is priced %s, want {1}{U}", got)
	}

	pushSilverFurMaster(g, me.ID)
	if got := sfmPrice(t, g, me.ID, deepHours, game.ZoneHand, 0); got != "{U}" {
		t.Errorf("ninjutsu {1}{U} is priced %s, want {U}", got)
	}
	// {U}{B} has no generic mana to take off.
	if got := sfmPrice(t, g, me.ID, inHand, game.ZoneHand, 0); got != "{U}{B}" {
		t.Errorf("ninjutsu {U}{B} is priced %s, want {U}{B}", got)
	}
	// Commander ninjutsu is matched (CR 702.49d) and Yuriko's {U}{B} has
	// nothing to reduce either.
	yuriko := yurikoInCommandZone(me, 0)
	if got := sfmPrice(t, g, me.ID, yuriko, game.ZoneCommand, 0); got != "{U}{B}" {
		t.Errorf("commander ninjutsu {U}{B} is priced %s, want {U}{B}", got)
	}
	// An opponent's ninjutsu is priced in full.
	theirs := pushNinjaToHand(opp, "Ninja of the Deep Hours", "Creature — Human Ninja", ninjaOfTheDeepHoursOracle, 2, 2)
	if got := sfmPrice(t, g, opp.ID, theirs, game.ZoneHand, 0); got != "{1}{U}" {
		t.Errorf("an opponent's ninjutsu is priced %s, want {1}{U}", got)
	}
	// A boast ability is not a ninjutsu ability.
	dwarf := pushLiberator(g, me.ID)
	if got := sfmPrice(t, g, me.ID, dwarf, game.ZoneBattlefield, 0); got != "{2}{R}" {
		t.Errorf("a {2}{R} boast is priced %s, want {2}{R}", got)
	}
}

// A granted ninjutsu is a ninjutsu ability: Satoru Umezawa's {2}{U}{B}
// costs {1}{U}{B} under one Master and {U}{B} under two, and a snapshot
// round trip prices it the same.
func TestSilverFurMasterDiscountsAGrantedNinjutsuAndStacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	card := zgHandCard(me, "Plain Creature", "Creature — Human")
	b12Push(g, me.ID, "Satoru Umezawa", "Legendary Creature — Human Ninja", satoruUmezawaOracle, 2, 4)
	if got := sfmPrice(t, g, me.ID, card, game.ZoneHand, 0); got != "{2}{U}{B}" {
		t.Fatalf("setup: granted ninjutsu is priced %s, want {2}{U}{B}", got)
	}
	pushSilverFurMaster(g, me.ID)
	if got := sfmPrice(t, g, me.ID, card, game.ZoneHand, 0); got != "{1}{U}{B}" {
		t.Errorf("under one Master the granted ninjutsu is priced %s, want {1}{U}{B}", got)
	}
	restored := restoreRoundTrip(t, g, false)
	if got := sfmPrice(t, restored, me.ID, card, game.ZoneHand, 0); got != "{1}{U}{B}" {
		t.Errorf("restored: the granted ninjutsu is priced %s, want {1}{U}{B}", got)
	}
	pushSilverFurMaster(g, me.ID)
	if got := sfmPrice(t, g, me.ID, card, game.ZoneHand, 0); got != "{U}{B}" {
		t.Errorf("under two Masters the granted ninjutsu is priced %s, want {U}{B}", got)
	}
}

// "Other Ninja and Rogue creatures you control get +1/+1."
func TestSilverFurMasterPumpsOtherNinjasAndRogues(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := pushSilverFurMaster(g, me.ID)
	ninja := b12Creature(g, me.ID, "Test Ninja", "Creature — Human Ninja", 1, 1)
	rogue := b12Creature(g, me.ID, "Test Rogue", "Creature — Faerie Rogue", 1, 1)
	both := b12Creature(g, me.ID, "Test Ninja Rogue", "Creature — Ninja Rogue", 1, 1)
	rat := b12Creature(g, me.ID, "Test Rat", "Creature — Rat", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Ninja", "Creature — Human Ninja", 1, 1)

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{
		{"the Master itself (other)", master, 2, 2},
		{"a Ninja you control", ninja, 2, 2},
		{"a Rogue you control", rogue, 2, 2},
		{"a Ninja Rogue, pumped once", both, 2, 2},
		{"a Rat that is neither", rat, 1, 1},
		{"an opponent's Ninja", theirs, 1, 1},
	} {
		if p, tt := ptOf(t, g, tc.id); p != tc.p || tt != tc.t {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tt, tc.p, tc.t)
		}
	}

	// Two Masters pump each other.
	second := pushSilverFurMaster(g, me.ID)
	for _, id := range []uuid.UUID{master, second} {
		if p, tt := ptOf(t, g, id); p != 3 || tt != 3 {
			t.Errorf("a Master beside another is %d/%d, want 3/3", p, tt)
		}
	}
	if p, tt := ptOf(t, g, ninja); p != 3 || tt != 3 {
		t.Errorf("a Ninja under two Masters is %d/%d, want 3/3", p, tt)
	}
}

// The Master's own ninjutsu: {U}{B} from the hand, arriving attacking.
func TestSilverFurMasterEntersByItsOwnNinjutsu(t *testing.T) {
	g := newCatalogGame(t)
	master, attacker, me, opp := ninjutsuBoard(t, g, "Silver-Fur Master", silverFurMasterOracle, "{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(master)
	if !ok || !g.Battlefield.Contains(master) {
		t.Fatal("the Master is not on the battlefield")
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("the Master entered tapped=%v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, opp.ID)
	}
}
