package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_seam_6098_cards_test.go — the GAP2 batch 11 cards that waited on seam
// #6098 (S82): Noble Heritage, Shelob, Child of Ungoliant, Baeloth Barrityl,
// Entertainer and Emrakul, the World Anew. Each test plays the real catalog
// entry and asserts what its text says.

const (
	seam6098NobleHeritageOracle = "2e342aa3-2590-4b59-852a-9adb31bafe5c"
	seam6098ShelobOracle        = "8372115d-c963-4a41-8759-72f7f6a8935a"
	seam6098BaelothOracle       = "33b401b1-cf94-4c7a-8a8a-426ac3d87bd0"
	seam6098EmrakulOracle       = "33300e69-29f6-41e4-92ab-82d8b018b217"
)

func TestSeam6098CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		seam6098NobleHeritageOracle: "Noble Heritage",
		seam6098ShelobOracle:        "Shelob, Child of Ungoliant",
		seam6098BaelothOracle:       "Baeloth Barrityl, Entertainer",
		seam6098EmrakulOracle:       "Emrakul, the World Anew",
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

func seam6098Seats(g *game.Game) (me, opp, third *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
}

func seam6098PlayerHas(g *game.Game, p *game.Player, kw string) bool {
	var has bool
	g.WithWriteLock(func() { has = slices.Contains(g.PlayerAbilitiesForEffect(p), kw) })
	return has
}

// --- Noble Heritage ----------------------------------------------------------

// The commander has the granted ability: at its controller's upkeep each
// player may put two counters on a creature; the controller gains
// protection from each opponent who did, and from nobody else.
func TestSeam6098NobleHeritageGivesProtectionFromEachOpponentWhoTookCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := seam6098Seats(g)
	effectstest.B12Permanent(g, me.ID, "Noble Heritage", "Legendary Enchantment — Background")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Noble Heritage" {
				g.Battlefield.Cards[i].OracleID = seam6098NobleHeritageOracle
			}
		}
	})
	cmdr := effectstest.B12Creature(g, me.ID, "My Commander", "Legendary Creature — Human", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == cmdr {
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	oppBear := effectstest.B12Creature(g, opp.ID, "Opp Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, third.ID, "Third Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayCounter(t, g, me.ID, cmdr)
	effectstest.AnswerMayCounter(t, g, opp.ID, oppBear)
	effectstest.AnswerMayCounter(t, g, third.ID, uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountersOn(g, oppBear, game.CounterPlusOne); n != 2 {
		t.Fatalf("the opponent's creature has %d +1/+1 counters, want 2", n)
	}
	if n := effectstest.CountersOn(g, cmdr, game.CounterPlusOne); n != 2 {
		t.Errorf("my commander has %d +1/+1 counters, want 2", n)
	}
	if !seam6098PlayerHas(g, me, game.ProtectionFromPlayer(opp.ID)) {
		t.Error("I have no protection from the opponent who took the counters")
	}
	if seam6098PlayerHas(g, me, game.ProtectionFromPlayer(third.ID)) {
		t.Error("I have protection from the opponent who declined")
	}
	if seam6098PlayerHas(g, me, game.ProtectionFromPlayer(me.ID)) {
		t.Error("I have protection from myself")
	}
}

// A creature that is not a commander does not have the ability.
func TestSeam6098NobleHeritageReachesOnlyCommanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := seam6098Seats(g)
	heritage := effectstest.B12Permanent(g, me.ID, "Noble Heritage", "Legendary Enchantment — Background")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == heritage {
				g.Battlefield.Cards[i].OracleID = seam6098NobleHeritageOracle
			}
		}
	})
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		t.Error("a non-commander creature had Noble Heritage's ability")
	}
}

// --- Shelob, Child of Ungoliant ---------------------------------------------

func seam6098TokenNamed(g *game.Game, name string, controller uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == name && c.IsToken() && c.Controller == controller {
			return c
		}
	}
	return nil
}

// A Spider I control blocks and kills a creature: Shelob makes a token copy
// of it that is a Food artifact and no longer a creature, with the Food
// ability.
func TestSeam6098ShelobCopiesACreatureASpiderKilledAsAFoodArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := seam6098Seats(g)
	effectstest.B12Push(g, opp.ID, "Shelob, Child of Ungoliant", "Legendary Creature — Spider Demon", seam6098ShelobOracle, 8, 8)
	spider := effectstest.B12Creature(g, opp.ID, "Little Spider", "Creature — Spider", 1, 1)
	ogre := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, spider, "deathtouch") {
		t.Fatal("another Spider does not have deathtouch")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, ogre)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{spider: ogre})
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.FindBattlefieldCardForTest(g, ogre) != nil {
		t.Fatal("setup: the Ogre survived the deathtouch Spider")
	}
	tok := seam6098TokenNamed(g, "Ogre", opp.ID)
	if tok == nil {
		t.Fatal("Shelob made no token copy of the Ogre")
	}
	if tok.IsCreature() || !tok.IsArtifact() || !tok.HasSubtype("Food") || tok.HasSubtype("Ogre") {
		t.Errorf("the copy's type line is %q: want a Food artifact and nothing else", tok.TypeLine)
	}
	if !slices.Contains(tok.GrantedAbilities, game.GrantKey("shelob-child-of-ungoliant/food-ability")) {
		t.Errorf("the Food copy does not have the quoted Food ability: %v", tok.GrantedAbilities)
	}
}

