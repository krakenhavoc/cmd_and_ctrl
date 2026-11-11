package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// raiding_schemes_tap_cost_test.go — #5688: conspire on a spell whose own
// cost taps permanents. The spell's own taps ride teamwork_ids and
// conspire's two creatures ride conspire_ids, so both costs are paid on
// one cast (CR 601.2b, 601.2h) with different permanents: a tapped
// creature can't be tapped again to pay a cost (CR 118.3).

// rsTapTable is tapTable with Raiding Schemes under the caster's control.
func rsTapTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g, me, opp = tapTable(t)
	pushCatalogPermanent(g, me.ID, "Raiding Schemes", "Enchantment", raidingSchemesOracle, false)
	return g, me, opp
}

// rsOptionalCosts is the optional costs `p` is offered for the card
// `id` in their hand: its own, then the granted ones.
func rsOptionalCosts(g *game.Game, p *game.Player, id uuid.UUID) (out []game.AdditionalCost) {
	g.WithWriteLock(func() {
		for _, c := range p.Hand.Cards {
			if c.InstanceID == id {
				out = g.OptionalCostsForCastLocked(p.ID, c)
			}
		}
	})
	return out
}

// rsConspireTriggers counts the conspire triggers of `spell` on the stack.
func rsConspireTriggers(g *game.Game, spell uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == spell {
			n++
		}
	}
	return n
}

// New Blood taps a Vampire for its own cost and, under Raiding Schemes,
// two more black creatures for conspire: three creatures tapped, the
// Vampire alone on the spell's own payment record, and a conspire trigger.
func TestNewBloodIsConspiredBesideItsOwnTapCost(t *testing.T) {
	g, me, opp := rsTapTable(t)
	blood := handCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", newBloodOracle)
	vamp := tacVampire(g, me.ID, "Vampire")
	a := pr7Creature(g, me.ID, "Rat A", 1, "B")
	b := pr7Creature(g, me.ID, "Rat B", 1, "B", "G")
	goblin := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2, Colors: []string{"R"}})
	g.WithWriteLock(func() {
		fillPoolColored(me, "B", 2)
		fillPool(me, 2)
	})

	if err := g.CastSpell(me.ID, blood, game.CastSpellParams{
		Targets:       cardRefs(goblin),
		OptionalCosts: []int{0},
		TeamworkIDs:   []uuid.UUID{vamp},
		ConspireIDs:   []uuid.UUID{a, b},
		Strict:        true,
	}); err != nil {
		t.Fatalf("New Blood with conspire: %v", err)
	}
	for name, id := range map[string]uuid.UUID{"the Vampire": vamp, "Rat A": a, "Rat B": b} {
		if !tappedForTest(t, g, id) {
			t.Errorf("%s was not tapped", name)
		}
	}
	item := g.StackMeta[blood]
	if item == nil || len(item.Paid.TappedOthers) != 1 || item.Paid.TappedOthers[0].ID != vamp {
		t.Errorf("the payment records %+v as tapped for New Blood's own cost, want the Vampire alone", item)
	}
	if got := game.GrantedOptionalTimesPaid(item.Paid, game.ConspireKey); got != 1 {
		t.Errorf("conspire recorded as paid %d times, want 1", got)
	}
	if n := rsConspireTriggers(g, blood); n != 1 {
		t.Errorf("%d conspire triggers, want 1", n)
	}
}

