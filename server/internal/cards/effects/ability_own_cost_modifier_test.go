package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ability_own_cost_modifier_test.go — #1296: an activated ability's
// OWN cost clause (ActivatedAbility.CostModifiers), and the price that
// reads the ability's target. ADR 0020 amendment 2026-09-24.
//
// The report was "Equip costs were not paid when equipping to Vivi
// Ornitier" with Dragonfire Blade. Two things were wrong and both are
// pinned here or in the client: the client never asked the engine to
// charge an activation (routes/Game.svelte, strictMana.test.ts), and
// the engine had no way to say "{1} less for each colour of the
// creature it targets". These are the cards, played the way the
// reporter played them, with the payment enforced (Strict).

const (
	ghostfireBladeOracle = "093b1d8d-c836-4e9b-855b-4020361aa6ac"
	warriorsBladesOracle = "f47419a5-b975-4934-a987-ea064a0c7c1a"
	otawaraOracle        = "e9b6a394-691c-425a-9307-76d8edc7375e"
	boseijuOracle        = "bf1341dd-41a3-49f6-87ec-63170dde4324"
)

// seedColoredCreature puts a 2/2 on the battlefield with the given
// colours (nil for colourless) and +1/+1 counters.
func seedColoredCreature(g *game.Game, owner uuid.UUID, name string, colors []string, plusOnes int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Colors: colors, Owner: owner, Controller: owner,
	}
	if plusOnes > 0 {
		c.Counters = map[string]int{game.CounterPlusOne: plusOnes}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

// strictEquip floats `mana` and activates `equipment`'s equip ability
// (its only activated ability) on `target` with the payment enforced.
// Returns the activation's error and leaves the item on the stack.
func strictEquip(t *testing.T, g *game.Game, controller, equipment, target uuid.UUID, mana string) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(controller, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	return g.ActivateCatalogAbility(controller, equipment, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		Strict:  true,
	})
}

func poolSize(g *game.Game, player uuid.UUID) int {
	for _, p := range g.Seats {
		if p.ID == player {
			return len(p.ManaPool)
		}
	}
	return -1
}

func attachedTo(t *testing.T, g *game.Game, equipment, host uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, equipment)
	if !ok {
		t.Fatalf("equipment %s left the battlefield", equipment)
	}
	return c.IsAttachedTo(host)
}

// --- Dragonfire Blade: the reporter's board ------------------------

// TestDragonfireBladeEquipsATwoColorCreatureForTwo is the report:
// Vivi Ornitier is blue and red, so the equip costs {4} − 2 = {2}, and
// {2} is what leaves the pool — not nothing (the bug) and not {4} (the
// old caveat).
func TestDragonfireBladeEquipsATwoColorCreatureForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vivi := seedColoredCreature(g, me.ID, "Vivi Ornitier", []string{"U", "R"}, 0)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	if err := strictEquip(t, g, me.ID, blade, vivi, "{C}{C}{C}"); err != nil {
		t.Fatalf("equip to a two-colour creature with three mana floating: %v", err)
	}
	if got := poolSize(g, me.ID); got != 1 {
		t.Errorf("pool after equip = %d mana, want 1 — the equip charges exactly {2}", got)
	}
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, blade, vivi) {
		t.Error("the Blade did not attach to Vivi")
	}
}

// TestDragonfireBladeChargesFullPriceForAColorlessCreature: the
// discount is per colour of the TARGET, so an artifact creature pays
// the printed {4} and {3} is refused with nothing spent.
func TestDragonfireBladeChargesFullPriceForAColorlessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	golem := seedColoredCreature(g, me.ID, "Golem", nil, 0)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	err := strictEquip(t, g, me.ID, blade, golem, "{C}{C}{C}")
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip to a colourless creature with {3}: err = %v, want insufficient mana", err)
	}
	if got := poolSize(g, me.ID); got != 3 {
		t.Errorf("a refused equip spent mana: pool = %d, want 3", got)
	}
	if err := strictEquip(t, g, me.ID, blade, golem, "{C}"); err != nil {
		t.Fatalf("equip to a colourless creature with {4}: %v", err)
	}
	if got := poolSize(g, me.ID); got != 0 {
		t.Errorf("pool after a full-price equip = %d, want 0", got)
	}
}

// TestDragonfireBladeFiveColorCreatureEquipsForFree: five colours take
// {5} off a {4}, and the engine's generic floor stops at zero.
func TestDragonfireBladeFiveColorCreatureEquipsForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sliver := seedColoredCreature(g, me.ID, "Sliver Queen", []string{"W", "U", "B", "R", "G"}, 0)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	if err := strictEquip(t, g, me.ID, blade, sliver, ""); err != nil {
		t.Fatalf("equip to a five-colour creature with an empty pool: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, blade, sliver) {
		t.Error("the Blade did not attach")
	}
}

