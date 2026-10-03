package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0109_pr7_cards_test.go — the cards ADR 0109 §8 and §9 unblock
// (#1862, #1842): the activated abilities that read the card discarded
// to pay for them, and the targets bounded by an X or by the counters
// removed. The world permanents are in world_permanents_test.go.

const (
	p7LandsEdge    = "82d58f2d-66a4-4154-9826-01ca8e8d32d0"
	p7Hisoka       = "b69701d6-8e46-4184-991b-9f5e95a5e16d"
	p7Chemister    = "4243baf6-9ac2-4f87-988d-3ebf5e177f3b"
	p7MoriaScav    = "a598e824-ad8e-4947-896b-70c5275b615a"
	p7Stockpile    = "dc47a97b-f511-4b97-97b2-e0309055544b"
	p7Tora         = "9317e70d-3d01-4652-b9ea-b4015a0f7b80"
	p7Volrath      = "4e11c838-68a7-4c97-b064-a46833cf1d2f"
	p7Manipulator  = "8f06fcc9-9018-4c55-af63-c44350a6cfeb"
	p7Quillmane    = "736168bb-0a5c-4b29-9f82-1319cd112873"
	p7KillingGlare = "7b47891f-130a-4d76-b262-53f5e3bb6b95"
	p7Sightbender  = "c4f1d8c8-29e5-43fe-8281-7d42d0173dc5"
	p7Aryel        = "9956f9b7-0140-484c-b606-3685690b84cc"
	p7Finale       = "3113afec-d4ae-46b5-9952-89e2e2c9ae7b"
	p7Lyre         = "9f70b907-586f-4c5d-bb7f-8aadf640ada9"
	p7Technomancer = "4e58ad76-37c7-4531-b207-6890b39a2679"
)

// p7Hand puts a card with a printed cost into p's hand.
func p7Hand(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

// p7Table is a catalog game in the active seat's main phase with empty
// hands and graveyards for the two seats a test uses.
func p7Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g, me, opp := exileCostTable(t)
	g.WithWriteLock(func() {
		opp.Hand.Cards = nil
		opp.Graveyard.Cards = nil
	})
	return g, me, opp
}

// p7Activate activates row `index` of `source` for p, paid on paper,
// and resolves it.
func p7Activate(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(p.ID, source, index, params); err != nil {
		t.Fatalf("%s activates row %d: %v", p.Name, index, err)
	}
	passPriorityAroundTable(t, g)
}

func p7Countered(g *game.Game, spell uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && ev.CardID == spell {
			return true
		}
	}
	return false
}

// --- §8: the card discarded to pay -----------------------------------

// Land's Edge: any player may activate it, from their own hand, and
// only a land card deals the damage.
func TestLandsEdgeDealsTwoOnlyForALandAndAnyPlayerMayActivate(t *testing.T) {
	g, me, opp := p7Table(t)
	edge := apaPush(g, me.ID, me.ID, game.Card{Name: "Land's Edge", TypeLine: "World Enchantment", OracleID: p7LandsEdge})
	land := p7Hand(opp, "Opp Land", "Land", "")
	spell := p7Hand(opp, "Opp Spell", "Sorcery", "{1}{R}")
	life := me.Life
	p7Activate(t, g, opp, edge, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{spell},
		Targets:    []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if me.Life != life {
		t.Fatalf("a nonland discard dealt damage: life %d -> %d", life, me.Life)
	}
	p7Activate(t, g, opp, edge, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{land},
		Targets:    []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	})
	if me.Life != life-2 {
		t.Errorf("a land discard: life %d, want %d", me.Life, life-2)
	}
	if !opp.Graveyard.Contains(land) || !opp.Graveyard.Contains(spell) {
		t.Error("the discards come from the ACTIVATOR's hand")
	}
}

// Hisoka counters only a spell whose mana value matches the discarded
// card's.
func TestHisokaCountersOnlyAMatchingManaValue(t *testing.T) {
	for _, tc := range []struct {
		name    string
		discard string
		counter bool
	}{
		{"matching", "{1}{U}", true},
		{"different", "{2}{U}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := p7Table(t)
			hisoka := apaPush(g, me.ID, me.ID, apaCreature("Hisoka, Minamo Sensei", p7Hisoka, 1, 3))
			paid := p7Hand(me, "Paid Card", "Creature — Test", tc.discard)
			spell := p7Hand(me, "Two Drop", "Sorcery", "{1}{R}")
			if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			p7Activate(t, g, me, hisoka, 0, game.ActivateAbilityParams{
				DiscardIDs: []uuid.UUID{paid}, Targets: cardRefs(spell),
			})
			if got := p7Countered(g, spell); got != tc.counter {
				t.Errorf("countered = %v, want %v", got, tc.counter)
			}
		})
	}
}

