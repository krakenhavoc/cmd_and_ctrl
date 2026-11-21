package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b42_test.go — GAP batch 42, Assassin's Creed sorceries (#5948).
// Each test plays the real catalog entry.

const (
	gapb42PettyLarcenyOracle = "5513a221-dfff-4e02-8037-dcff3d0381d7"
	gapb42AmbushOracle       = "72075231-e3b3-46cc-b994-2d95b24cf904"
	gapb42ViewpointOracle    = "2e338e09-5758-4cc2-9143-21f050a28677"
	gapb42MonasteryOracle    = "62f4a2db-6d7c-487b-be37-22f06c1e779e"
	gapb42OverpoweringOracle = "e31235f0-1f25-4158-b46e-cacde2f58b7e"
	gapb42ReconstructOracle  = "97fad58a-8ee4-4f01-a47a-5ddc9ed845af"
	gapb42DistractOracle     = "5b3687e5-546f-4a70-aa5c-b7db3300a3f4"
	gapb42WhatMustOracle     = "bd0df0a8-c2ec-4d1f-9ada-cc91f7a20e1a"
)

func TestGapB42CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb42PettyLarcenyOracle: "Petty Larceny",
		gapb42AmbushOracle:       "Auditore Ambush",
		gapb42ViewpointOracle:    "Viewpoint Synchronization",
		gapb42MonasteryOracle:    "Monastery Raid",
		gapb42OverpoweringOracle: "Overpowering Attack",
		gapb42ReconstructOracle:  "Reconstruct History",
		gapb42DistractOracle:     "Distract the Guards",
		gapb42WhatMustOracle:     "What Must Be Done",
	} {
		def, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

// gapb42Freerun casts `name` (a card of `typeLine`, `cost`, oracle) by its
// freerunning cost once an Assassin has connected this turn.
func gapb42Freerun(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, colors []string, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	id := handCardFull(me, name, typeLine, cost, oracle, colors)
	floatForTest(g, me, mana)
	params.Strict = true
	params.AlternativeCost = "freerunning"
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("%s: freerunning was refused: %v", name, err)
	}
	return id
}

func TestGapB42DistractTheGuardsMakesThreeHumanRogues(t *testing.T) {
	g := newCatalogGame(t)
	gapb42Freerun(t, g, "Distract the Guards", "Sorcery", "{1}{W}{W}", gapb42DistractOracle, []string{"W"}, "CW", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Human Rogue" {
			n++
			if c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Rogue") || !c.HasSubtype("Human") {
				t.Errorf("Human Rogue token = %+v", c)
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Human Rogue tokens, want 3", n)
	}
}

func TestGapB42PettyLarcenyExilesTwoFaceDownAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushLibraryTyped(opp, "Top Card", "Instant")
	b := pushLibraryTyped(opp, "Second Card", "Creature — Bear")
	below := pushLibraryTyped(opp, "Bottom Card", "Land")
	_ = below
	// The top of the library is the last card pushed in the helper's order.
	gapb42Freerun(t, g, "Petty Larceny", "Sorcery", "{3}{B}", gapb42PettyLarcenyOracle, []string{"B"}, "CB",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range []uuid.UUID{a, b, below} {
		if exileHas(g, id) {
			exiled++
			if grantedPermissionOn(g, me.ID, id, game.ZoneExile) == nil {
				t.Errorf("the caster has no permission to play an exiled card")
			}
		}
	}
	if exiled != 2 {
		t.Errorf("%d cards exiled, want the top two", exiled)
	}
	treasure := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Treasure" && c.Controller == me.ID {
			treasure++
		}
	}
	if treasure != 1 {
		t.Errorf("%d Treasures, want 1", treasure)
	}
}

func TestGapB42ViewpointSynchronizationFetchesThreeBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		pushLibraryTyped(me, "Forest", "Basic Land — Forest")
	}
	acrb1ConnectThenPostcombat(t, g, "Creature — Human Assassin")
	spell := handCardFull(me, "Viewpoint Synchronization", "Sorcery", "{4}{G}", gapb42ViewpointOracle, []string{"G"})
	floatForTest(g, me, "CCG")
	hand := me.Hand.Size()
	lib := me.Library.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning"}); err != nil {
		t.Fatalf("freerunning {2}{G} was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	forests := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == me.ID {
			forests++
			if !c.Tapped {
				t.Errorf("a fetched land entered untapped")
			}
		}
	}
	if forests != 2 {
		t.Errorf("%d Forests on the battlefield, want 2", forests)
	}
	// The spell left the hand and one Forest came into it.
	if got := me.Hand.Size(); got != hand-1+1 {
		t.Errorf("hand %d -> %d, want the spell gone and one land in", hand, got)
	}
	if got := me.Library.Size(); got != lib-3 {
		t.Errorf("library %d -> %d, want 3 fewer", lib, got)
	}
}

