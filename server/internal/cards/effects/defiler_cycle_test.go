package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// defiler_cycle_test.go — #5659: the Defiler cycle of Dominaria United
// and the seam under it, an optional "pay 2 life" a permanent offers on
// coloured permanent spells with its linked coloured discount.

const (
	defilerOfDreamsOracle   = "9d4ec5e5-5520-45c3-8071-53191811117d"
	defilerOfFaithOracle    = "88ac2d09-7993-4bce-8aae-b2683c5c25a5"
	defilerOfFleshOracle    = "3ad002de-8be3-48d2-a0e0-865538144789"
	defilerOfInstinctOracle = "590137df-af61-4e93-b6f1-3a2e438b8037"
	defilerOfVigorOracle    = "0a1dc797-7531-4494-afe6-add6ac3f74c0"
)

func TestDefilersAreRegistered(t *testing.T) {
	for name, id := range map[string]string{
		"Defiler of Dreams": defilerOfDreamsOracle, "Defiler of Faith": defilerOfFaithOracle,
		"Defiler of Flesh": defilerOfFleshOracle, "Defiler of Instinct": defilerOfInstinctOracle,
		"Defiler of Vigor": defilerOfVigorOracle,
	} {
		spec, ok := Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Name = %q, Completeness = %v, Caveats = %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dfTable is a main-phase table with the named Defiler under the active
// seat's control.
func dfTable(t *testing.T, name, oracle string) (g *game.Game, me *game.Player, defiler uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	defiler = pushCatalogPermanent(g, me.ID, name, "Creature — Phyrexian", oracle, false)
	return g, me, defiler
}

// dfSpell puts a spell card with no catalog row into p's hand.
func dfSpell(g *game.Game, p *game.Player, name, typeLine, cost string, colors ...string) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() { id = handCardFull(p, name, typeLine, cost, "", colors) })
	return id
}

// dfPrice is the total the cast would charge with `chosen` announced.
func dfPrice(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, chosen ...int) string {
	t.Helper()
	c, ok := g.LookupCardForEffect(card)
	if !ok {
		t.Fatal("the card is gone")
	}
	price, err := g.PriceCast(p.ID, c, game.CastSpellParams{OptionalCosts: chosen})
	if err != nil {
		t.Fatalf("PriceCast%v: %v", chosen, err)
	}
	if s := price.Total.String(); s != "" {
		return s
	}
	return "{0}"
}

func dfResolveAll(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

// Paying 2 life takes {U} off a blue permanent spell: {1}{U}{U} is cast
// with {1}{U} in the pool, and the life is paid as the spell is cast
// (CR 601.2b, 601.2f, 601.2h).
func TestDefilerOfDreamsPayTwoLifeForBlueOff(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	spell := dfSpell(g, me, "Blue Bird", "Creature — Bird", "{1}{U}{U}", "U")
	if got := dfPrice(t, g, me, spell); got != "{1}{U}{U}" {
		t.Errorf("price with the offer declined = %s, want {1}{U}{U}", got)
	}
	if got := dfPrice(t, g, me, spell, 0); got != "{1}{U}" {
		t.Errorf("price with 2 life paid = %s, want {1}{U}", got)
	}
	floatMana(t, g, me, "{C}{U}")
	// Declined, the pool is one blue short.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast {1}{U}{U} with {1}{U} and no life paid")
	}
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast paying 2 life: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life went %d -> %d, want 2 paid", life, me.Life)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool", me.ManaPool.Total())
	}
	item := g.StackItemForEffect(spell)
	if item == nil || game.GrantedOptionalTimesPaid(item.Paid, game.PaidLifeForDiscountKey) != 1 {
		t.Fatalf("the payment is not on the spell's record: %+v", item)
	}
	// The record survives a restore with the spell on the stack.
	restored := restoreRoundTrip(t, g, false)
	if ri := restored.StackItemForEffect(spell); ri == nil || game.GrantedOptionalTimesPaid(ri.Paid, game.PaidLifeForDiscountKey) != 1 {
		t.Errorf("the payment did not survive a restore: %+v", ri)
	}
}