// One creature cannot pay both costs, and neither cost can be paid short:
// every such cast is refused with nothing tapped and the spell in hand.
func TestOneCreatureCannotPayNewBloodAndConspire(t *testing.T) {
	g, me, opp := rsTapTable(t)
	blood := handCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", newBloodOracle)
	vamp := tacVampire(g, me.ID, "Vampire")
	a := pr7Creature(g, me.ID, "Rat A", 1, "B")
	b := pr7Creature(g, me.ID, "Rat B", 1, "B")
	white := pr7Creature(g, me.ID, "Human", 1, "W")
	goblin := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	g.WithWriteLock(func() {
		fillPoolColored(me, "B", 2)
		fillPool(me, 2)
	})
	type lists struct{ team, conspire []uuid.UUID }
	for name, l := range map[string]lists{
		"the Vampire on both lists":              {[]uuid.UUID{vamp}, []uuid.UUID{vamp, a}},
		"conspire with one creature":             {[]uuid.UUID{vamp}, []uuid.UUID{a}},
		"conspire with a creature of no colour":  {[]uuid.UUID{vamp}, []uuid.UUID{a, white}},
		"conspire's creatures on teamwork_ids":   {[]uuid.UUID{vamp, a, b}, nil},
		"conspire paid and New Blood's own not":  {nil, []uuid.UUID{a, b}},
		"New Blood's Vampire on conspire's list": {nil, []uuid.UUID{vamp, a}},
	} {
		if err := g.CastSpell(me.ID, blood, game.CastSpellParams{
			Targets:       cardRefs(goblin),
			OptionalCosts: []int{0},
			TeamworkIDs:   l.team,
			ConspireIDs:   l.conspire,
			Strict:        true,
		}); err == nil {
			t.Fatalf("%s: the cast was accepted", name)
		}
	}
	// conspire_ids with no conspire announced is refused, not ignored.
	if err := g.CastSpell(me.ID, blood, game.CastSpellParams{
		Targets: cardRefs(goblin), TeamworkIDs: []uuid.UUID{vamp}, ConspireIDs: []uuid.UUID{a, b}, Strict: true,
	}); err == nil {
		t.Fatal("conspire_ids was accepted with no conspire cost announced")
	}
	for _, id := range []uuid.UUID{vamp, a, b, white} {
		if tappedForTest(t, g, id) {
			t.Fatal("a refused cast tapped a creature")
		}
	}
	if !me.Hand.Contains(blood) {
		t.Fatal("a refused cast moved the spell")
	}
	// Unconspired, the cast is what it always was.
	if err := tacCast(g, me, blood, cardRefs(goblin), vamp); err != nil {
		t.Fatalf("New Blood without conspire under Raiding Schemes: %v", err)
	}
	if rsConspireTriggers(g, blood) != 0 || tappedForTest(t, g, a) {
		t.Error("an unconspired cast tapped a creature for conspire or made a trigger")
	}
}

// Burn at the Stake counts only the creatures its own cost tapped: one
// creature tapped for it and two for conspire is three damage and a copy
// for three more, not nine.
func TestConspiredBurnAtTheStakeCountsOnlyItsOwnTaps(t *testing.T) {
	g, me, opp := rsTapTable(t)
	burn := handCardOf(g, me, "Burn at the Stake", "Sorcery", "{2}{R}{R}{R}", burnAtTheStakeOracle)
	team := taoCreatures(g, me.ID, 3)
	g.WithWriteLock(func() {
		fillPoolColored(me, "R", 3)
		fillPool(me, 2)
	})
	life := opp.Life
	if err := g.CastSpell(me.ID, burn, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		TeamworkIDs:   team[:1],
		ConspireIDs:   team[1:],
		Strict:        true,
	}); err != nil {
		t.Fatalf("Burn at the Stake with conspire: %v", err)
	}
	for _, id := range team {
		if !tappedForTest(t, g, id) {
			t.Fatal("a creature named to a cost was not tapped")
		}
	}
	sqResolveAll(t, g, me)
	if lost := life - opp.Life; lost != 6 {
		t.Errorf("the opponent lost %d, want 6: three for the one creature tapped to the spell, and its copy's three", lost)
	}
}

