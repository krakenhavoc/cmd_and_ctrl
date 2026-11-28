package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b852_cards_test.go — Fallen Empires cards, S82 OLD batch 852 (#5637),
// played through the real catalog entries.

const (
	o852Bed      = "173f2b8e-8072-4fdd-a469-8a1fb0bb6e64"
	o852Hunter   = "49b351d8-8b63-4f9e-a323-344ed35fdfff"
	o852Medic    = "541b7b2f-7bb0-4858-8dc5-6e50de443c33"
	o852Kites    = "6efd11c8-be8a-4d44-a776-66d058638ad1"
	o852Warrior  = "6facb0f6-fa49-4ce6-9f95-15bb3ce4bcf2"
	o852Infantry = "9eb274f3-ef55-4545-81b4-29454a48f075"
	o852Veteran  = "a3b66180-ee08-4bef-90aa-7859dc77cb10"
	o852Druid    = "adbc8970-1bbc-4f30-9a2e-eebaf074eae6"
	o852Flats    = "bf39d0f7-f689-45fd-b37b-12d1618fd6e7"
	o852Warrens  = "c1161505-1623-4475-a88e-2e48072cc2f1"
	o852Spawn    = "d1aba343-d85a-4829-bdfe-db31fcb3c320"
	o852Priest   = "fb7adbee-a3b0-42bf-988f-9d4faf6bd006"
)

func TestS82OldB852CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o852Bed: "Homarid Spawning Bed", o852Hunter: "Elvish Hunter", o852Medic: "Combat Medic", o852Kites: "Goblin Kites",
		o852Warrior: "Homarid Warrior", o852Infantry: "Icatian Infantry", o852Veteran: "Orcish Veteran",
		o852Druid: "Thelonite Druid", o852Flats: "Tidal Flats", o852Warrens: "Goblin Warrens", o852Spawn: "Deep Spawn",
		o852Priest: "Svyelunite Priest",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o852Perm(g *game.Game, owner uuid.UUID, name, typeLine, cost string, p, tough int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func o852Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			n++
		}
	}
	return n
}

func TestS82OldB852HomaridSpawningBedMakesCamaridsEqualToTheSacrificedManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bed := effectstest.B12Push(g, me.ID, "Homarid Spawning Bed", "Enchantment", o852Bed, 0, 0)
	blue := o852Perm(g, me.ID, "Blue Three", "Creature — Merfolk", "{2}{U}", 2, 2, "U")
	red := o852Perm(g, me.ID, "Red Three", "Creature — Goblin", "{2}{R}", 2, 2, "R")
	effectstest.B06AddMana(me, "C", "U", "U", "C", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, bed, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("sacrificed a nonblue creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, bed, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{blue}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blue) {
		t.Error("the blue creature was not sacrificed")
	}
	if n := o852Count(g, me.ID, "Camarid"); n != 3 {
		t.Errorf("%d Camarid tokens, want 3", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Camarid" && (c.Power != 1 || c.Toughness != 1 || !c.HasColorWord("U")) {
			t.Errorf("Camarid is %d/%d %v, want a 1/1 blue token", c.Power, c.Toughness, c.Colors)
		}
	}
	// A blue token has mana value 0: no Camarids.
	tok := o852Perm(g, me.ID, "Blue Token", "Token Creature — Illusion", "", 1, 1, "U")
	effectstest.B06AddMana(me, "C", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, bed, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{tok}}); err != nil {
		t.Fatalf("activate with a token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := o852Count(g, me.ID, "Camarid"); n != 3 {
		t.Errorf("%d Camarids after sacrificing a token, want still 3", n)
	}
}

func TestS82OldB852ElvishHunterFreezesATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hunter := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Hunter", "Creature — Elf Archer", o852Hunter, false)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardFor(g, foe).Tapped = true })
	effectstest.B06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, hunter, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.BattlefieldCardFor(g, hunter).Tapped {
		t.Error("the Hunter did not tap")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.HmlUntapSkips(t, g, foe)) == 0 {
		t.Error("the target was not marked to skip its next untap step")
	}
	// Through the target's own controller's next untap step it stays tapped.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.BattlefieldCardFor(g, foe).Tapped {
		t.Error("the frozen creature untapped")
	}
}

func TestS82OldB852CombatMedicPreventsTheNextDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	medic := effectstest.B12Push(g, me.ID, "Combat Medic", "Creature — Human Cleric Soldier", o852Medic, 0, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Human", 1, 1)
	effectstest.B06AddMana(me, "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, medic, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealDamageTo(t, g, src, bear, 3)
	if d := effectstest.BattlefieldCardFor(g, bear).DamageMarked; d != 2 {
		t.Errorf("%d damage marked, want 2 (1 prevented)", d)
	}
	// It is a one-damage shield: a second hit lands whole.
	effectstest.DealDamageTo(t, g, src, bear, 1)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		if d := effectstest.BattlefieldCardFor(g, bear).DamageMarked; d != 3 {
			t.Errorf("%d damage marked after the second hit, want 3", d)
		}
	}
}

