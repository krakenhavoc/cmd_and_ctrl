package testsds

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsa "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/a"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestEnergyAltCostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{effectstest.NissaWorldsoulOracle, effectstest.PrimalPrayersOracle, effectstest.AmpedRaptorOracle,
		effectstest.ReiteratingBoltOracle, effectstest.InventorsAxeOracle, effectstest.SalvationColossusOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
}

// Nissa: eight energy pays for a permanent spell from hand under the
// strict gate with no mana at all; seven is not offered and is refused;
// an instant is never offered.
func TestNissaWorldsoulSpeakerOffersEightEnergyForPermanentSpells(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Nissa, Worldsoul Speaker", OracleID: effectstest.NissaWorldsoulOracle,
		TypeLine: "Legendary Creature — Elf Druid", Power: 3, Toughness: 3})
	big := effectstest.EacHandCard(g, me, "Big Creature", "Creature — Test", "{6}{G}{G}")
	instant := effectstest.EacHandCard(g, me, "Some Instant", "Instant", "{1}{U}")

	effectstest.SetEnergy(t, g, me, 7)
	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, big, game.ZoneHand), effects.GrantedAltCostEightEnergyPermanents); o != nil {
		t.Fatalf("offered with seven energy: %+v", o)
	}
	err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostEightEnergyPermanents})
	if err == nil {
		t.Fatal("the claim was accepted with seven energy")
	}
	if effectstest.EnergyOf(me) != 7 || !me.Hand.Contains(big) {
		t.Fatalf("a refused claim paid something: energy %d, in hand %v", effectstest.EnergyOf(me), me.Hand.Contains(big))
	}

	effectstest.SetEnergy(t, g, me, 8)
	o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, big, game.ZoneHand), effects.GrantedAltCostEightEnergyPermanents)
	if o == nil || o.Energy != 8 || o.ManaCost != "" || !o.Granted {
		t.Fatalf("offer with eight energy = %+v", o)
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, instant, game.ZoneHand), effects.GrantedAltCostEightEnergyPermanents) != nil {
		t.Error("an instant was offered the permanent-spell price")
	}
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostEightEnergyPermanents}); err != nil {
		t.Fatalf("cast for eight energy: %v", err)
	}
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy = %d after the cast, want 0", effectstest.EnergyOf(me))
	}

	// ADR 0129 §4: Cast anyway waives only the mana. The eight energy is
	// still charged.
	effectstest.PassPriorityAroundTable(t, g)
	other := effectstest.EacHandCard(g, me, "Other Creature", "Creature — Test", "{6}{G}{G}")
	effectstest.SetEnergy(t, g, me, 8)
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true, ForceCast: true, AlternativeCost: effects.GrantedAltCostEightEnergyPermanents}); err != nil {
		t.Fatalf("Cast anyway for eight energy: %v", err)
	}
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("Cast anyway left energy %d, want 0", effectstest.EnergyOf(me))
	}
}

// Primal Prayers: creature spells with mana value 3 or less only, one
// energy, and as though they had flash only when cast this way (CR
// 601.3c): out of a main phase the printed cost is refused for timing
// and the claim is accepted.
func TestPrimalPrayersCastsSmallCreaturesForOneEnergyWithFlash(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Primal Prayers", OracleID: effectstest.PrimalPrayersOracle, TypeLine: "Enchantment"})
	three := effectstest.EacHandCard(g, me, "Three Drop", "Creature — Test", "{2}{G}")
	four := effectstest.EacHandCard(g, me, "Four Drop", "Creature — Test", "{3}{G}")
	sorcery := effectstest.EacHandCard(g, me, "Cheap Sorcery", "Sorcery", "{G}")
	effectstest.SetEnergy(t, g, me, 2)

	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, three, game.ZoneHand), effects.GrantedAltCostEnergySmallCreatures); o == nil || o.Energy != 1 || !o.AsThoughFlash {
		t.Fatalf("offer on a mana value 3 creature = %+v", o)
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, four, game.ZoneHand), effects.GrantedAltCostEnergySmallCreatures) != nil {
		t.Error("a mana value 4 creature was offered the price")
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, sorcery, game.ZoneHand), effects.GrantedAltCostEnergySmallCreatures) != nil {
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
		if p.AlternativeCost != effects.GrantedAltCostEnergySmallCreatures || m.Cost == nil || m.Cost.Energy != 1 {
			t.Errorf("move %q: claim %q cost %+v, want the energy offer paying 1", m.Label, p.AlternativeCost, m.Cost)
		}
		clone := g.Clone()
		if err := actions.Dispatch(clone, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}

	if err := g.CastSpell(me.ID, three, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostEnergySmallCreatures}); err != nil {
		t.Fatalf("flash cast for one energy: %v", err)
	}
	if effectstest.EnergyOf(me) != 1 {
		t.Errorf("energy = %d, want 1", effectstest.EnergyOf(me))
	}
}