// Mercurial Chemister deals damage equal to the discarded card's mana
// value.
func TestMercurialChemisterDealsTheDiscardedManaValue(t *testing.T) {
	g, me, opp := p7Table(t)
	chem := apaPush(g, me.ID, me.ID, apaCreature("Mercurial Chemister", p7Chemister, 2, 3))
	victim := apaPush(g, opp.ID, opp.ID, apaCreature("Big Victim", "", 1, 5))
	paid := p7Hand(me, "Four Drop", "Creature — Test", "{3}{U}")
	p7Activate(t, g, me, chem, 1, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{paid}, Targets: cardRefs(victim),
	})
	if got := findBattlefieldCardForTest(g, victim).DamageMarked; got != 4 {
		t.Errorf("damage marked %d, want 4", got)
	}
}

// Moria Scavenger amasses only for a creature card; Necromancer's
// Stockpile makes its tapped Zombie only for a Zombie card.
func TestMoriaScavengerAndStockpileReadTheDiscardedCard(t *testing.T) {
	g, me, _ := p7Table(t)
	scav := apaPush(g, me.ID, me.ID, apaCreature("Moria Scavenger", p7MoriaScav, 1, 4))
	stock := apaPush(g, me.ID, me.ID, game.Card{Name: "Necromancer's Stockpile", TypeLine: "Enchantment", OracleID: p7Stockpile})
	land := p7Hand(me, "A Land", "Land", "")
	p7Activate(t, g, me, scav, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}})
	if n := countTokensControlled(g, me.ID, "Army"); n != 0 {
		t.Fatalf("a land discard amassed: %d Armies", n)
	}
	bear := p7Hand(me, "A Bear", "Creature — Bear", "{1}{G}")
	findBattlefieldCardForTest(g, scav).Tapped = false
	p7Activate(t, g, me, scav, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{bear}})
	if n := countTokensControlled(g, me.ID, "Army"); n != 1 {
		t.Errorf("a creature discard: %d Armies, want 1", n)
	}

	elf := p7Hand(me, "An Elf", "Creature — Elf", "{G}")
	p7Activate(t, g, me, stock, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{elf}})
	if n, _ := countTokensNamed(g, me.ID, "Zombie"); n != 0 {
		t.Fatalf("a non-Zombie discard made %d Zombies", n)
	}
	zombie := p7Hand(me, "A Zombie", "Creature — Zombie", "{B}")
	p7Activate(t, g, me, stock, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{zombie}})
	if n, tapped := countTokensNamed(g, me.ID, "Zombie"); n != 1 || tapped != 1 {
		t.Errorf("a Zombie discard: %d Zombies, %d tapped, want one tapped", n, tapped)
	}
	land2 := p7Hand(me, "Another Land", "Land", "")
	if err := g.ActivateCatalogAbility(me.ID, stock, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land2}}); err == nil {
		t.Error("the Stockpile took a land for \"Discard a creature card\"")
	}
}