// TestDragonfireBladePricesAtTheTargetsCurrentColors: the colours are
// read when the cost is determined (CR 602.2b → 601.2f), off the
// target the activation named — not off whatever the Blade is attached
// to now. A Blade on a mono-colour creature moving to a three-colour
// one pays {1}.
func TestDragonfireBladePricesAtTheTargetsCurrentColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	mono := seedColoredCreature(g, me.ID, "Mono", []string{"G"}, 0)
	three := seedColoredCreature(g, me.ID, "Three", []string{"B", "R", "G"}, 0)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	if err := strictEquip(t, g, me.ID, blade, mono, "{C}{C}{C}"); err != nil {
		t.Fatalf("equip to mono: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := strictEquip(t, g, me.ID, blade, three, "{C}"); err != nil {
		t.Fatalf("re-equip to a three-colour creature with {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !attachedTo(t, g, blade, three) {
		t.Error("the Blade did not move")
	}
}

// --- the menu, the enumerator and the wire -------------------------

// TestDragonfireBladeViewPricesEachLegalTarget: the row's
// charged_mana_cost is the no-target price ({4}), and the new
// target_charged_mana_costs carries each candidate's real price, from
// the same function the payment charges — so the client can show {2}
// on Vivi before the click that commits the activation.
func TestDragonfireBladeViewPricesEachLegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vivi := seedColoredCreature(g, me.ID, "Vivi Ornitier", []string{"U", "R"}, 0)
	golem := seedColoredCreature(g, me.ID, "Golem", nil, 0)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)

	row := activatedRowOf(t, g, me.ID, blade, 0)
	if row.ChargedManaCost == nil || *row.ChargedManaCost != "{4}" {
		t.Errorf("charged_mana_cost = %v, want {4} (the price before a target is chosen)", row.ChargedManaCost)
	}
	want := map[string]string{vivi.String(): "{2}", golem.String(): "{4}"}
	for id, price := range want {
		if got := row.TargetChargedManaCosts[id]; got != price {
			t.Errorf("target_charged_mana_costs[%s] = %q, want %q (all: %v)", id, got, price, row.TargetChargedManaCosts)
		}
	}
}

// TestPlainEquipShipsNoPerTargetPrices: an equip whose price does not
// read its target ships no map — the field is not a second copy of
// charged_mana_cost on every Equipment in the game.
func TestPlainEquipShipsNoPerTargetPrices(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedColoredCreature(g, me.ID, "Vivi Ornitier", []string{"U", "R"}, 0)
	saw := seedEquipment(g, me.ID, "Bone Saw", boneSawOracle)

	if row := activatedRowOf(t, g, me.ID, saw, 0); row.TargetChargedManaCosts != nil {
		t.Errorf("Bone Saw's equip shipped per-target prices %v", row.TargetChargedManaCosts)
	}
}

// TestDragonfireBladeEnumeratorPricesPerTarget: with two mana, the
// bot is offered the equip onto the two-colour creature and not onto
// the colourless one. A nil-targets price ({4}) as the gate would hide
// both (#544 with the sign reversed); a single price would offer both.
func TestDragonfireBladeEnumeratorPricesPerTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vivi := seedColoredCreature(g, me.ID, "Vivi Ornitier", []string{"U", "R"}, 0)
	golem := seedColoredCreature(g, me.ID, "Golem", nil, 0)
	blade := seedEquipment(g, me.ID, "Dragonfire Blade", dragonfireBladeOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}

	offered := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "activate_ability" || m.Source != blade {
			continue
		}
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		for _, tr := range p.Targets {
			if id, err := uuid.Parse(tr.ID); err == nil {
				offered[id] = true
			}
		}
	}
	if !offered[vivi] {
		t.Error("the enumerator withheld the {2} equip onto the two-colour creature")
	}
	if offered[golem] {
		t.Error("the enumerator offered a {4} equip with two mana available")
	}
}

// --- Ghostfire Blade: a condition on the target ---------------------

func TestGhostfireBladeEquipsAColorlessCreatureForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	golem := seedColoredCreature(g, me.ID, "Golem", nil, 0)
	red := seedColoredCreature(g, me.ID, "Goblin", []string{"R"}, 0)
	blade := seedEquipment(g, me.ID, "Ghostfire Blade", ghostfireBladeOracle)

	if err := strictEquip(t, g, me.ID, blade, golem, "{C}"); err != nil {
		t.Fatalf("equip to a colourless creature with {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, golem); got != 4 {
		t.Errorf("golem power %d, want 4 (2 + 2)", got)
	}
	err := strictEquip(t, g, me.ID, blade, red, "{C}{C}")
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip to a red creature with {2}: err = %v, want insufficient mana — the discount is for colourless targets only", err)
	}
}

// --- Warrior's Blades: a count on the target, and its ETB -----------