// Amped Raptor: cast from hand, its trigger gives two energy and exiles
// down to a nonland card, which may be cast for energy equal to its mana
// value and for nothing else.
func TestAmpedRaptorCastsTheExiledCardForEnergy(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	hit := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	})
	effectstest.CastCatalogSpell(t, g, "Amped Raptor", "Creature — Dinosaur", effectstest.AmpedRaptorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.EnergyOf(me) != 2 {
		t.Fatalf("energy = %d after the trigger, want 2", effectstest.EnergyOf(me))
	}
	if !g.Exile.Contains(hit) || !g.Exile.Contains(land) {
		t.Fatalf("the land and the two-drop are not both in exile")
	}
	o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, hit, game.ZoneExile), cardsa.AmpedRaptorAltCostKey)
	if o == nil || o.Energy != 2 || o.ManaCost != "" {
		t.Fatalf("offer on the exiled two-drop = %+v", o)
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Fatal("the exiled card was cast without claiming the energy price")
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: cardsa.AmpedRaptorAltCostKey}); err != nil {
		t.Fatalf("cast for two energy: %v", err)
	}
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy = %d after the cast, want 0", effectstest.EnergyOf(me))
	}
}

// Amped Raptor that was not cast from hand gives its energy and exiles
// nothing.
func TestAmpedRaptorNotCastFromHandOnlyGivesEnergy(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	before := me.Library.Size()
	raptor := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Amped Raptor", OracleID: effectstest.AmpedRaptorOracle,
		TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 1})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: raptor, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d, want 2", effectstest.EnergyOf(me))
	}
	if me.Library.Size() != before {
		t.Errorf("a Raptor not cast from hand exiled cards")
	}
}

// Reiterating Bolt: replicate is paid three energy at a time, the cast
// is refused past the caster's energy with nothing paid, and each
// payment is one copy.
func TestReiteratingBoltReplicatesForThreeEnergyEach(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	a := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	b := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear B", TypeLine: "Creature — Bear", Power: 2, Toughness: 6})
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Reiterating Bolt", TypeLine: "Sorcery", ManaCost: "{1}{R}",
			OracleID: effectstest.ReiteratingBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}}

	effectstest.SetEnergy(t, g, me, 8)
	err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: targets, OptionalCosts: []int{0, 0, 0}})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("three replicates with eight energy: err = %v, want ErrInsufficientEnergy", err)
	}
	if effectstest.EnergyOf(me) != 8 || !me.Hand.Contains(bolt) {
		t.Fatalf("a refused cast paid something: energy %d, in hand %v", effectstest.EnergyOf(me), me.Hand.Contains(bolt))
	}

	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: targets, OptionalCosts: []int{0, 0}}); err != nil {
		t.Fatalf("two replicates with eight energy: %v", err)
	}
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy = %d after two replicates, want 2", effectstest.EnergyOf(me))
	}
	if it := effectstest.TriggerOnStack(g, bolt); it == nil || it.XValue != 2 {
		t.Fatalf("replicate trigger = %+v, want one counting two payments", it)
	}
	prompts := 0
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
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
	g, _, opp := effectstest.P7Table(t)
	a := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	bolt := effectstest.CastCatalogSpell(t, g, "Reiterating Bolt", "Sorcery", effectstest.ReiteratingBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: a}})
	if effectstest.TriggerOnStack(g, bolt) != nil {
		t.Fatal("a replicate trigger with no replicate paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("the bear survived 3 damage")
	}
}

// ADR 0129 §7: the enumerator never offers a replicate count the seat
// cannot pay for, and names the energy on the move.
func TestReiteratingBoltEnumeratesOnlyAffordableReplicates(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Bear A", TypeLine: "Creature — Bear", Power: 2, Toughness: 3})
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Reiterating Bolt", TypeLine: "Sorcery", ManaCost: "{1}{R}",
			OracleID: effectstest.ReiteratingBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.SetEnergy(t, g, me, 4)
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
	g, me, _ := effectstest.P7Table(t)
	axe := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Inventor's Axe", OracleID: effectstest.InventorsAxeOracle, TypeLine: "Artifact — Equipment"})
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}

	effectstest.SetEnergy(t, g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, axe, 0, params); !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("equip with one energy: err = %v, want ErrInsufficientEnergy", err)
	}
	effectstest.SetEnergy(t, g, me, 2)
	effectstest.P7Activate(t, g, me, axe, 0, params)
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", effectstest.EnergyOf(me))
	}
	if c := effectstest.FindBattlefieldCardForTest(g, axe); c == nil || !c.IsAttachedTo(bear) {
		t.Error("the Axe is not attached to the bear")
	}
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c == nil || c.Effective().Power != 4 {
		t.Errorf("equipped bear power = %v, want 4", c.Effective().Power)
	}
}

// Salvation Colossus: unearth pays eight energy from the graveyard, and
// is refused with seven.
func TestSalvationColossusUnearthsForEightEnergy(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	colossus := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: colossus, Name: "Salvation Colossus", OracleID: effectstest.SalvationColossusOracle,
			TypeLine: "Artifact Creature — Construct", Power: 9, Toughness: 9, Owner: me.ID, Controller: me.ID})
	})
	effectstest.SetEnergy(t, g, me, 7)
	if err := g.ActivateCatalogAbility(me.ID, colossus, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("unearth with seven energy: err = %v, want ErrInsufficientEnergy", err)
	}
	effectstest.SetEnergy(t, g, me, 8)
	effectstest.P7Activate(t, g, me, colossus, 0, game.ActivateAbilityParams{})
	if effectstest.EnergyOf(me) != 0 {
		t.Errorf("energy = %d, want 0", effectstest.EnergyOf(me))
	}
	if !g.Battlefield.Contains(colossus) {
		t.Error("Salvation Colossus did not return to the battlefield")
	}
}
