package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// energy_alt_cost_cards_test.go — ADR 0129 PR 4 (#1995): energy as an
// alternative cost (Nissa, Worldsoul Speaker; Primal Prayers; Amped
// Raptor), replicate (Reiterating Bolt) and a keyword's cost (Inventor's
// Axe's equip, Salvation Colossus's unearth), each through the real
// catalog and the real cast or activation path.

const (
	nissaWorldsoulOracle    = "00037840-6089-42ec-8c5c-281f9f474504"
	primalPrayersOracle     = "8d087fe0-d554-4d7c-ba22-32db2cf71887"
	ampedRaptorOracle       = "3ed14ea7-ee62-4604-af3b-deafb00108c3"
	reiteratingBoltOracle   = "371da0d8-3845-458e-9b0a-6477017b890e"
	inventorsAxeOracle      = "364fc1a7-b91d-470b-80fe-275828f36cfc"
	salvationColossusOracle = "8b5c6754-36bd-44ac-b0c0-5a3cd2d9f2c3"
)

func TestEnergyAltCostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{nissaWorldsoulOracle, primalPrayersOracle, ampedRaptorOracle,
		reiteratingBoltOracle, inventorsAxeOracle, salvationColossusOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// eacHandCard puts a plain card in p's hand.
func eacHandCard(g *game.Game, p *game.Player, name, typeLine, mana string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: mana,
			Owner: p.ID, Controller: p.ID})
	})
	return id
}

// Nissa: eight energy pays for a permanent spell from hand under the
// strict gate with no mana at all; seven is not offered and is refused;
// an instant is never offered.
func TestNissaWorldsoulSpeakerOffersEightEnergyForPermanentSpells(t *testing.T) {
	g, me, _ := p7Table(t)
	apaPush(g, me.ID, me.ID, game.Card{Name: "Nissa, Worldsoul Speaker", OracleID: nissaWorldsoulOracle,
		TypeLine: "Legendary Creature — Elf Druid", Power: 3, Toughness: 3})
	big := eacHandCard(g, me, "Big Creature", "Creature — Test", "{6}{G}{G}")
	instant := eacHandCard(g, me, "Some Instant", "Instant", "{1}{U}")

	setEnergy(t, g, me, 7)
	if o := offerWithKey(grantedCardOffers(g, me.ID, big, game.ZoneHand), GrantedAltCostEightEnergyPermanents); o != nil {
		t.Fatalf("offered with seven energy: %+v", o)
	}
	err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostEightEnergyPermanents})
	if err == nil {
		t.Fatal("the claim was accepted with seven energy")
	}
	if energyOf(me) != 7 || !me.Hand.Contains(big) {
		t.Fatalf("a refused claim paid something: energy %d, in hand %v", energyOf(me), me.Hand.Contains(big))
	}

	setEnergy(t, g, me, 8)
	o := offerWithKey(grantedCardOffers(g, me.ID, big, game.ZoneHand), GrantedAltCostEightEnergyPermanents)
	if o == nil || o.Energy != 8 || o.ManaCost != "" || !o.Granted {
		t.Fatalf("offer with eight energy = %+v", o)
	}
	if offerWithKey(grantedCardOffers(g, me.ID, instant, game.ZoneHand), GrantedAltCostEightEnergyPermanents) != nil {
		t.Error("an instant was offered the permanent-spell price")
	}
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostEightEnergyPermanents}); err != nil {
		t.Fatalf("cast for eight energy: %v", err)
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d after the cast, want 0", energyOf(me))
	}

	// ADR 0129 §4: Cast anyway waives only the mana. The eight energy is
	// still charged.
	passPriorityAroundTable(t, g)
	other := eacHandCard(g, me, "Other Creature", "Creature — Test", "{6}{G}{G}")
	setEnergy(t, g, me, 8)
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true, ForceCast: true, AlternativeCost: GrantedAltCostEightEnergyPermanents}); err != nil {
		t.Fatalf("Cast anyway for eight energy: %v", err)
	}
	if energyOf(me) != 0 {
		t.Errorf("Cast anyway left energy %d, want 0", energyOf(me))
	}
}