// A creature killed by a non-Spider is not copied.
func TestSeam6098ShelobIgnoresACreatureNoSpiderDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := seam6098Seats(g)
	effectstest.B12Push(g, opp.ID, "Shelob, Child of Ungoliant", "Legendary Creature — Spider Demon", seam6098ShelobOracle, 8, 8)
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wolf: bear})
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("setup: the Bear survived")
	}
	if seam6098TokenNamed(g, "Bear", opp.ID) != nil {
		t.Error("Shelob copied a creature no Spider had damaged")
	}
}

// --- Baeloth Barrityl, Entertainer -----------------------------------------

// An opponent's creature with less power than Baeloth is goaded, one with
// as much is not; when the goaded one dies attacking, Baeloth's controller
// creates a Treasure, and when the other dies attacking, nobody does.
func TestSeam6098BaelothGoadsSmallerCreaturesAndPaysForTheirDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := seam6098Seats(g)
	effectstest.B12Push(g, me.ID, "Baeloth Barrityl, Entertainer", "Legendary Creature — Elf Shaman", seam6098BaelothOracle, 2, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Goblin", 1, 1)
	even := effectstest.B12Creature(g, opp.ID, "Even", "Creature — Goblin", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Goblin", 1, 1)
	wall := effectstest.B12Creature(g, third.ID, "Wall", "Creature — Wall", 0, 6)

	goaded := func(id uuid.UUID) bool {
		var out bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			out = effectstest.FindBattlefieldCardForTest(g, id).Goaded()
		})
		return out
	}
	if !goaded(small) || goaded(even) || goaded(mine) {
		t.Fatalf("goaded: small %v, even %v, mine %v; want only the smaller opponent creature", goaded(small), goaded(even), goaded(mine))
	}

	effectstest.AdvanceToDeclareAttackersOf(t, g, (g.Turn.ActiveSeat+1)%4)
	effectstest.Pr7bAttack(t, g, third.ID, small, even)
	// The wall kills nothing: destroy both attackers while they are
	// attacking (one of them blocked), so each dies in combat.
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{wall: small})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(small); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(even); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Treasure", me.ID); n != 1 {
		t.Errorf("Baeloth's controller has %d Treasures, want 1 (only the goaded attacker counts)", n)
	}
}

// --- Emrakul, the World Anew ------------------------------------------------

// Casting it takes every creature the target player controls, before the
// spell itself resolves.
func TestSeam6098EmrakulCastTriggerTakesTheTargetPlayersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := seam6098Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "Opp One", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Opp Two", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, third.ID, "Third", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreatureNoResolve(t, g, "Emrakul, the World Anew", "Legendary Creature — Eldrazi", seam6098EmrakulOracle, 12, 12)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.ControllerOf(t, g, id); got != me.ID {
			t.Errorf("%s is controlled by %s, want the caster", id, got)
		}
	}
	if got := effectstest.ControllerOf(t, g, other); got != third.ID {
		t.Error("a creature of a player who was not the target changed hands")
	}
}

// Protection from permanents that were cast this turn: a creature whose
// spell was cast this turn can't block Emrakul; one that was not can. And
// a spell is refused, an ability is not.
func TestSeam6098EmrakulHasProtectionFromSpellsAndPermanentsCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := seam6098Seats(g)
	emrakul := effectstest.B12Push(g, me.ID, "Emrakul, the World Anew", "Legendary Creature — Eldrazi", seam6098EmrakulOracle, 12, 12)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh", "Creature — Bear", 2, 2)
	old := effectstest.B12Creature(g, opp.ID, "Old", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			// Reach, so flying is not what stops a block.
			switch g.Battlefield.Cards[i].InstanceID {
			case fresh:
				g.Battlefield.Cards[i].Provenance.CastThisTurn = true
				g.Battlefield.Cards[i].Keywords = []string{"reach"}
			case old:
				g.Battlefield.Cards[i].Keywords = []string{"reach"}
			}
		}
	})
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		e := effectstest.FindBattlefieldCardForTest(g, emrakul)
		if _, ok := game.MatchedProtection(e, &game.Characteristic{SourceIsSpell: true}); !ok {
			t.Error("Emrakul has no protection from a colorless spell")
		}
		if _, ok := game.MatchedProtection(e, &game.Characteristic{SourceIsPermanent: true}); ok {
			t.Error("Emrakul has protection from an ability of a permanent that was not cast this turn")
		}
	})
	effectstest.Pr7bAttack(t, g, opp.ID, emrakul)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.Clone().DeclareBlocker(fresh, emrakul); err == nil {
		t.Error("a creature cast this turn blocked Emrakul")
	}
	if err := g.Clone().DeclareBlocker(old, emrakul); err != nil {
		t.Errorf("a creature not cast this turn could not block Emrakul: %v", err)
	}
}

// When Emrakul leaves, its controller sacrifices every creature they
// control, and nobody else does.
func TestSeam6098EmrakulLeavingSacrificesAllYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := seam6098Seats(g)
	emrakul := effectstest.B12Push(g, me.ID, "Emrakul, the World Anew", "Legendary Creature — Eldrazi", seam6098EmrakulOracle, 12, 12)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(emrakul); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, mine) != nil {
		t.Error("my creature survived Emrakul leaving")
	}
	if effectstest.FindBattlefieldCardForTest(g, theirs) == nil {
		t.Error("an opponent's creature was sacrificed")
	}
}
