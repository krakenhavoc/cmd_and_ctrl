package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b05_cards_test.go — March of the Machine, S82 batch 5 (#5477).
// Each test plays the real catalog entries through the engine.

const (
	momB05Kavu        = "20a362c3-d278-437f-881c-e7d9fe78fc2d"
	momB05Bondwarden  = "c17d4818-5825-4bff-8fc5-5fd740cc0c70"
	momB05Berserker   = "93107139-b743-4d62-b0ac-9997202e0562"
	momB05Negotiator  = "74b631c5-9420-496e-8845-3b050d3d5340"
	momB05Thrillseek  = "b3cc2ed9-1da3-4ad2-a9a7-de7665d800ff"
	momB05Artisan     = "ce8ec966-ecff-425e-944c-36dc8d4aef23"
	momB05Brawler     = "d8d5d0a4-9a56-404c-b459-a37dc5ae6bc1"
	momB05Storm       = "ae7643ea-ef1c-43a3-9605-4e8585239ea3"
	momB05RenderInert = "e9ece3c6-81c2-441a-acce-c109aaa9026b"
	momB05Favor       = "e2da8c3e-6ffc-41dd-a68a-62e64f7e9c60"
)

func TestMomB05EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		momB05Kavu: "Chomping Kavu", momB05Bondwarden: "Enduring Bondwarden",
		momB05Berserker: "Scorn-Blade Berserker", momB05Negotiator: "Streetwise Negotiator",
		momB05Thrillseek: "Voldaren Thrillseeker", momB05Artisan: "Axgard Artisan",
		momB05Brawler: "Botanical Brawler", momB05Storm: "Storm the Seedcore",
		momB05RenderInert: "Render Inert", momB05Favor: "Fertilid's Favor",
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

// momB05Backup casts a backup creature for the active player and aims its
// backup ability at target (uuid.Nil: itself), then resolves it.
func momB05Backup(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int, target uuid.UUID) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastHoldCreature(t, g, name, typeLine, oracle, p, tough)
	if target == uuid.Nil {
		target = id
	}
	effectstest.AnswerTriggerTargets(t, g, active.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Chomping Kavu -------------------------------------------------------

func TestMomB05ChompingKavuCantBeBlockedBySmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kavu := effectstest.B12Push(g, me.ID, "Chomping Kavu", "Creature — Kavu", momB05Kavu, 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, kavu)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, kavu); err == nil {
		t.Error("a power 2 creature blocked the Kavu")
	}
}

func TestMomB05ChompingKavuBackupGrantsTheRestrictionToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	kavu := momB05Backup(t, g, "Chomping Kavu", "Creature — Kavu", momB05Kavu, 3, 3, bear)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, kavu, game.CounterPlusOne); n != 0 {
		t.Errorf("the Kavu has %d counters, want 0", n)
	}
	if got := effectstest.ScopedBlockRuleCount(g); got != 1 {
		t.Fatalf("%d block-rule records, want 1", got)
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, bear); err == nil {
		t.Error("a power 2 creature blocked the creature that gained the restriction")
	}
	if err := g.DeclareBlocker(big, bear); err != nil {
		t.Errorf("a power 3 creature could not block it: %v", err)
	}
}

func TestMomB05ChompingKavuBackupOnItselfGrantsNothingExtra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	kavu := momB05Backup(t, g, "Chomping Kavu", "Creature — Kavu", momB05Kavu, 3, 3, uuid.Nil)
	if n := effectstest.CountersOn(g, kavu, game.CounterPlusOne); n != 1 {
		t.Errorf("the Kavu has %d counters, want 1", n)
	}
	if got := effectstest.ScopedBlockRuleCount(g); got != 0 {
		t.Errorf("%d block-rule records for a backup on itself, want 0", got)
	}
}

// --- Streetwise Negotiator -----------------------------------------------

func TestMomB05StreetwiseNegotiatorAssignsByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neg := effectstest.B12Push(g, me.ID, "Streetwise Negotiator", "Creature — Cat Citizen", momB05Negotiator, 0, 2)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, neg)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent at %d, want %d (the 0/2 deals its toughness)", opp.Life, before-2)
	}
}

func TestMomB05StreetwiseNegotiatorBackupGrantsItUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	momB05Backup(t, g, "Streetwise Negotiator", "Creature — Cat Citizen", momB05Negotiator, 0, 2, wall)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("opponent at %d, want %d (the 1/6 Wall deals its toughness)", opp.Life, before-6)
	}
}

// --- Enduring Bondwarden -------------------------------------------------

func TestMomB05EnduringBondwardenMovesItsCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bw := momB05Backup(t, g, "Enduring Bondwarden", "Creature — Human Scout", momB05Bondwarden, 0, 1, uuid.Nil)
	if n := effectstest.CountersOn(g, bw, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bondwarden has %d counters, want 1", n)
	}
	effectstest.StdB9Kill(t, g, bw)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
}