func TestS82OldB852HomaridWarriorAndDeepSpawnGainShroudFreezeAndTap(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, tl string
		p, tough         int
	}{
		{"Homarid Warrior", o852Warrior, "Creature — Homarid Warrior", 3, 3},
		{"Deep Spawn", o852Spawn, "Creature — Homarid", 6, 6},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		c := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.tl, tc.oracle, true) // summoning sick: no {T} in the cost
		effectstest.B06AddMana(me, "U")
		if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, c, "shroud") {
			t.Errorf("%s: no shroud", tc.name)
		}
		if !effectstest.BattlefieldCardFor(g, c).Tapped {
			t.Errorf("%s: not tapped", tc.name)
		}
		if len(effectstest.HmlUntapSkips(t, g, c)) == 0 {
			t.Errorf("%s: not marked to skip its next untap step", tc.name)
		}
	}
}

func TestS82OldB852DeepSpawnMillsTwoOrIsSacrificedInUpkeep(t *testing.T) {
	for _, mill := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		spawn := effectstest.B12Push(g, me.ID, "Deep Spawn", "Creature — Homarid", o852Spawn, 6, 6)
		lib := me.Library.Size()
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if mill {
			effectstest.AnswerOptionPick(t, g, me.ID, 0)
		} else {
			effectstest.AnswerOptionPick(t, g, me.ID, 1)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if mill {
			if !g.Battlefield.Contains(spawn) {
				t.Error("milling two did not keep the Spawn")
			}
			if me.Library.Size() >= lib {
				t.Errorf("library %d → %d, want two milled", lib, me.Library.Size())
			}
		} else if g.Battlefield.Contains(spawn) {
			t.Error("declining did not sacrifice the Spawn")
		}
	}
	// A library of fewer than two cards can't pay.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spawn := effectstest.B12Push(g, me.ID, "Deep Spawn", "Creature — Homarid", o852Spawn, 6, 6)
	g.WithWriteLock(func() { me.Library.Cards = me.Library.Cards[:1] })
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spawn) {
		t.Error("a one-card library paid for the Spawn")
	}
}

func TestS82OldB852IcatianInfantryGainsFirstStrikeAndBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	inf := effectstest.B12Push(g, me.ID, "Icatian Infantry", "Creature — Human Soldier", o852Infantry, 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, inf, "first strike") || effectstest.HasEffectiveKeyword(t, g, inf, game.KeywordBanding) {
		t.Fatal("the Infantry starts with a grant")
	}
	for i, kw := range []string{"first strike", game.KeywordBanding} {
		effectstest.B06AddMana(me, "C")
		if err := g.ActivateCatalogAbility(me.ID, inf, i, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %s: %v", kw, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, inf, kw) {
			t.Errorf("no %s after its ability", kw)
		}
	}
}

func TestS82OldB852OrcishVeteranCantBlockBigWhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vet := effectstest.B12Push(g, opp.ID, "Orcish Veteran", "Creature — Orc", o852Veteran, 2, 2)
	bigWhite := o852Perm(g, me.ID, "Big White", "Creature — Knight", "{2}{W}", 2, 2, "W")
	smallWhite := o852Perm(g, me.ID, "Small White", "Creature — Knight", "{W}", 1, 1, "W")
	bigRed := o852Perm(g, me.ID, "Big Red", "Creature — Goblin", "{2}{R}", 3, 3, "R")
	effectstest.BrAttack(t, g, bigWhite, smallWhite, bigRed)
	if err := g.DeclareBlocker(vet, bigWhite); err == nil {
		t.Error("the Veteran blocked a white creature with power 2")
	}
	if err := g.DeclareBlocker(vet, smallWhite); err != nil {
		t.Errorf("the Veteran could not block a white creature with power 1: %v", err)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	vet2 := effectstest.B12Push(g2, opp2.ID, "Orcish Veteran", "Creature — Orc", o852Veteran, 2, 2)
	red2 := o852Perm(g2, me2.ID, "Big Red", "Creature — Goblin", "{2}{R}", 3, 3, "R")
	effectstest.BrAttack(t, g2, red2)
	if err := g2.DeclareBlocker(vet2, red2); err != nil {
		t.Errorf("the Veteran could not block a red creature: %v", err)
	}
}

func TestS82OldB852OrcishVeteranGainsFirstStrikeForR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vet := effectstest.B12Push(g, me.ID, "Orcish Veteran", "Creature — Orc", o852Veteran, 2, 2)
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, vet, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, vet, "first strike") {
		t.Error("no first strike")
	}
}