func TestWarriorsBladesEquipCostsLessPerCounterOnTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	grown := seedColoredCreature(g, me.ID, "Grown", []string{"G"}, 2)
	blades := seedEquipment(g, me.ID, "Warrior's Blades", warriorsBladesOracle)

	if err := strictEquip(t, g, me.ID, blades, grown, "{C}"); err != nil {
		t.Fatalf("equip to a creature with two +1/+1 counters with {1}: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The layered power: 2 base + 2 from the Blades. The two +1/+1
	// counters sit on top of it (CurrentPower), not in it.
	if got := effectivePower(t, g, grown); got != 4 {
		t.Errorf("layered power %d, want 4", got)
	}
	if !attachedTo(t, g, blades, grown) {
		t.Error("the Blades did not attach")
	}
}

func TestWarriorsBladesEntersDealingThreeAndGainingThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	lifeBefore, oppBefore := me.Life, opp.Life

	castCatalogSpell(t, g, "Warrior's Blades", equipTypeLine, warriorsBladesOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)

	if got := oppBefore - opp.Life; got != 3 {
		t.Errorf("damage to the opponent = %d, want 3", got)
	}
	if got := me.Life - lifeBefore; got != 3 {
		t.Errorf("life gained = %d, want 3", got)
	}
}

// --- the channel lands: an own clause that works from the HAND ------

// TestChannelDiscountAppliesFromTheHand pins the Takenuma fix. Its
// discount used to be a battlefield modifier, and a channel ability is
// activated from the hand, so the scan never found it: one legend and
// {B}{C}{C} was refused. The ability's own slot travels with it.
func TestChannelDiscountAppliesFromTheHand(t *testing.T) {
	cases := []struct {
		name, oracle, typeLine, mana string
		legends                      int
	}{
		{"Takenuma, Abandoned Mire", b02cTakenumaOracle, "Legendary Land", "{B}{C}{C}", 1},
		{"Otawara, Soaring City", otawaraOracle, "Legendary Land", "{U}", 3},
		{"Boseiju, Who Endures", boseijuOracle, "Legendary Land", "{G}", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			opp := g.Seats[1]
			advanceToMain(t, g)
			for i := 0; i < tc.legends; i++ {
				pushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
				})
			}
			// A target for the two channels that take one: an
			// opponent's artifact satisfies both Otawara's and
			// Boseiju's clauses.
			relic := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact",
				Owner: opp.ID, Controller: opp.ID,
			})
			var targets []game.TargetRef
			if tc.oracle != b02cTakenumaOracle {
				targets = []game.TargetRef{{Kind: game.TargetCard, ID: relic}}
			}
			id := pushCatalogHandCard(me, tc.name, tc.typeLine, tc.oracle)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.mana); err != nil {
				t.Fatal(err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: targets, Strict: true}); err != nil {
				t.Fatalf("channel with %s and %d legends: %v", tc.mana, tc.legends, err)
			}
			if got := poolSize(g, me.ID); got != 0 {
				t.Errorf("pool after channel = %d, want 0", got)
			}
		})
	}
}

// TestChannelDiscountNeverTouchesColoredMana: CR 601.2f reduces
// generic mana only, so Boseiju with five legends still costs {G}.
func TestChannelDiscountNeverTouchesColoredMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	advanceToMain(t, g)
	for i := 0; i < 5; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Human",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		})
	}
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact",
		Owner: opp.ID, Controller: opp.ID,
	})
	id := pushCatalogHandCard(me, "Boseiju, Who Endures", "Legendary Land", boseijuOracle)
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
		Strict:  true,
	})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("Boseiju with an empty pool and five legends: err = %v, want insufficient mana ({G} stays)", err)
	}
}

// --- Register refuses the shapes the engine would ignore ------------

func TestRegisterRefusesAnAbilityCostModifierWithNothingToPrice(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Register accepted a cost clause on an ability with no mana cost")
		}
	}()
	checkAbilityCostModifiers("Test Card", 0, ActivatedAbility{
		Label:         "{T}: Do a thing.",
		Cost:          game.AbilityCost{Tap: true},
		CostModifiers: []game.CostModifier{CostsLess(1, "This ability costs {1} less to activate.")},
	})
}

func TestRegisterRefusesAnAbilityCostFloor(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Register accepted a CostFloor as an ability's own clause")
		}
	}()
	checkAbilityCostModifiers("Test Card", 0, ActivatedAbility{
		Label:         "{2}: Do a thing.",
		Cost:          game.AbilityCost{Mana: "{2}"},
		CostModifiers: []game.CostModifier{CostsAtLeast(3, "floor")},
	})
}

// --- helpers --------------------------------------------------------

// activatedRowOf reads `source`'s activated ability row `index` from
// the viewer's frame.
func activatedRowOf(t *testing.T, g *game.Game, viewer, source uuid.UUID, index int) protocol.ActivatedAbilityView {
	t.Helper()
	// The seeding helpers push straight onto the battlefield without
	// the entry path's knower stamp, and the per-viewer filter redacts
	// a card nobody is marked as knowing. A permanent is public.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, p := range g.Seats {
				c.KnownBy[p.ID] = true
			}
		}
	})
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != source.String() {
			continue
		}
		for _, row := range cv.ActivatedAbilities {
			if row.Index == index {
				return row
			}
		}
	}
	t.Fatalf("no activated ability row %d on %s", index, source)
	return protocol.ActivatedAbilityView{}
}