func TestMomB05EnduringBondwardenBackupGrantsTheDiesTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	momB05Backup(t, g, "Enduring Bondwarden", "Creature — Human Scout", momB05Bondwarden, 0, 1, bear)
	effectstest.StdB9Kill(t, g, bear)
	effectstest.AnswerTriggerTargets(t, g, me.ID, ox)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, ox, game.CounterPlusOne); n != 1 {
		t.Errorf("the Ox has %d +1/+1 counters, want the dead Bear's 1", n)
	}
}

// --- Scorn-Blade Berserker -----------------------------------------------

func TestMomB05ScornBladeBerserkerSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B")
	id := effectstest.B12Push(g, me.ID, "Scorn-Blade Berserker", "Creature — Human Berserker", momB05Berserker, 0, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || me.Hand.Size()-hand != 1 {
		t.Errorf("on battlefield %v, drew %d; want sacrificed and 1 card", g.Battlefield.Contains(id), me.Hand.Size()-hand)
	}
}

func TestMomB05ScornBladeBerserkerBackupGrantsTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	momB05Backup(t, g, "Scorn-Blade Berserker", "Creature — Human Berserker", momB05Berserker, 0, 1, bear)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the Bear gained no activated ability")
	}
	hand := me.Hand.Size()
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || me.Hand.Size()-hand != 1 {
		t.Errorf("the Bear survived %v or drew %d", g.Battlefield.Contains(bear), me.Hand.Size()-hand)
	}
}

// --- Voldaren Thrillseeker -----------------------------------------------

func TestMomB05VoldarenThrillseekerSacrificesForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	th := momB05Backup(t, g, "Voldaren Thrillseeker", "Creature — Vampire Warrior", momB05Thrillseek, 1, 1, uuid.Nil)
	if n := effectstest.CountersOn(g, th, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters, want 2", n)
	}
	life := opp.Life
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, th, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d (the 3/3's power)", opp.Life, life-3)
	}
}

func TestMomB05VoldarenThrillseekerBackupGrantsItToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Their Ogre", "Creature — Ogre", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	momB05Backup(t, g, "Voldaren Thrillseeker", "Creature — Vampire Warrior", momB05Thrillseek, 1, 1, bear)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the Bear gained no activated ability")
	}
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true, Targets: effectstest.CardRefs(victim),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 4/4 Bear's sacrifice (power 4) did not kill the 3/3")
	}
}

// --- Axgard Artisan ------------------------------------------------------

func momB05Treasures(g *game.Game, owner uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Treasure" {
				n++
			}
		}
	})
	return n
}

func TestMomB05AxgardArtisanMakesATreasureOnTheFirstPlusOneCountersEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := effectstest.B12Push(g, me.ID, "Axgard Artisan", "Creature — Dwarf Artificer", momB05Artisan, 2, 1)
	effectstest.StdB9AddCounter(t, g, art, "quest", 1)
	if got := momB05Treasures(g, me.ID); got != 0 {
		t.Fatalf("a quest counter made %d Treasures", got)
	}
	effectstest.StdB9AddCounter(t, g, art, game.CounterPlusOne, 1)
	if got := momB05Treasures(g, me.ID); got != 1 {
		t.Fatalf("first +1/+1 counters made %d Treasures, want 1", got)
	}
	effectstest.StdB9AddCounter(t, g, art, game.CounterPlusOne, 2)
	if got := momB05Treasures(g, me.ID); got != 1 {
		t.Errorf("second placement the same turn: %d Treasures, want still 1", got)
	}
}

// --- Botanical Brawler ---------------------------------------------------

func TestMomB05BotanicalBrawlerGrowsOnTheFirstCountersOnAnotherPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brawler := effectstest.B12Push(g, me.ID, "Botanical Brawler", "Creature — Elemental Warrior", momB05Brawler, 0, 0)
	if !effectstest.HasEffectiveKeyword(t, g, brawler, "trample") {
		t.Error("no trample")
	}
	effectstest.StdB9AddCounter(t, g, brawler, game.CounterPlusOne, 1)
	if n := effectstest.CountersOn(g, brawler, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters on the Brawler itself fed it: %d, want 1", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.StdB9AddCounter(t, g, theirs, game.CounterPlusOne, 1)
	if n := effectstest.CountersOn(g, brawler, game.CounterPlusOne); n != 1 {
		t.Errorf("an opponent's permanent fed it: %d, want 1", n)
	}
	effectstest.StdB9AddCounter(t, g, bear, "quest", 1)
	if n := effectstest.CountersOn(g, brawler, game.CounterPlusOne); n != 1 {
		t.Errorf("a quest counter fed it: %d, want 1", n)
	}
	effectstest.StdB9AddCounter(t, g, bear, game.CounterPlusOne, 1)
	if n := effectstest.CountersOn(g, brawler, game.CounterPlusOne); n != 2 {
		t.Fatalf("first +1/+1 counters on the Bear: Brawler has %d, want 2", n)
	}
	effectstest.StdB9AddCounter(t, g, bear, game.CounterPlusOne, 1)
	if n := effectstest.CountersOn(g, brawler, game.CounterPlusOne); n != 2 {
		t.Errorf("second placement on the Bear the same turn: Brawler has %d, want still 2", n)
	}
	land := effectstest.B12Permanent(g, me.ID, "Odd Land", "Land")
	effectstest.StdB9AddCounter(t, g, land, game.CounterPlusOne, 1)
	if n := effectstest.CountersOn(g, brawler, game.CounterPlusOne); n != 3 {
		t.Errorf("a noncreature permanent you control: Brawler has %d, want 3", n)
	}
}

func TestMomB05BotanicalBrawlerEntersWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.AvB16CastCreature(t, g, "Botanical Brawler", "Creature — Elemental Warrior", momB05Brawler, 0, 0)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Errorf("entered with %d +1/+1 counters, want 2", n)
	}
}

// --- Storm the Seedcore --------------------------------------------------

func TestMomB05StormTheSeedcoreDistributesAndGrantsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := effectstest.D1658CastErr(t, g, "Storm the Seedcore", "Sorcery", momB05Storm, game.CastSpellParams{
		Strict: true, Targets: effectstest.CardRefs(theirs), Distribution: map[uuid.UUID]int{theirs: 4},
	}); err == nil {
		t.Fatal("a creature you don't control was accepted")
	}
	effectstest.D1658Cast(t, g, "Storm the Seedcore", "Sorcery", momB05Storm, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, a, game.CounterPlusOne); n != 1 {
		t.Errorf("A: %d counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, b, game.CounterPlusOne); n != 3 {
		t.Errorf("B: %d counters, want 3", n)
	}
	for name, id := range map[string]uuid.UUID{"A": a, "B": b, "bystander": bystander} {
		for _, kw := range []string{"vigilance", "trample"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %s", name, kw)
			}
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature gained vigilance")
	}
}

// --- Render Inert --------------------------------------------------------

func TestMomB05RenderInertRemovesUpToFiveCountersThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	big := effectstest.B12Creature(g, opp.ID, "Pumped Bear", "Creature — Bear", 2, 2)
	for kind, n := range map[string]int{game.CounterPlusOne: 4, "shield": 3} {
		if err := g.AddCounter(big, kind, n); err != nil {
			t.Fatal(err)
		}
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Render Inert", "Sorcery", momB05RenderInert, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.NumberPrompt(t, g, me.ID)
	if first.PayAmount.Max != 4 {
		t.Fatalf("first question allows %d, want 4", first.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(first.ID, me.ID, 4); err != nil {
		t.Fatal(err)
	}
	second := effectstest.NumberPrompt(t, g, me.ID)
	if second.PayAmount.Max != 1 {
		t.Fatalf("second question allows %d, want the 1 left of five", second.PayAmount.Max)
	}
	if err := g.ResolvePayAmount(second.ID, me.ID, 1); err != nil {
		t.Fatal(err)
	}
	if n := effectstest.CountersOn(g, big, game.CounterPlusOne); n != 0 {
		t.Errorf("%d +1/+1 counters left, want 0", n)
	}
	if n := effectstest.CountersOn(g, big, "shield"); n != 2 {
		t.Errorf("%d shield counters left, want 2", n)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestMomB05RenderInertDrawsEvenWhenThePermanentHasNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	land := effectstest.B12Permanent(g, me.ID, "Odd Land", "Land")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Render Inert", "Sorcery", momB05RenderInert, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Fertilid's Favor ----------------------------------------------------

func TestMomB05FertilidsFavorFetchesALandAndPutsTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedSearchLibrary(opp, effectstest.SearchTestLand("Forest", "Basic Land — Forest"), effectstest.SearchTestLand("Plains", "Basic Land — Plains"))
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fertilid's Favor", "Instant", momB05Favor,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: bear, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, opp.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", n)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Forest" && c.Controller == opp.ID {
			found = true
			if !c.Tapped {
				t.Error("the land entered untapped")
			}
		}
	}
	if !found {
		t.Error("the targeted player's land did not enter under their control")
	}
}

func TestMomB05FertilidsFavorWorksWithNoSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
	effectstest.CastCatalogSpell(t, g, "Fertilid's Favor", "Instant", momB05Favor,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldNamed(g, "Forest") != 1 {
		t.Error("the Forest did not enter")
	}
}