// Primal Prayers: creature spells with mana value 3 or less only, one
// energy, and as though they had flash only when cast this way (CR
// 601.3c): out of a main phase the printed cost is refused for timing
// and the claim is accepted.
func TestPrimalPrayersCastsSmallCreaturesForOneEnergyWithFlash(t *testing.T) {
	g, me, _ := p7Table(t)
	apaPush(g, me.ID, me.ID, game.Card{Name: "Primal Prayers", OracleID: primalPrayersOracle, TypeLine: "Enchantment"})
	three := eacHandCard(g, me, "Three Drop", "Creature — Test", "{2}{G}")
	four := eacHandCard(g, me, "Four Drop", "Creature — Test", "{3}{G}")
	sorcery := eacHandCard(g, me, "Cheap Sorcery", "Sorcery", "{G}")
	setEnergy(t, g, me, 2)

	if o := offerWithKey(grantedCardOffers(g, me.ID, three, game.ZoneHand), GrantedAltCostEnergySmallCreatures); o == nil || o.Energy != 1 || !o.AsThoughFlash {
		t.Fatalf("offer on a mana value 3 creature = %+v", o)
	}
	if offerWithKey(grantedCardOffers(g, me.ID, four, game.ZoneHand), GrantedAltCostEnergySmallCreatures) != nil {
		t.Error("a mana value 4 creature was offered the price")
	}
	if offerWithKey(grantedCardOffers(g, me.ID, sorcery, game.ZoneHand), GrantedAltCostEnergySmallCreatures) != nil {
		t.Error("a sorcery was offered the price")
	}

	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("printed cost in beginning of combat: err = %v, want ErrSorcerySpeedRequired", err)
	}

	// The enumerator offers the flash cast, and only through the offer,
	// with its one energy on the move's cost; the move is accepted.
	var found []legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "cast_spell" && m.Source == three {
			found = append(found, m)
		}
	}
	if len(found) == 0 {
		t.Fatal("no instant-speed cast of the three-drop was enumerated")
	}
	for _, m := range found {
		var p struct {
			AlternativeCost string `json:"alternative_cost"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.AlternativeCost != GrantedAltCostEnergySmallCreatures || m.Cost == nil || m.Cost.Energy != 1 {
			t.Errorf("move %q: claim %q cost %+v, want the energy offer paying 1", m.Label, p.AlternativeCost, m.Cost)
		}
		clone := g.Clone()
		if err := actions.Dispatch(clone, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}

	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Strict: true, AlternativeCost: GrantedAltCostEnergySmallCreatures}); err != nil {
		t.Fatalf("flash cast for one energy: %v", err)
	}
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", energyOf(me))
	}
}

// Amped Raptor: cast from hand, its trigger gives two energy and exiles
// down to a nonland card, which may be cast for energy equal to its mana
// value and for nothing else.
func TestAmpedRaptorCastsTheExiledCardForEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	hit := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	})
	castCatalogSpell(t, g, "Amped Raptor", "Creature — Dinosaur", ampedRaptorOracle, nil)
	passPriorityAroundTable(t, g)

	if energyOf(me) != 2 {
		t.Fatalf("energy = %d after the trigger, want 2", energyOf(me))
	}
	if !g.Exile.Contains(hit) || !g.Exile.Contains(land) {
		t.Fatalf("the land and the two-drop are not both in exile")
	}
	o := offerWithKey(grantedCardOffers(g, me.ID, hit, game.ZoneExile), ampedRaptorAltCostKey)
	if o == nil || o.Energy != 2 || o.ManaCost != "" {
		t.Fatalf("offer on the exiled two-drop = %+v", o)
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Fatal("the exiled card was cast without claiming the energy price")
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: ampedRaptorAltCostKey}); err != nil {
		t.Fatalf("cast for two energy: %v", err)
	}
	if energyOf(me) != 0 {
		t.Errorf("energy = %d after the cast, want 0", energyOf(me))
	}
}

// Amped Raptor that was not cast from hand gives its energy and exiles
// nothing.
func TestAmpedRaptorNotCastFromHandOnlyGivesEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	before := me.Library.Size()
	raptor := apaPush(g, me.ID, me.ID, game.Card{Name: "Amped Raptor", OracleID: ampedRaptorOracle,
		TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 1})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: raptor, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
	if energyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", energyOf(me))
	}
	if me.Library.Size() != before {
		t.Errorf("a Raptor not cast from hand exiled cards")
	}
}

// Reiterating Bolt: replicate is paid three energy at a time, the cast
// is refused past the caster's energy with nothing paid, and each
// payment is one copy.
func TestReiteratingBoltReplicatesForThreeEnergyEach(t *testing.T) {
	g, me, opp := p7Table(t)
	a := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	b := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear B", TypeLine: "Creature — Bear", Power: 2, Toughness: 6})
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Reiterating Bolt", TypeLine: "Sorcery", ManaCost: "{1}{R}",
			OracleID: reiteratingBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}}

	setEnergy(t, g, me, 8)
	err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: targets, OptionalCosts: []int{0, 0, 0}})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("three replicates with eight energy: err = %v, want ErrInsufficientEnergy", err)
	}
	if energyOf(me) != 8 || !me.Hand.Contains(bolt) {
		t.Fatalf("a refused cast paid something: energy %d, in hand %v", energyOf(me), me.Hand.Contains(bolt))
	}

	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: targets, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("two replicates with eight energy: %v", err)
	}
	if energyOf(me) != 2 {
		t.Errorf("energy = %d after two replicates, want 2", energyOf(me))
	}
	if it := triggerOnStack(g, bolt); it == nil || it.XValue != 2 {
		t.Fatalf("replicate trigger = %+v, want one counting two payments", it)
	}
	prompts := 0
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) && latestPickTarget(g, me.ID) == nil {
			break
		}
		if p := latestPickTarget(g, me.ID); p != nil {
			prompts++
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: b}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if prompts != 2 {
		t.Errorf("%d re-target prompts, want 2 (one per copy)", prompts)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Errorf("3 damage to A and 6 to B should kill both: A %v, B %v", g.Battlefield.Contains(a), g.Battlefield.Contains(b))
	}
}

// A Reiterating Bolt cast with no replicate paid has no trigger (CR
// 702.56a's "if a replicate cost was paid", CR 603.4).
func TestReiteratingBoltWithoutReplicateHasNoTrigger(t *testing.T) {
	g, _, opp := p7Table(t)
	a := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	bolt := castCatalogSpell(t, g, "Reiterating Bolt", "Sorcery", reiteratingBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	if triggerOnStack(g, bolt) != nil {
		t.Fatal("a replicate trigger with no replicate paid")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the bear survived 3 damage")
	}
}

// ADR 0129 §7: the enumerator never offers a replicate count the seat
// cannot pay for, and names the energy on the move.
func TestReiteratingBoltEnumeratesOnlyAffordableReplicates(t *testing.T) {
	g, me, opp := p7Table(t)
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Reiterating Bolt", TypeLine: "Sorcery", ManaCost: "{1}{R}",
			OracleID: reiteratingBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{R}{R}")
	setEnergy(t, g, me, 4)
	seen := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != "cast_spell" || m.Source != bolt {
			continue
		}
		var p struct {
			OptionalCosts []int `json:"optional_costs"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		n := len(p.OptionalCosts)
		seen[n] = true
		want := 3 * n
		got := 0
		if m.Cost != nil {
			got = m.Cost.Energy
		}
		if got != want {
			t.Errorf("move %q: energy %d, want %d", m.Label, got, want)
		}
	}
	if !seen[0] || !seen[1] {
		t.Errorf("replicate counts offered = %v, want 0 and 1", seen)
	}
	if seen[2] || seen[3] {
		t.Errorf("replicate counts offered = %v: two or more is past four energy", seen)
	}
}