func TestGapB42MonasteryRaidExilesTwoOrXWhenFreerun(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 6; i++ {
		pushLibraryTyped(me, "Card", "Creature — Bear")
	}
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Monastery Raid", "Sorcery", gapb42MonasteryOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib-2 {
		t.Errorf("printed cost: library %d -> %d, want 2 exiled", lib, got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	for i := 0; i < 6; i++ {
		pushLibraryTyped(me2, "Card", "Creature — Bear")
	}
	acrb1ConnectThenPostcombat(t, g2, "Creature — Human Assassin")
	lib2 := me2.Library.Size()
	spell := handCardFull(me2, "Monastery Raid", "Sorcery", "{2}{R}", gapb42MonasteryOracle, []string{"R"})
	floatForTest(g2, me2, "CCCR")
	if err := g2.CastSpell(me2.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: "freerunning", XValue: 3}); err != nil {
		t.Fatalf("freerunning {X}{R} with X=3 was refused: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if got := me2.Library.Size(); got != lib2-3 {
		t.Errorf("freerunning X=3: library %d -> %d, want 3 exiled", lib2, got)
	}
}

func TestGapB42AuditoreAmbushBouncesAndFetchesEzio(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	ezio := pushLibraryTyped(me, "Ezio, Blade of Vengeance", "Legendary Creature — Human Assassin")
	pushLibraryTyped(me, "Other Card", "Creature — Bear")
	d1658Cast(t, g, "Auditore Ambush", "Sorcery", gapb42AmbushOracle, game.CastSpellParams{
		Modes: []int{0, 1},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: ogre, Mode: 0},
			{Kind: game.TargetPlayer, ID: me.ID, Mode: 1},
		},
	})
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if !inHandOf(opp, ogre) {
		t.Error("the creature is not in its owner's hand")
	}
	if !inHandOf(me, ezio) {
		t.Error("Ezio did not come into the target player's hand")
	}
}

func TestGapB42ReconstructHistoryReturnsOnePerTypeAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	art := seedGraveyardCard(t, g, "Old Relic", "Artifact", "")
	ench := seedGraveyardCard(t, g, "Old Aura", "Enchantment", "")
	inst := seedGraveyardCard(t, g, "Old Instant", "Instant", "")
	sorc := seedGraveyardCard(t, g, "Old Sorcery", "Sorcery", "")
	bear := seedGraveyardCard(t, g, "Old Bear", "Creature — Bear", "")
	spell := d1658Cast(t, g, "Reconstruct History", "Sorcery", gapb42ReconstructOracle, game.CastSpellParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: art, Slot: 0},
			{Kind: game.TargetCard, ID: ench, Slot: 1},
			{Kind: game.TargetCard, ID: inst, Slot: 2},
			{Kind: game.TargetCard, ID: sorc, Slot: 3},
		},
	})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{art, ench, inst, sorc} {
		if !inHandOf(me, id) {
			t.Errorf("a targeted card was not returned to hand")
		}
	}
	if graveyardHas(me, bear) == false {
		t.Error("the untargeted creature left the graveyard")
	}
	if !exileHas(g, spell) {
		t.Error("Reconstruct History did not exile itself")
	}
	if graveyardHas(me, spell) {
		t.Error("Reconstruct History went to the graveyard")
	}
}

func TestGapB42WhatMustBeDoneLetTheWorldBurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")
	land := pushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	d1658Cast(t, g, "What Must Be Done", "Sorcery", gapb42WhatMustOracle, game.CastSpellParams{Modes: []int{0}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature or an artifact survived")
	}
	if !g.Battlefield.Contains(ench) || !g.Battlefield.Contains(land) {
		t.Error("an enchantment or a land was destroyed")
	}
}

func TestGapB42WhatMustBeDoneReleaseJunoAddsCountersToACreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		counters       int
	}{
		{"Legendary Bear", "Legendary Creature — Bear", 2},
		{"Old Relic", "Artifact", 0},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		dead := seedGraveyardCard(t, g, tc.name, tc.typeLine, "")
		plain := seedGraveyardCard(t, g, "Plain Bear", "Creature — Bear", "")
		if err := castGapB42RefusedTarget(t, g, plain); err == nil {
			t.Errorf("%s: a nonhistoric card was a legal target", tc.name)
		}
		d1658Cast(t, g, "What Must Be Done", "Sorcery", gapb42WhatMustOracle, game.CastSpellParams{
			Modes: []int{1}, Targets: cardRefs(dead),
		})
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(dead) {
			t.Fatalf("%s: the card was not returned to the battlefield", tc.name)
		}
		c, _ := battlefieldCard(g, dead)
		if c.Controller != me.ID {
			t.Errorf("%s: returned under the wrong control", tc.name)
		}
		if got := counterOn(g, dead, game.CounterPlusOne); got != tc.counters {
			t.Errorf("%s: %d +1/+1 counters, want %d", tc.name, got, tc.counters)
		}
	}
}

func castGapB42RefusedTarget(t *testing.T, g *game.Game, target uuid.UUID) error {
	t.Helper()
	return d1658CastErr(t, g, "What Must Be Done", "Sorcery", gapb42WhatMustOracle, game.CastSpellParams{
		Strict: true, Modes: []int{1}, Targets: cardRefs(target),
	})
}

func TestGapB42OverpoweringAttackUntapsAttackersAndAddsCombatAndMain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(idle) })
	if !isTapped(g, attacker) {
		t.Fatal("the attacker should be tapped after attacking")
	}
	spell := handCardFull(me, "Overpowering Attack", "Sorcery", "{3}{R}{R}", gapb42OverpoweringOracle, []string{"R"})
	floatForTest(g, me, "CCCRR")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, attacker) {
		t.Error("the creature that attacked was not untapped")
	}
	if !isTapped(g, idle) {
		t.Error("a creature that did not attack was untapped")
	}
	next := stepOnce(t, g)
	if next.Step != game.StepBeginCombat {
		t.Fatalf("after the spell: %+v, want the added combat", next)
	}
}
