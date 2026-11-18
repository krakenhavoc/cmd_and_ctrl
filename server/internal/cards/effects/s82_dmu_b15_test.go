package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 DMU batch 15 — multicolored and "all colors" (#5487).

const (
	s82DmuWayfarer  = "01de1f9d-3b63-49b9-b2c6-929c2a69ef54"
	s82DmuMaelstrom = "92cc43e5-782f-4540-bc60-3d18f357caf3"
	s82DmuJared     = "f9485c76-122d-4faf-9194-b84ed9c41378"
	s82DmuJenson    = "68cefb44-afad-4638-8fdc-65a099b51d55"
	s82DmuCannons   = "26d28b68-e03b-4367-ae4b-5e54f42f8a88"
	s82DmuManaRig   = "58678469-77e5-4c1b-88d6-5e6642b0c8ca"
	s82DmuHellkite  = "27596982-f4ab-488f-a974-5b0404438504"
	s82DmuUnite     = "164651c5-b831-4070-8491-16d30ac8395b"
)

var s82DmuAllColors = []string{"W", "U", "B", "R", "G"}

func TestS82DmuB15CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82DmuWayfarer: "Fallaji Wayfarer", s82DmuMaelstrom: "Iridian Maelstrom",
		s82DmuJared: "Jared Carthalion", s82DmuJenson: "Jenson Carthalion, Druid Exile",
		s82DmuCannons: "Mana Cannons", s82DmuManaRig: "The Mana Rig",
		s82DmuHellkite: "Two-Headed Hellkite", s82DmuUnite: "Unite the Coalition",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82DmuCast casts a spell with the given cost from the active seat's hand,
// paid for with the floating pool.
func s82DmuCast(t *testing.T, g *game.Game, cost, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := socB17Hand(me, "Test Spell", "Sorcery", cost, socB17PlainSorceryOracle)
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", cost, err)
	}
	return id
}

// --- Fallaji Wayfarer -------------------------------------------------

func TestS82DmuWayfarerIsAllColorsAndConvokesOnlyMulticoloredSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	w := b27Push(g, me.ID, "Fallaji Wayfarer", "Creature — Human Scout", s82DmuWayfarer, "{2}{G}", 2, 4, "G")
	if got := len(layeredCard(t, g, w).EffectiveColors()); got != 5 {
		t.Errorf("the Wayfarer has %d colors, want all five", got)
	}
	elves := cgTwoGreenCreatures(g, me.ID)
	elf := elves[:1]
	gold := socB17Hand(me, "Gold Spell", "Sorcery", "{G}{G}", socB17PlainSorceryOracle)
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a monocolored-cost pair of creatures paid for a {G}{G} spell with convoke; it is not multicolored")
	}
	multi := socB17Hand(me, "Multi Spell", "Sorcery", "{G}{W}", socB17PlainSorceryOracle)
	// The green elf pays {G}; the all-colors Wayfarer pays {W}.
	tap := append([]uuid.UUID{w}, elf...)
	if err := g.CastSpell(me.ID, multi, game.CastSpellParams{Strict: true, TapIDs: tap}); err != nil {
		t.Fatalf("convoking a multicolored spell: %v", err)
	}
	if !cgTapped(g, tap) {
		t.Error("the convoking creatures were not tapped")
	}
}

func TestS82DmuWayfarerGivesNoConvokeWithoutIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	elves := cgTwoGreenCreatures(g, me.ID)
	multi := socB17Hand(me, "Multi Spell", "Sorcery", "{G}{W}", socB17PlainSorceryOracle)
	if err := g.CastSpell(me.ID, multi, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Error("a spell was convoked with no Wayfarer")
	}
}

// --- Iridian Maelstrom ------------------------------------------------

func TestS82DmuMaelstromDestroysEveryCreatureThatIsNotAllColors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b27Push(g, me.ID, "My Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	theirs := b27Push(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", "{3}", 3, 3)
	four := b27Push(g, opp.ID, "Four Colors", "Creature — Dragon", "", "{W}{U}{B}{R}", 4, 4, "W", "U", "B", "R")
	rainbow := b27Push(g, opp.ID, "Rainbow", "Creature — Dragon", "", "{W}{U}{B}{R}{G}", 5, 5, s82DmuAllColors...)
	rock := b27Push(g, opp.ID, "Rock", "Artifact", "", "{2}", 0, 0)
	castXSpell(t, g, "Iridian Maelstrom", "Sorcery", s82DmuMaelstrom, "{W}{U}{B}{R}{G}", 0, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, four} {
		if findBattlefieldCardByID(g, id) != nil {
			t.Errorf("a creature that is not all colors survived")
		}
	}
	if findBattlefieldCardByID(g, rainbow) == nil {
		t.Error("the all-colors creature was destroyed")
	}
	if findBattlefieldCardByID(g, rock) == nil {
		t.Error("a noncreature artifact was destroyed")
	}
}

