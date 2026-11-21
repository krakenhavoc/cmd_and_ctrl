package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB86DugganOracle     = "0fbe3805-b59b-47fc-ab21-f05a4aa4cfe0"
	gapB86VastraOracle     = "4789357c-3ab7-4327-b5ee-b1110ddb876f"
	gapB86FugitiveOracle   = "b3a03768-7676-4b33-a813-cc69dea77c4d"
	gapB86StraxOracle      = "ffde2558-2168-4b40-a31a-86d160bdb289"
	gapB86SixthOracle      = "49cb8d81-d8a0-4ae7-9750-6788a3e9c81b"
	gapB86FourthOracle     = "c19fdee3-5fbc-4e2d-ac23-85b46eacafdc"
	gapB86ForetoldOracle   = "02a45719-ee4a-49c4-96ec-a140cd6e1547"
	gapB86TypeDoctor       = "Legendary Creature — Time Lord Doctor"
	gapB86TypeArtifactSpec = "Artifact"
)

func TestGapB86EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB86DugganOracle:   "Duggan, Private Detective",
		gapB86VastraOracle:   "Madame Vastra",
		gapB86FugitiveOracle: "The Fugitive Doctor",
		gapB86StraxOracle:    "Strax, Sontaran Nurse",
		gapB86SixthOracle:    "The Sixth Doctor",
		gapB86FourthOracle:   "The Fourth Doctor",
		gapB86ForetoldOracle: "The Foretold Soldier",
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

func gapB86Count(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

func TestGapB86DugganSizeInvestigateAndTheOnlyPunch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	for i := 0; i < 3; i++ {
		handCardOf(g, me, "Card", "Instant", "{R}", "")
	}
	duggan := castCatalogSpell(t, g, "Duggan, Private Detective", gapB86TypeDoctor, gapB86DugganOracle, nil)
	passPriorityAroundTable(t, g)
	if n := gapB86Count(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("entering made %d Clues, want 1", n)
	}
	if p, tough := effectivePower(t, g, duggan), effectiveToughness(t, g, duggan); p != 3 || tough != 3 {
		t.Fatalf("with three cards in hand Duggan is %d/%d, want 3/3", p, tough)
	}
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Ogre", 1, 6)
	survivor := b12Creature(g, opp.ID, "Survivor", "Creature — Ogre", 1, 7)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, duggan).SummonedThisTurn = false
		fillPoolColored(me, "G", 1)
		fillPool(me, 3)
	})
	if err := g.ActivateCatalogAbility(me.ID, duggan, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(duggan)}); err == nil {
		t.Fatal("Duggan targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, duggan, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("the punch: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("twice Duggan's power (6) kills a 1/6")
	}
	// Activate only once, even after untapping.
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, duggan).Tapped = false })
	if err := g.ActivateCatalogAbility(me.ID, duggan, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(survivor)}); err == nil {
		t.Error("the punch was activated a second time")
	}
}

func TestGapB86VastraMakesAClueAndAFoodWhenACreatureSheHurtDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vastra := apaPush(g, me.ID, me.ID, game.Card{Name: "Madame Vastra", OracleID: gapB86VastraOracle,
		TypeLine: "Legendary Creature — Lizard Detective", Power: 3, Toughness: 6})
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Rat", 1, 2)
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Rat", 1, 1)
	declareAttack(t, g, opp.ID, vastra)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, vastra); err != nil {
		t.Fatalf("block: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Fatal("a 3-power Vastra kills a 1/2")
	}
	if n := gapB86Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
	if n := gapB86Count(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Foods, want 1", n)
	}
	// A creature she never damaged dying makes nothing.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bystander) })
	passPriorityAroundTable(t, g)
	if n := gapB86Count(g, me.ID, "Clue"); n != 1 {
		t.Errorf("an undamaged creature's death made a Clue: %d", n)
	}
}

func TestGapB86FugitiveDoctorInvestigatesAndGivesFlashbackForAClue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doc := castCatalogSpell(t, g, "The Fugitive Doctor", gapB86TypeDoctor, gapB86FugitiveOracle, nil)
	passPriorityAroundTable(t, g)
	if n := gapB86Count(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("entering made %d Clues, want 1", n)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, doc).SummonedThisTurn = false })
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Sorcery", TypeLine: "Sorcery", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	declareAttack(t, g, opp.ID, doc)
	passPriorityAroundTable(t, g)
	clue := func() uuid.UUID {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Clue") {
				return c.InstanceID
			}
		}
		return uuid.Nil
	}()
	answerOwnPermanents(t, g, me.ID, clue)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(clue) {
		t.Fatal("the Clue was not sacrificed")
	}
	advanceToMain(t, g)
	g.WithWriteLock(func() { fillPoolColored(me, "R", 1); fillPoolColored(me, "G", 1); fillPool(me, 2) })
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback from the graveyard: %v", err)
	}
}