// "This effect reduces only the amount of blue mana you pay": nothing
// comes off the generic part, whatever CR 118.7b and 118.7c would do.
func TestDefilerDiscountTakesOnlyItsColor(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	for _, c := range []struct{ cost, want string }{
		{"{3}{U}", "{3}"},
		{"{U}", "{0}"},
		{"{4}", "{4}"},       // a blue spell with no blue mana in its cost
		{"{2}{G}", "{2}{G}"}, // blue by its colours, green in its cost
		{"{1}{G/U}", "{1}"},  // a hybrid symbol blue mana pays
		{"{X}{U}{U}", "{X}{U}"},
	} {
		spell := dfSpell(g, me, "Blue Thing "+c.cost, "Artifact", c.cost, "U")
		if got := dfPrice(t, g, me, spell, 0); got != c.want {
			t.Errorf("%s with 2 life paid costs %s, want %s", c.cost, got, c.want)
		}
	}
}

// The offer reaches blue permanent spells its controller casts and
// nothing else (CR 110.4b: an instant or sorcery is not a permanent
// spell).
func TestDefilerOfferReachesOnlyItsSpells(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	for _, c := range []struct {
		name, typeLine, cost string
		colors               []string
		offered              bool
	}{
		{"Blue Creature", "Creature — Bird", "{U}", []string{"U"}, true},
		{"Blue Enchantment", "Enchantment", "{U}", []string{"U"}, true},
		{"Blue Planeswalker", "Legendary Planeswalker — Test", "{U}", []string{"U"}, true},
		{"Gold Artifact", "Artifact", "{G}{U}", []string{"G", "U"}, true},
		{"Blue Instant", "Instant", "{U}", []string{"U"}, false},
		{"Blue Sorcery", "Sorcery", "{U}", []string{"U"}, false},
		{"Red Creature", "Creature — Goblin", "{R}", []string{"R"}, false},
		{"Colourless Artifact", "Artifact", "{2}", nil, false},
	} {
		spell := dfSpell(g, me, c.name, c.typeLine, c.cost, c.colors...)
		life := me.Life
		err := g.CastSpell(me.ID, spell, game.CastSpellParams{OptionalCosts: []int{0}})
		if c.offered != (err == nil) {
			t.Errorf("%s: offered = %v, cast error = %v", c.name, c.offered, err)
		}
		if err == nil && me.Life != life-2 {
			t.Errorf("%s: life went %d -> %d", c.name, life, me.Life)
		}
		if err != nil && me.Life != life {
			t.Errorf("%s: a refused cast paid life", c.name)
		}
		dfResolveAll(t, g)
	}
	// An opponent's blue permanent spell is not "you".
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := dfSpell(g, opp, "Their Bird", "Creature — Bird", "{U}", "U")
	c, _ := g.LookupCardForEffect(theirs)
	if _, err := g.PriceCast(opp.ID, c, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		// PriceCast does not validate the announcement; the reduction must
		// still not apply.
		if got := dfPrice(t, g, opp, theirs, 0); got != "{U}" {
			t.Errorf("an opponent's spell costs %s with index 0 announced, want {U}", got)
		}
	}
}

// Each Defiler makes its own offer and its own discount: two Defilers of
// Dreams are two payments of 2 life for {U} each, and one declined
// reduces nothing.
func TestTwoDefilersAreTwoOffers(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	pushCatalogPermanent(g, me.ID, "Defiler of Dreams", "Creature — Phyrexian Sphinx", defilerOfDreamsOracle, false)
	spell := dfSpell(g, me, "Blue Bird", "Creature — Bird", "{1}{U}{U}", "U")
	for _, c := range []struct {
		chosen []int
		want   string
	}{{nil, "{1}{U}{U}"}, {[]int{0}, "{1}{U}"}, {[]int{1}, "{1}{U}"}, {[]int{0, 1}, "{1}"}} {
		if got := dfPrice(t, g, me, spell, c.chosen...); got != c.want {
			t.Errorf("announcing %v costs %s, want %s", c.chosen, got, c.want)
		}
	}
	floatMana(t, g, me, "{C}")
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, OptionalCosts: []int{0, 1}}); err != nil {
		t.Fatalf("cast paying both: %v", err)
	}
	if me.Life != life-4 {
		t.Errorf("life went %d -> %d, want 4 paid", life, me.Life)
	}
}