// --- Mana Cannons -----------------------------------------------------

func TestS82DmuManaCannonsDealsDamageEqualToTheSpellsColors(t *testing.T) {
	for _, tc := range []struct {
		name, cost, pool string
		want             int
	}{
		{"two colors", "{W}{B}", "WB", 2},
		{"three colors", "{W}{U}{B}", "WUB", 3},
		{"one color", "{1}{R}", "RR", 0},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		b27Push(g, me.ID, "Mana Cannons", "Enchantment", s82DmuCannons, "{2}{R}", 0, 0, "R")
		before := opp.Life
		s82DmuCast(t, g, tc.cost, tc.pool)
		if tc.want > 0 {
			p := latestPickTarget(g, me.ID)
			if p == nil {
				t.Fatalf("%s: no target prompt", tc.name)
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, playerTarget(opp.ID)); err != nil {
				t.Fatalf("%s: target: %v", tc.name, err)
			}
		}
		passPriorityAroundTable(t, g)
		if got := before - opp.Life; got != tc.want {
			t.Errorf("%s: %d damage, want %d", tc.name, got, tc.want)
		}
	}
}

// --- The Mana Rig -----------------------------------------------------

func TestS82DmuManaRigMakesATappedPowerstoneForAMulticoloredSpell(t *testing.T) {
	for _, tc := range []struct {
		cost, pool string
		want       int
	}{{"{W}{B}", "WB", 1}, {"{1}{R}", "RR", 0}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b27Push(g, me.ID, "The Mana Rig", "Legendary Artifact", s82DmuManaRig, "{3}", 0, 0)
		s82DmuCast(t, g, tc.cost, tc.pool)
		passPriorityAroundTable(t, g)
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Powerstone" && c.Controller == me.ID {
				n++
				if !c.Tapped {
					t.Error("the Powerstone entered untapped")
				}
			}
		}
		if n != tc.want {
			t.Errorf("%s: %d Powerstones, want %d", tc.cost, n, tc.want)
		}
	}
}

func TestS82DmuManaRigLooksAtXAndTakesUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rig := b27Push(g, me.ID, "The Mana Rig", "Legendary Artifact", s82DmuManaRig, "{3}", 0, 0)
	for i := 0; i < 6; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	fillPool(me, 12)
	if err := g.ActivateCatalogAbility(me.ID, rig, 0, game.ActivateAbilityParams{XValue: 4, Strict: true}); err != nil {
		t.Fatalf("activate with X=4 and twelve mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestChooseCardsFor(g, me.ID)
	if p == nil {
		t.Fatal("no choice of cards for the Rig")
	}
	if len(p.ChooseCards) != 4 {
		t.Errorf("the Rig offered %d cards, want the top X = 4", len(p.ChooseCards))
	}
	if p.ChooseMin != 0 || p.ChooseMax != 2 {
		t.Errorf("bounds = %d..%d, want up to two", p.ChooseMin, p.ChooseMax)
	}
	handBefore := me.Hand.Size()
	if err := g.ResolveChooseCards(p.ID, me.ID, p.ChooseCards[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("took %d cards, want 2", got)
	}
}

func TestS82DmuManaRigRefusesAnXItCannotPay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rig := b27Push(g, me.ID, "The Mana Rig", "Legendary Artifact", s82DmuManaRig, "{3}", 0, 0)
	fillPool(me, 8)
	if err := g.ActivateCatalogAbility(me.ID, rig, 0, game.ActivateAbilityParams{XValue: 3, Strict: true}); err == nil {
		t.Error("X=3 costs nine, and only eight mana was available")
	}
}

// --- Two-Headed Hellkite ----------------------------------------------

func TestS82DmuHellkiteDrawsTwoWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := b27Push(g, me.ID, "Two-Headed Hellkite", "Creature — Dragon", s82DmuHellkite, "{1}{W}{U}{B}{R}{G}", 5, 5, s82DmuAllColors...)
	assertKeywords(t, g, h, "flying", "menace", "haste")
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	before := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(h, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// --- Jenson Carthalion, Druid Exile -----------------------------------

func TestS82DmuJensonScriesAndMakesAnAngelOnlyForAnAllColorsSpell(t *testing.T) {
	for _, tc := range []struct {
		name, cost, pool string
		angel            int
	}{
		{"two colors", "{W}{B}", "WB", 0},
		{"all colors", "{W}{U}{B}{R}{G}", "WUBRG", 1},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b27Push(g, me.ID, "Jenson Carthalion, Druid Exile", "Legendary Creature — Human Druid", s82DmuJenson, "{G}{W}", 2, 2, "G", "W")
		for i := 0; i < 3; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		}
		s82DmuCast(t, g, tc.cost, tc.pool)
		passPriorityAroundTable(t, g)
		if got := countBattlefieldByName(g, "Angel"); got != tc.angel {
			t.Errorf("%s: %d Angels, want %d", tc.name, got, tc.angel)
		}
		if p := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); p == nil {
			t.Errorf("%s: no scry prompt", tc.name)
		}
	}
}