func TestGapB86StraxFightsAPlayersCreatureAtRandomAndGrows(t *testing.T) {
	for attempt := 0; attempt < 16; attempt++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		strax := apaPush(g, me.ID, me.ID, game.Card{Name: "Strax, Sontaran Nurse", OracleID: gapB86StraxOracle,
			TypeLine: "Legendary Creature — Alien Cleric", Power: 5, Toughness: 5})
		rock := b12Permanent(g, me.ID, "Rock", "Artifact")
		var theirs []uuid.UUID
		for _, p := range g.Seats[1:] {
			theirs = append(theirs, b12Creature(g, p.ID, "Rat", "Creature — Rat", 1, 1))
		}
		g.WithWriteLock(func() { fillPool(me, 2) })
		if err := g.ActivateCatalogAbility(me.ID, strax, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}}); err != nil {
			t.Fatalf("grenades: %v", err)
		}
		if g.Battlefield.Contains(rock) {
			t.Fatal("the artifact was not sacrificed as a cost")
		}
		passPriorityAroundTable(t, g)
		for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				break
			}
		}
		if latestPickTarget(g, me.ID) == nil {
			// The random player was Strax's controller: nothing else of theirs to fight.
			for _, id := range theirs {
				if !g.Battlefield.Contains(id) {
					t.Fatal("no target prompt yet a creature died")
				}
			}
			continue
		}
		fought := uuid.Nil
		for _, id := range theirs {
			p := latestPickTarget(g, me.ID)
			if p == nil {
				break
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
				fought = id
				break
			}
		}
		if fought == uuid.Nil {
			t.Fatal("none of the opposing creatures was a legal target")
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(fought) {
			t.Error("a 5/5 Strax fights a 1/1: it should die")
		}
		if n := countersOn(g, strax, game.CounterPlusOne); n != 1 {
			t.Errorf("Strax has %d counters after dealing damage to a creature, want 1", n)
		}
		return
	}
	t.Fatal("the random player was never an opponent in 16 tries")
}

func TestGapB86SixthDoctorCopiesAHistoricSpellOnceEachTurnWithoutLegendary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "The Sixth Doctor", gapB86TypeDoctor, gapB86SixthOracle, false)
	castCatalogSpell(t, g, "Legend", "Legendary Creature — Elf", "", nil)
	passPriorityAroundTable(t, g)
	var elves, tokens int
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Elf") {
			elves++
			if c.IsToken() {
				tokens++
				if c.HasSupertype("Legendary") {
					t.Error("the copy is legendary")
				}
			}
		}
	}
	if elves != 2 || tokens != 1 {
		t.Fatalf("%d Elves (%d tokens), want the original and one token copy", elves, tokens)
	}
	// The second historic spell this turn is not copied.
	castCatalogSpell(t, g, "Second Rock", gapB86TypeArtifactSpec, "", nil)
	passPriorityAroundTable(t, g)
	rocks := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Second Rock" {
			rocks++
		}
	}
	if rocks != 1 {
		t.Errorf("%d Second Rocks, want 1 (the ability triggers once each turn)", rocks)
	}
	// A nonhistoric spell is never copied.
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := gapB86Count(g, me.ID, "Bear"); n != 1 {
		t.Errorf("%d Bears", n)
	}
}

func TestGapB86FourthDoctorPlaysOneHistoricCardFromTheTopEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "The Fourth Doctor", gapB86TypeDoctor, gapB86FourthOracle, false)
	advanceToMain(t, g)
	plain := seedLibraryTop(me, "Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("a nonhistoric creature was cast from the top")
	}
	rock := seedLibraryTop(me, "Rock", "Artifact", "")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("a historic spell from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := gapB86Foods(g, me.ID); n != 1 {
		t.Fatalf("%d Foods after the first historic play, want 1", n)
	}
	rock2 := seedLibraryTop(me, "Rock Two", "Artifact", "")
	if err := g.CastSpell(me.ID, rock2, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("a second historic spell from the top in the same turn")
	}
}

func TestGapB86FourthDoctorFoodForAHistoricLandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "The Fourth Doctor", gapB86TypeDoctor, gapB86FourthOracle, false)
	advanceToMain(t, g)
	land := seedLibraryTop(me, "Legendary Land", "Legendary Land", "")
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("a historic land from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land was not played")
	}
	if n := gapB86Foods(g, me.ID); n != 1 {
		t.Fatalf("%d Foods after the land play, want 1", n)
	}
}

func TestGapB86ForetoldSoldierExilesItselfForetoldWhenItDealsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soldier := apaPush(g, me.ID, me.ID, game.Card{Name: "The Foretold Soldier", OracleID: gapB86ForetoldOracle,
		TypeLine: "Creature — Alien Zombie Soldier", Power: 6, Toughness: 6})
	declareAttack(t, g, opp.ID, soldier)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(soldier) {
		t.Fatal("the Soldier stayed on the battlefield after dealing damage")
	}
	var exiled *game.Card
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == soldier {
			exiled = &g.Exile.Cards[i]
		}
	}
	if exiled == nil || !game.CardIsForetold(*exiled) {
		t.Fatalf("the Soldier is %+v, want a foretold card in exile", exiled)
	}
	// Not castable this turn, castable for {1}{G} on a later turn.
	g.WithWriteLock(func() { fillPoolColored(me, "G", 1); fillPool(me, 1) })
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, soldier, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: "foretell"}); err == nil {
		t.Fatal("cast the foretold Soldier on the turn it was foretold")
	}
	// On a later turn of its owner it casts for {1}{G}.
	start := g.Turn.Seq
	for i := 0; i < 400 && !(g.Turn.Seq > start && g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Seq == start || g.Turn.ActiveSeat != 0 {
		t.Fatal("never reached the owner's later turn")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool(); fillPoolColored(me, "G", 1); fillPool(me, 1) })
	if err := g.CastSpell(me.ID, soldier, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: "foretell"}); err != nil {
		t.Fatalf("cast the foretold Soldier on a later turn for {1}{G}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(soldier) {
		t.Error("the Soldier did not return to the battlefield")
	}
}

func gapB86Foods(g *game.Game, controller uuid.UUID) int { return gapB86Count(g, controller, "Food") }