// A white-blue permanent spell under a Defiler of each colour: the life
// paid to one takes only that one's colour.
func TestDefilersOfTwoColorsAreLinkedSeparately(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	pushCatalogPermanent(g, me.ID, "Defiler of Faith", "Creature — Phyrexian Human", defilerOfFaithOracle, false)
	spell := dfSpell(g, me, "Azorius Thing", "Enchantment", "{W}{U}", "W", "U")
	for _, c := range []struct {
		chosen []int
		want   string
	}{{nil, "{W}{U}"}, {[]int{0}, "{W}"}, {[]int{1}, "{U}"}, {[]int{0, 1}, "{0}"}} {
		if got := dfPrice(t, g, me, spell, c.chosen...); got != c.want {
			t.Errorf("announcing %v costs %s, want %s", c.chosen, got, c.want)
		}
	}
}

// A player who can't pay 2 life can't announce it (CR 118.3, 119.4).
func TestDefilerLifeMustBePayable(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	spell := dfSpell(g, me, "Blue Bird", "Creature — Bird", "{U}", "U")
	me.Life = 1
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		t.Fatal("paid 2 life at 1 life")
	}
	if me.Life != 1 {
		t.Errorf("a refused cast changed life to %d", me.Life)
	}
	// A life total that can't change pays no life either.
	me.Life = 20
	pushPermanentForTest(g, me.ID, "Platinum Emperion", platinumEmperionOracle, "Artifact Creature — Golem")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{OptionalCosts: []int{0}}); err == nil {
		t.Fatal("paid 2 life under Platinum Emperion")
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "cast_spell" && m.Source == spell && dfAnnounced(t, m.Params) > 0 {
			t.Error("the enumerator offers a life payment that can't be made")
		}
	}
}