func TestS82DmuJensonMakesFiveColorsOfManaForFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	j := b27Push(g, me.ID, "Jenson Carthalion, Druid Exile", "Legendary Creature — Human Druid", s82DmuJenson, "{G}{W}", 2, 2, "G", "W")
	fillPool(me, 5)
	if err := g.ActivateManaAbility(me.ID, j, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, col := range s82DmuAllColors {
		if me.ManaPool.CountColor(col) < 1 {
			t.Errorf("no %s mana in the pool", col)
		}
	}
}

// --- Jared Carthalion -------------------------------------------------

func s82DmuJaredOnTheBattlefield(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return pushCatalogWalkerWithLine(g, owner, "Jared Carthalion", "Legendary Planeswalker — Jared", s82DmuJared, loyalty)
}

func TestS82DmuJaredPlusOneMakesAnAllColorsKavu(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	j := s82DmuJaredOnTheBattlefield(g, me.ID, 5)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, j, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	var kavu *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Kavu" {
			kavu = &g.Battlefield.Cards[i]
		}
	}
	if kavu == nil {
		t.Fatal("no Kavu token")
	}
	if !kavu.IsToken() || kavu.Power != 3 || kavu.Toughness != 3 || len(kavu.EffectiveColors()) != 5 {
		t.Errorf("Kavu: token %v, %d/%d, colors %v", kavu.IsToken(), kavu.Power, kavu.Toughness, kavu.EffectiveColors())
	}
	assertKeywords(t, g, kavu.InstanceID, "trample")
	if got := loyaltyCount(g, j); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82DmuJaredMinusThreePutsCountersEqualToEachCreaturesColors(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	j := s82DmuJaredOnTheBattlefield(g, me.ID, 5)
	a := b27Push(g, me.ID, "Gold", "Creature — Elf", "", "{G}{W}", 2, 2, "G", "W")
	b := b27Push(g, opp.ID, "Rainbow", "Creature — Dragon", "", "{W}{U}{B}{R}{G}", 5, 5, s82DmuAllColors...)
	c := b27Push(g, opp.ID, "Golem", "Artifact Creature — Golem", "", "{3}", 3, 3)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, j, 1, game.ActivateAbilityParams{Targets: cardTargets(a, b)}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("two-color creature has %d counters, want 2", got)
	}
	if got := counterCount(g, b, game.CounterPlusOne); got != 5 {
		t.Errorf("all-colors creature has %d counters, want 5", got)
	}
	if got := counterCount(g, c, game.CounterPlusOne); got != 0 {
		t.Errorf("untargeted creature has %d counters", got)
	}
	if got := loyaltyCount(g, j); got != 2 {
		t.Errorf("loyalty %d, want 2", got)
	}
}

func TestS82DmuJaredMinusThreeGivesAColorlessCreatureNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	j := s82DmuJaredOnTheBattlefield(g, me.ID, 5)
	c := b27Push(g, opp.ID, "Golem", "Artifact Creature — Golem", "", "{3}", 3, 3)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, j, 1, game.ActivateAbilityParams{Targets: cardTargets(c)}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, c, game.CounterPlusOne); got != 0 {
		t.Errorf("colorless creature has %d counters, want 0", got)
	}
}