// A replicate cost that taps and conspire on one cast: Psionic Ritual's
// Horror rides teamwork_ids, conspire's two blue creatures conspire_ids,
// and the Horror cannot be one of them.
func TestPsionicRitualPaysReplicateAndConspireWithDifferentCreatures(t *testing.T) {
	g, me, _ := rsTapTable(t)
	ritual := handCardOf(g, me, "Psionic Ritual", "Sorcery", "{4}{U}{U}", "1c3e6980-97fe-4e15-93a0-368280da8608")
	horror := apaPush(g, me.ID, me.ID, game.Card{Name: "Horror", TypeLine: "Creature — Horror", Power: 2, Toughness: 2, Colors: []string{"U"}})
	a := pr7Creature(g, me.ID, "Drake A", 1, "U")
	b := pr7Creature(g, me.ID, "Drake B", 1, "U")
	target := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: target, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
		fillPoolColored(me, "U", 2)
		fillPool(me, 4)
	})
	optional := rsOptionalCosts(g, me, ritual)
	if len(optional) != 2 || optional[0].Key != game.ReplicateKey || optional[1].Key != game.ConspireKey {
		t.Fatalf("optional costs %+v, want replicate then the granted conspire", optional)
	}
	cast := func(team, conspire []uuid.UUID) error {
		return g.CastSpell(me.ID, ritual, game.CastSpellParams{
			Targets:       cardRefs(target),
			OptionalCosts: []int{0, 1},
			TeamworkIDs:   team,
			ConspireIDs:   conspire,
			Strict:        true,
		})
	}
	if err := cast([]uuid.UUID{horror}, []uuid.UUID{horror, a}); err == nil {
		t.Fatal("the Horror paid replicate and conspire")
	}
	if tappedForTest(t, g, horror) || tappedForTest(t, g, a) {
		t.Fatal("a refused cast tapped a creature")
	}
	if err := cast([]uuid.UUID{horror}, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("Psionic Ritual with replicate and conspire: %v", err)
	}
	if !tappedForTest(t, g, horror) || !tappedForTest(t, g, a) || !tappedForTest(t, g, b) {
		t.Error("not all three creatures were tapped")
	}
	// Both costs are on the record, each where its trigger reads it.
	paid := g.StackItemPaidForEffect(ritual)
	if !slices.Equal(paid.OptionalCosts, []int{0}) || game.GrantedOptionalTimesPaid(paid, game.ConspireKey) != 1 {
		t.Errorf("payment record %+v, want replicate paid once and conspire paid once", paid)
	}
	if len(paid.TappedOthers) != 1 || paid.TappedOthers[0].ID != horror {
		t.Errorf("tapped for the spell's own cost: %+v, want the Horror alone", paid.TappedOthers)
	}
}

// The enumerator offers New Blood with conspire only when three creatures
// can pay both costs, names the two lists apart, and the engine accepts
// the move it offers.
func TestConspireBesideATapCostIsEnumerated(t *testing.T) {
	g, me, opp := rsTapTable(t)
	blood := handCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", newBloodOracle)
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	vamp := tacVampire(g, me.ID, "Vampire")
	a := pr7Creature(g, me.ID, "Rat A", 1, "B")
	g.WithWriteLock(func() {
		fillPoolColored(me, "B", 2)
		fillPool(me, 2)
	})
	type wire struct {
		OptionalCosts []int       `json:"optional_costs"`
		TeamworkIDs   []uuid.UUID `json:"teamwork_ids"`
		ConspireIDs   []uuid.UUID `json:"conspire_ids"`
	}
	moves := func() (plain, conspired []legal.Move) {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != legal.TypeCastSpell || m.Source != blood {
				continue
			}
			var p wire
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if len(p.OptionalCosts) > 0 {
				conspired = append(conspired, m)
			} else {
				plain = append(plain, m)
			}
		}
		return plain, conspired
	}
	// The Vampire and one Rat: the Vampire pays New Blood, and one Rat is
	// not two creatures.
	plain, conspired := moves()
	if len(plain) == 0 {
		t.Fatal("New Blood is not offered without conspire")
	}
	if len(conspired) != 0 {
		t.Fatalf("conspire is offered with two creatures for three taps: %s", conspired[0].Label)
	}
	b := pr7Creature(g, me.ID, "Rat B", 1, "B")
	_, conspired = moves()
	if len(conspired) == 0 {
		t.Fatal("no conspired New Blood with a Vampire and two black creatures")
	}
	for _, m := range conspired {
		var p wire
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.TeamworkIDs) != 1 || len(p.ConspireIDs) != 2 || slices.Contains(p.ConspireIDs, p.TeamworkIDs[0]) ||
			p.ConspireIDs[0] == p.ConspireIDs[1] {
			t.Fatalf("move %q names teamwork_ids %v and conspire_ids %v, want one Vampire and two other creatures", m.Label, p.TeamworkIDs, p.ConspireIDs)
		}
	}
	m := conspired[0]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: me.ID, Caller: me.ID, Params: m.Params}); err != nil {
		t.Fatalf("the enumerated conspired cast is refused: %v", err)
	}
	if !tappedForTest(t, g, vamp) || !tappedForTest(t, g, a) || !tappedForTest(t, g, b) {
		t.Error("the enumerated cast did not tap all three creatures")
	}
	if rsConspireTriggers(g, blood) != 1 {
		t.Error("no conspire trigger for the enumerated cast")
	}
}