// Inventor's Axe: equip pays two energy, and is refused with one.
func TestInventorsAxeEquipsForTwoEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	axe := apaPush(g, me.ID, me.ID, game.Card{Name: "Inventor's Axe", OracleID: inventorsAxeOracle, TypeLine: "Artifact — Equipment"})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}

	setEnergy(t, g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, axe, 0, params); !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("equip with one energy: err = %v, want ErrInsufficientEnergy", err)
	}
	setEnergy(t, g, me, 2)
	p7Activate(t, g, me, axe, 0, params)
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
	if c := findBattlefieldCardForTest(g, axe); c == nil || !c.IsAttachedTo(bear) {
		t.Error("the Axe is not attached to the bear")
	}
	if c := findBattlefieldCardForTest(g, bear); c == nil || c.Effective().Power != 4 {
		t.Errorf("equipped bear power = %v, want 4", c.Effective().Power)
	}
}

// Salvation Colossus: unearth pays eight energy from the graveyard, and
// is refused with seven.
func TestSalvationColossusUnearthsForEightEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	colossus := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: colossus, Name: "Salvation Colossus", OracleID: salvationColossusOracle,
			TypeLine: "Artifact Creature — Construct", Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID})
	})
	setEnergy(t, g, me, 7)
	if err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("unearth with seven energy: err = %v, want ErrInsufficientEnergy", err)
	}
	setEnergy(t, g, me, 8)
	p7Activate(t, g, me, colossus, 0, game.ActivateAbilityParams{})
	if energyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", energyOf(me))
	}
	if !g.Battlefield.Contains(colossus) {
		t.Error("Salvation Colossus did not return to the battlefield")
	}
}