func TestS82DmuJaredMinusSixReturnsAMulticoloredCardAndRewardsAnAllColorsOne(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		colors     []string
		drawn      int
		treasures  int
	}{
		{"two colors", "{G}{W}", []string{"G", "W"}, 0, 0},
		{"all colors", "{W}{U}{B}{R}{G}", s82DmuAllColors, 1, 2},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		j := s82DmuJaredOnTheBattlefield(g, me.ID, 6)
		card := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Gold Card", TypeLine: "Creature — Elf", ManaCost: tc.cost,
			Colors: tc.colors, Owner: me.ID, Controller: me.ID})
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		advanceToMain(t, g)
		handBefore := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, j, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}}}); err != nil {
			t.Fatalf("%s: -6: %v", tc.name, err)
		}
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(card) {
			t.Errorf("%s: the card did not return to hand", tc.name)
		}
		if got := me.Hand.Size() - handBefore; got != 1+tc.drawn {
			t.Errorf("%s: hand grew by %d, want %d", tc.name, got, 1+tc.drawn)
		}
		if got := countBattlefieldByName(g, "Treasure"); got != tc.treasures {
			t.Errorf("%s: %d Treasures, want %d", tc.name, got, tc.treasures)
		}
	}
}

func TestS82DmuJaredMinusSixRefusesAMonocoloredCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	j := s82DmuJaredOnTheBattlefield(g, me.ID, 6)
	card := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Mono Card", TypeLine: "Creature — Elf", ManaCost: "{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, j, 2, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}}}); err == nil {
		t.Error("a monocolored card was a legal target")
	}
}

// --- Unite the Coalition ----------------------------------------------

func TestS82DmuUniteTheCoalitionRunsFiveModesAndRepeatsOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	rock := b27Push(g, opp.ID, "Their Rock", "Artifact", "", "{2}", 0, 0)
	ench := b27Push(g, opp.ID, "Their Aura", "Enchantment", "", "{1}{W}", 0, 0)
	old := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	})
	handBefore, lifeBefore := me.Hand.Size(), opp.Life
	castModal(t, g, "Unite the Coalition", "Instant", s82DmuUnite,
		[]int{1, 1, 2, 3, 4},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, me.ID, 0, 0),
			modeRef(game.TargetPlayer, me.ID, 1, 0),
			modeRef(game.TargetPlayer, opp.ID, 2, 0),
			modeRef(game.TargetPlayer, opp.ID, 3, 0),
			modeRef(game.TargetCard, rock, 4, 0),
		})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("two draw bullets drew %d cards, want 2", got)
	}
	if opp.Graveyard.Contains(old) {
		t.Error("the opponent's graveyard was not exiled")
	}
	if got := lifeBefore - opp.Life; got != 2 {
		t.Errorf("%d damage, want 2", got)
	}
	if findBattlefieldCardByID(g, rock) != nil {
		t.Error("the artifact was not destroyed")
	}
	if findBattlefieldCardByID(g, ench) == nil {
		t.Error("an untargeted enchantment was destroyed")
	}
}

func TestS82DmuUniteTheCoalitionPhasesOutAndDestroysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b27Push(g, opp.ID, "Their Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	ench := b27Push(g, opp.ID, "Their Aura", "Enchantment", "", "{1}{W}", 0, 0)
	castModal(t, g, "Unite the Coalition", "Instant", s82DmuUnite,
		[]int{0, 4, 4, 4, 4},
		[]game.TargetRef{
			modeRef(game.TargetCard, bear, 0, 0),
			modeRef(game.TargetCard, ench, 1, 0),
			modeRef(game.TargetCard, ench, 2, 0),
			modeRef(game.TargetCard, ench, 3, 0),
			modeRef(game.TargetCard, ench, 4, 0),
		})
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, ench) != nil {
		t.Error("the enchantment was not destroyed")
	}
	if !phasedOutInCatalogGame(g, bear) {
		t.Error("the Bear did not phase out")
	}
}

func TestS82DmuUniteTheCoalitionRefusesTheWrongCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Unite the Coalition", TypeLine: "Instant",
		OracleID: s82DmuUnite, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1, 1, 1, 1}, Strict: true})
	if err == nil {
		t.Error("four modes were accepted; the card says choose five")
	}
}