func TestS82OldB852TheloniteDruidAnimatesYourForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	druid := effectstest.PushCatalogPermanent(g, me.ID, "Thelonite Druid", "Creature — Human Cleric Druid", o852Druid, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.B06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	if p, tough := effectstest.PtOf(t, g, forest); p != 2 || tough != 3 {
		t.Errorf("your Forest is %d/%d, want 2/3", p, tough)
	}
	for _, id := range []uuid.UUID{forest} {
		c := effectstest.BattlefieldCardCopy(t, g, id)
		if !c.IsCreature() || !c.IsLand() {
			t.Errorf("your Forest is creature %v, land %v; want both", c.IsCreature(), c.IsLand())
		}
	}
	for name, id := range map[string]uuid.UUID{"your Island": island, "an opponent's Forest": theirs, "a Forest that arrived later": late} {
		if effectstest.BattlefieldCardCopy(t, g, id).IsCreature() {
			t.Errorf("%s became a creature", name)
		}
	}
}

func TestS82OldB852GoblinWarrensTurnsTwoGoblinsIntoThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	warrens := effectstest.B12Push(g, me.ID, "Goblin Warrens", "Enchantment", o852Warrens, 0, 0)
	g1 := o852Perm(g, me.ID, "Goblin A", "Creature — Goblin", "{R}", 1, 1, "R")
	g2 := o852Perm(g, me.ID, "Goblin B", "Creature — Goblin", "{R}", 1, 1, "R")
	bear := o852Perm(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2, "G")
	effectstest.B06AddMana(me, "C", "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, warrens, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{g1, bear}}); err == nil {
		t.Fatal("sacrificed a non-Goblin")
	}
	if err := g.ActivateCatalogAbility(me.ID, warrens, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{g1, g2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(g1) || g.Battlefield.Contains(g2) {
		t.Error("the Goblins were not sacrificed")
	}
	if n := o852Count(g, me.ID, "Goblin"); n != 3 {
		t.Errorf("%d Goblin tokens, want 3", n)
	}
}

func TestS82OldB852SvyelunitePriestGrantsShroudOnlyInYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Svyelunite Priest", "Creature — Merfolk Cleric", o852Priest, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, priest, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("activated outside your upkeep")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	priest2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Svyelunite Priest", "Creature — Merfolk Cleric", o852Priest, false)
	bear2 := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g2, 0)
	effectstest.B06AddMana(me2, "U", "U")
	if err := g2.ActivateCatalogAbility(me2.ID, priest2, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear2)}); err != nil {
		t.Fatalf("activate in upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.HasEffectiveKeyword(t, g2, bear2, "shroud") {
		t.Error("the target has no shroud")
	}
}

func TestS82OldB852GoblinKitesGrantsFlyingThenFlipsAtTheEndStep(t *testing.T) {
	sawLoss, sawWin := false, false
	for attempt := 0; attempt < 60 && !(sawLoss && sawWin); attempt++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		kites := effectstest.B12Push(g, me.ID, "Goblin Kites", "Enchantment", o852Kites, 0, 0)
		small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Goblin", 1, 2)
		big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
		effectstest.B06AddMana(me, "R", "R")
		if err := g.ActivateCatalogAbility(me.ID, kites, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(big)}); err == nil {
			t.Fatal("targeted a creature with toughness 3")
		}
		if err := g.ActivateCatalogAbility(me.ID, kites, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(small)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, small, "flying") {
			t.Fatal("no flying")
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		call := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if call == nil {
			t.Fatal("no coin was flipped at the end step")
		}
		before := len(g.Events)
		if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
			t.Fatal(err)
		}
		won := false
		for _, ev := range g.Events[before:] {
			if ev.Kind == game.EventFlipCoin {
				won = ev.Won
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if won {
			sawWin = true
			if !g.Battlefield.Contains(small) {
				t.Error("a won flip still sacrificed the creature")
			}
		} else {
			sawLoss = true
			if g.Battlefield.Contains(small) {
				t.Error("a lost flip did not sacrifice the creature")
			}
		}
		if !g.Battlefield.Contains(big) {
			t.Error("the other creature was sacrificed")
		}
	}
	if !sawWin || !sawLoss {
		t.Errorf("coverage: won flip %v, lost flip %v", sawWin, sawLoss)
	}
}

func TestS82OldB852TidalFlatsGivesBlockersFirstStrikeUnlessTheAttackerPays(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1] // seat 0 attacks seat 1, who owns the Flats
		flats := effectstest.B12Push(g, opp.ID, "Tidal Flats", "Enchantment", o852Flats, 0, 0)
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
		flyer := effectstest.B12Creature(g, me.ID, "Flyer", "Creature — Bird", 2, 2)
		g.WithWriteLock(func() {
			effectstest.BattlefieldCardFor(g, flyer).Keywords = append(effectstest.BattlefieldCardFor(g, flyer).Keywords, "flying")
		})
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		effectstest.BrAttack(t, g, attacker, flyer)
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatalf("block: %v", err)
		}
		effectstest.B06AddMana(opp, "U", "U")
		if err := g.ActivateCatalogAbility(opp.ID, flats, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("the attacker's controller was not asked whether to pay {1}")
		}
		if pay {
			effectstest.B06AddMana(me, "C")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("the flying attacker was asked too")
		}
		if got := effectstest.HasEffectiveKeyword(t, g, blocker, "first strike"); got == pay {
			t.Errorf("pay=%v: blocker first strike = %v", pay, got)
		}
	}
}