func dfAnnounced(t *testing.T, params json.RawMessage) int {
	t.Helper()
	var p struct {
		OptionalCosts []int `json:"optional_costs"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		t.Fatal(err)
	}
	return len(p.OptionalCosts)
}

// The enumerator prices the discounted cast: with {1}{U} floating, a
// {1}{U}{U} blue creature is castable only with the life paid.
func TestDefilerDiscountIsEnumerated(t *testing.T) {
	g, me, _ := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	spell := dfSpell(g, me, "Blue Bird", "Creature — Bird", "{1}{U}{U}", "U")
	floatMana(t, g, me, "{C}{U}")
	paid, unpaid := 0, 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "cast_spell" || m.Source != spell {
			continue
		}
		if dfAnnounced(t, m.Params) > 0 {
			paid++
		} else {
			unpaid++
		}
	}
	if paid == 0 || unpaid != 0 {
		t.Errorf("moves paying life = %d, not paying = %d; want only the paid cast", paid, unpaid)
	}
}

// dfCastBlank casts a free permanent spell of the given colour and
// leaves it on the stack under whatever it triggered.
func dfCastBlank(t *testing.T, g *game.Game, me *game.Player, typeLine string, colors ...string) uuid.UUID {
	t.Helper()
	spell := dfSpell(g, me, "Blank", typeLine, "{0}", colors...)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return spell
}

func TestDefilerOfDreamsDrawsOnABluePermanentSpell(t *testing.T) {
	g, me, defiler := dfTable(t, "Defiler of Dreams", defilerOfDreamsOracle)
	hand := len(me.Hand.Cards)
	dfCastBlank(t, g, me, "Instant", "U")
	if triggerOnStack(g, defiler) != nil {
		t.Fatal("triggered on a blue instant")
	}
	dfResolveAll(t, g)
	dfCastBlank(t, g, me, "Creature — Bear", "G")
	if triggerOnStack(g, defiler) != nil {
		t.Fatal("triggered on a green creature")
	}
	dfResolveAll(t, g)
	dfCastBlank(t, g, me, "Enchantment", "U")
	if triggerOnStack(g, defiler) == nil {
		t.Fatal("no trigger on a blue permanent spell")
	}
	dfResolveAll(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand went %d -> %d, want one card drawn", hand, got)
	}
}

func TestDefilerOfFaithMakesASoldier(t *testing.T) {
	g, me, defiler := dfTable(t, "Defiler of Faith", defilerOfFaithOracle)
	dfCastBlank(t, g, me, "Sorcery", "W")
	if triggerOnStack(g, defiler) != nil {
		t.Fatal("triggered on a white sorcery")
	}
	dfResolveAll(t, g)
	dfCastBlank(t, g, me, "Artifact", "U")
	if triggerOnStack(g, defiler) != nil {
		t.Fatal("triggered on a blue artifact")
	}
	dfResolveAll(t, g)
	dfCastBlank(t, g, me, "Artifact", "W")
	if triggerOnStack(g, defiler) == nil {
		t.Fatal("no trigger on a white permanent spell")
	}
	dfResolveAll(t, g)
	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Soldier" && c.Controller == me.ID {
			soldiers++
			if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasColor("W") || !c.HasSubtype("Soldier") {
				t.Errorf("the token is %d/%d %v %q", c.CurrentPower(), c.CurrentToughness(), c.Colors, c.TypeLine)
			}
		}
	}
	if soldiers != 1 {
		t.Errorf("%d Soldier tokens, want 1", soldiers)
	}
}

func TestDefilerOfFleshPumpsATargetCreatureYouControl(t *testing.T) {
	g, me, defiler := dfTable(t, "Defiler of Flesh", defilerOfFleshOracle)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	// A creature you don't control is not a legal target.
	spell := dfSpell(g, me, "Black Thing", "Enchantment", "{0}", "B")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the trigger asked for no target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if triggerOnStack(g, defiler) == nil {
		t.Fatal("no trigger on the stack")
	}
	dfResolveAll(t, g)
	c, _ := g.LookupCardForEffect(bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !game.HasKeyword(&c, "menace") {
		t.Errorf("Bear is %d/%d, menace = %v; want 3/3 with menace", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "menace"))
	}
}

func TestDefilerOfInstinctDealsOneDamageToAnyTarget(t *testing.T) {
	g, me, defiler := dfTable(t, "Defiler of Instinct", defilerOfInstinctOracle)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dfCastBlank(t, g, me, "Instant", "R")
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, defiler) != nil {
		t.Fatal("triggered on a red instant")
	}
	dfResolveAll(t, g)
	dfCastBlank(t, g, me, "Creature — Goblin", "R")
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the trigger asked for no target")
	}
	before := opp.Life
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	dfResolveAll(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life went %d -> %d, want 1 damage", before, opp.Life)
	}
}

func TestDefilerOfVigorPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g, me, defiler := dfTable(t, "Defiler of Vigor", defilerOfVigorOracle)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	dfCastBlank(t, g, me, "Sorcery", "G")
	if triggerOnStack(g, defiler) != nil {
		t.Fatal("triggered on a green sorcery")
	}
	dfResolveAll(t, g)
	spell := dfCastBlank(t, g, me, "Creature — Elf", "G")
	if triggerOnStack(g, defiler) == nil {
		t.Fatal("no trigger on a green permanent spell")
	}
	dfResolveAll(t, g)
	counters := func(id uuid.UUID) int {
		c, _ := g.LookupCardForEffect(id)
		return c.Counters["+1/+1"]
	}
	if counters(bear) != 1 || counters(defiler) != 1 {
		t.Errorf("Bear has %d, the Defiler %d; want 1 each", counters(bear), counters(defiler))
	}
	// The spell that caused it was still on the stack, and an
	// opponent's creature is not yours.
	if counters(spell) != 0 || counters(theirs) != 0 {
		t.Errorf("the spell got %d, an opponent's creature %d; want none", counters(spell), counters(theirs))
	}
}