// Slumbering Tora becomes an X/X Cat artifact creature; Volrath gets
// +X/+X — X the discarded card's mana value.
func TestToraAndVolrathSizeByTheDiscardedManaValue(t *testing.T) {
	g, me, _ := p7Table(t)
	tora := apaPush(g, me.ID, me.ID, game.Card{Name: "Slumbering Tora", TypeLine: "Artifact", OracleID: p7Tora})
	volrath := apaPush(g, me.ID, me.ID, apaCreature("Volrath the Fallen", p7Volrath, 6, 4))
	nonSpirit := p7Hand(me, "Plain Bear", "Creature — Bear", "{1}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tora, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{nonSpirit}}); err == nil {
		t.Fatal("Slumbering Tora took a card that is neither Spirit nor Arcane")
	}
	arcane := p7Hand(me, "Arcane Trick", "Instant — Arcane", "{2}{U}")
	p7Activate(t, g, me, tora, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{arcane}})
	tc := apaLive(g, tora)
	if !tc.IsCreature() || !tc.HasSubtype("Cat") || tc.CurrentPower() != 3 || tc.CurrentToughness() != 3 {
		t.Errorf("Tora is %v Cat=%v %d/%d, want a 3/3 Cat creature", tc.IsCreature(), tc.HasSubtype("Cat"), tc.CurrentPower(), tc.CurrentToughness())
	}

	big := p7Hand(me, "Big Creature", "Creature — Test", "{4}{B}")
	p7Activate(t, g, me, volrath, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{big}})
	if v := apaLive(g, volrath); v.CurrentPower() != 11 || v.CurrentToughness() != 9 {
		t.Errorf("Volrath is %d/%d, want 11/9", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- §9: a target bounded by X or by the counters removed --------------

// Simic Manipulator: the steal is bounded by the counters removed.
func TestSimicManipulatorStealsWithinTheCountersRemoved(t *testing.T) {
	g, me, opp := p7Table(t)
	man := apaPush(g, me.ID, me.ID, apaCreature("Simic Manipulator", p7Manipulator, 0, 1))
	pfAddCounters(g, man, game.CounterPlusOne, 2)
	two := apaPush(g, opp.ID, opp.ID, apaCreature("Two Power", "", 2, 2))
	three := apaPush(g, opp.ID, opp.ID, apaCreature("Three Power", "", 3, 3))
	if err := g.ActivateCatalogAbility(me.ID, man, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: cardRefs(three),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two counters for a power-3 creature: err = %v, want ErrIllegalTarget", err)
	}
	p7Activate(t, g, me, man, 0, game.ActivateAbilityParams{CounterCounts: []int{2}, Targets: cardRefs(two)})
	if pfController(g, two) != me.ID {
		t.Error("the power-2 creature was not taken")
	}
}

// Quillmane Baku: a ki counter per Spirit or Arcane spell, and a bounce
// bounded by the ki counters removed.
func TestQuillmaneBakuBouncesWithinTheKiCountersRemoved(t *testing.T) {
	g, me, opp := p7Table(t)
	baku := apaPush(g, me.ID, me.ID, apaCreature("Quillmane Baku", p7Quillmane, 3, 3))
	pfAddCounters(g, baku, "ki", 2)
	cheap := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Cheap", TypeLine: "Creature — Test", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	pricey := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Pricey", TypeLine: "Creature — Test", ManaCost: "{4}{G}", Power: 5, Toughness: 5})
	if err := g.ActivateCatalogAbility(me.ID, baku, 0, game.ActivateAbilityParams{
		CounterCounts: []int{2}, Targets: cardRefs(pricey),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two ki counters for a mana-value-5 creature: err = %v, want ErrIllegalTarget", err)
	}
	p7Activate(t, g, me, baku, 0, game.ActivateAbilityParams{CounterCounts: []int{2}, Targets: cardRefs(cheap)})
	if !opp.Hand.Contains(cheap) {
		t.Error("the mana-value-2 creature was not returned")
	}
}

// Killing Glare and Finale of Eternity: power and toughness bounded by
// the spell's X; Finale returns the creature cards at X of 10 or more.
func TestKillingGlareAndFinaleBoundByX(t *testing.T) {
	g, me, opp := p7Table(t)
	small := apaPush(g, opp.ID, opp.ID, apaCreature("Small", "", 2, 2))
	big := apaPush(g, opp.ID, opp.ID, apaCreature("Big", "", 5, 1))
	glare := p7Hand(me, "Killing Glare", "Instant", "{X}{B}")
	setOracle(me, glare, p7KillingGlare)
	if err := g.CastSpell(me.ID, glare, game.CastSpellParams{XValue: 3, Targets: cardRefs(big)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("X=3 for a power-5 creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.CastSpell(me.ID, glare, game.CastSpellParams{XValue: 3, Targets: cardRefs(small)}); err != nil {
		t.Fatalf("cast Killing Glare: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(small) {
		t.Fatal("Killing Glare did not destroy the power-2 creature")
	}

	// Finale: toughness bound — the 5/1 has toughness 1.
	mine := apaPush(g, me.ID, me.ID, apaCreature("My Dork", "", 1, 1))
	finale := p7Hand(me, "Finale of Eternity", "Sorcery", "{X}{B}{B}")
	setOracle(me, finale, p7Finale)
	if err := g.CastSpell(me.ID, finale, game.CastSpellParams{XValue: 10, Targets: cardRefs(big, mine)}); err != nil {
		t.Fatalf("cast Finale: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(big) {
		t.Error("Finale did not destroy the toughness-1 creature")
	}
	if findBattlefieldCardForTest(g, mine) == nil {
		t.Error("X=10: my destroyed creature card did not come back from the graveyard")
	}
}

// setOracle stamps an oracle ID on a hand card.
func setOracle(p *game.Player, id uuid.UUID, oracle string) {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].OracleID = oracle
		}
	}
}

// Minamo Sightbender and Entrancing Lyre: power bounded by an {X} in
// the activation cost.
func TestSightbenderAndLyreBoundPowerByX(t *testing.T) {
	g, me, opp := p7Table(t)
	sight := apaPush(g, me.ID, me.ID, apaCreature("Minamo Sightbender", p7Sightbender, 1, 2))
	lyre := apaPush(g, me.ID, me.ID, game.Card{Name: "Entrancing Lyre", TypeLine: "Artifact", OracleID: p7Lyre})
	small := apaPush(g, opp.ID, opp.ID, apaCreature("Small", "", 2, 2))
	big := apaPush(g, opp.ID, opp.ID, apaCreature("Big", "", 4, 4))
	if err := g.ActivateCatalogAbility(me.ID, sight, 0, game.ActivateAbilityParams{XValue: 2, Targets: cardRefs(big)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Sightbender X=2 on power 4: err = %v, want ErrIllegalTarget", err)
	}
	p7Activate(t, g, me, sight, 0, game.ActivateAbilityParams{XValue: 2, Targets: cardRefs(small)})
	if err := g.ActivateCatalogAbility(me.ID, lyre, 0, game.ActivateAbilityParams{XValue: 3, Targets: cardRefs(big)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Lyre X=3 on power 4: err = %v, want ErrIllegalTarget", err)
	}
	p7Activate(t, g, me, lyre, 0, game.ActivateAbilityParams{XValue: 4, Targets: cardRefs(big)})
	if !findBattlefieldCardForTest(g, big).Tapped {
		t.Error("the Lyre did not tap the power-4 creature at X=4")
	}
}

// Aryel's X is the Knights her cost taps; Ruthless Technomancer's is
// the artifacts his cost sacrifices, and it can't be 0.
func TestAryelAndTechnomancerBoundByTheirCostCounts(t *testing.T) {
	g, me, opp := p7Table(t)
	aryel := apaPush(g, me.ID, me.ID, apaCreature("Aryel, Knight of Windgrace", p7Aryel, 4, 4))
	k1 := apaPush(g, me.ID, me.ID, game.Card{Name: "Knight A", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	k2 := apaPush(g, me.ID, me.ID, game.Card{Name: "Knight B", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	three := apaPush(g, opp.ID, opp.ID, apaCreature("Three", "", 3, 3))
	two := apaPush(g, opp.ID, opp.ID, apaCreature("Two", "", 2, 2))
	if err := g.ActivateCatalogAbility(me.ID, aryel, 1, game.ActivateAbilityParams{
		XValue: 2, TapIDs: []uuid.UUID{k1, k2}, Targets: cardRefs(three),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("two Knights for a power-3 creature: err = %v, want ErrIllegalTarget", err)
	}
	p7Activate(t, g, me, aryel, 1, game.ActivateAbilityParams{XValue: 2, TapIDs: []uuid.UUID{k1, k2}, Targets: cardRefs(two)})
	if !opp.Graveyard.Contains(two) {
		t.Error("Aryel did not destroy the power-2 creature")
	}

	tech := apaPush(g, me.ID, me.ID, apaCreature("Ruthless Technomancer", p7Technomancer, 2, 4))
	art := apaPush(g, me.ID, me.ID, game.Card{Name: "Trinket", TypeLine: "Artifact"})
	dead := pushGraveyardCardTyped(me, "Dead Three", "Creature — Test")
	deadSmall := pushGraveyardCardTyped(me, "Dead One", "Creature — Test")
	for i := range me.Graveyard.Cards {
		switch me.Graveyard.Cards[i].InstanceID {
		case dead:
			me.Graveyard.Cards[i].Power = 3
		case deadSmall:
			me.Graveyard.Cards[i].Power = 1
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, tech, 0, game.ActivateAbilityParams{
		XValue: 0, Targets: cardRefs(deadSmall),
	}); err == nil {
		t.Fatal("Technomancer activated with X=0")
	}
	if err := g.ActivateCatalogAbility(me.ID, tech, 0, game.ActivateAbilityParams{
		XValue: 1, SacrificeIDs: []uuid.UUID{art}, Targets: cardRefs(dead),
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("one artifact for a power-3 card: err = %v, want ErrIllegalTarget", err)
	}
	p7Activate(t, g, me, tech, 0, game.ActivateAbilityParams{XValue: 1, SacrificeIDs: []uuid.UUID{art}, Targets: cardRefs(deadSmall)})
	if findBattlefieldCardForTest(g, deadSmall) == nil {
		t.Error("the power-1 creature card did not return")
	}
}