// CR 702.78b: two Raiding Schemes give a spell two instances of conspire,
// each paid with its own two creatures and each its own trigger. One paid
// is one copy; both paid are two.
func TestTwoRaidingSchemesGiveTwoConspireCosts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		announce []int
		taps     int
		lost     int
	}{
		{"one instance paid", []int{0}, 2, 6},
		{"both instances paid", []int{0, 1}, 4, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, bolt := rsTable(t)
			pushCatalogPermanent(g, me.ID, "Raiding Schemes", "Enchantment", raidingSchemesOracle, false)
			opp := g.Seats[1]
			var red []uuid.UUID
			for i := 0; i < 4; i++ {
				red = append(red, rsColoredCreature(g, me.ID, "Goblin", "R"))
			}
			if got := rsOptionalCosts(g, me, bolt); len(got) != 2 {
				t.Fatalf("%d optional costs with two Raiding Schemes, want two conspire costs", len(got))
			}
			before := opp.Life
			// Two instances announced and two creatures named is one
			// cost short.
			if len(tc.announce) == 2 {
				if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
					Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
					OptionalCosts: tc.announce,
					ConspireIDs:   red[:2],
				}); err == nil {
					t.Fatal("two conspire costs were paid with two creatures")
				}
			}
			if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
				Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
				OptionalCosts: tc.announce,
				ConspireIDs:   red[:tc.taps],
			}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			tapped := 0
			for _, id := range red {
				if rsTapped(g, id) {
					tapped++
				}
			}
			if tapped != tc.taps {
				t.Errorf("%d creatures tapped, want %d", tapped, tc.taps)
			}
			if n := rsConspireTriggers(g, bolt); n != len(tc.announce) {
				t.Errorf("%d conspire triggers, want one per cost paid (%d)", n, len(tc.announce))
			}
			sqResolveAll(t, g, me)
			if lost := before - opp.Life; lost != tc.lost {
				t.Errorf("the opponent lost %d, want %d", lost, tc.lost)
			}
		})
	}
}

// The view offers conspire on a spell whose own cost taps permanents: New
// Blood in hand carries its own tap picker and a conspire offer whose
// options are the untapped creatures that share its colour.
func TestTheViewOffersConspireOnASpellWithATapCost(t *testing.T) {
	g, me, _ := rsTapTable(t)
	blood := handCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", newBloodOracle)
	vamp := tacVampire(g, me.ID, "Vampire")
	rat := pr7Creature(g, me.ID, "Rat", 1, "B")
	pr7Creature(g, me.ID, "Human", 1, "W")

	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			c := &s.Hand.Cards[i]
			if c.InstanceID != blood.String() {
				continue
			}
			if c.AdditionalCost == nil || c.AdditionalCost.TapOthersOptions == nil {
				t.Fatal("New Blood lost its own tap picker")
			}
			if len(c.OptionalCosts) != 1 || c.OptionalCosts[0].Key != game.ConspireKey {
				t.Fatalf("optional_costs = %+v, want the granted conspire", c.OptionalCosts)
			}
			to := c.OptionalCosts[0].TapOthersOptions
			if to == nil || to.Min != 2 || to.Max != 2 {
				t.Fatalf("conspire's tap_others_options = %+v, want 2/2", to)
			}
			got := slices.Clone(to.Cards)
			want := []string{vamp.String(), rat.String()}
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("conspire options %v, want the two black creatures %v", got, want)
			}
			return
		}
	}
	t.Fatal("New Blood is not in the caster's hand view")
}
