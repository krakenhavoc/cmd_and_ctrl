package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b18_cards_test.go — S82 reprint pass, batch 18 (#5908).

const (
	rprB18KothophedOracle  = "85e1791f-f9a0-4e82-baf6-33cff2dcf60b"
	rprB18LavalancheOracle = "4475e9a7-4e94-4965-8e7e-211d40a98217"
	rprB18MaelstromOracle  = "ac50ef98-1791-4dd4-9c1f-8cea7db7ade5"
	rprB18BoskOracle       = "42b9d383-3fe2-4fc8-ab86-f80a288d502b"
	rprB18NehebOracle      = "ed8e9a2e-5fb8-4f39-b954-dbcbc96d909f"
	rprB18OdricOracle      = "bad76170-c773-4be5-9457-20dc9f745cb4"
	rprB18PathOracle       = "a933f443-dd53-4805-8446-72693a27cb56"
	rprB18PrimevalsOracle  = "8de8027b-072a-474b-b76b-dc49c417cc55"
	rprB18FlamesOracle     = "cfc8372f-13b6-4628-84c7-066881b287f5"
	rprB18RienneOracle     = "1ef2415f-f88f-41fb-b084-ea52631a60c1"
)

func TestRprB18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB18KothophedOracle:  "Kothophed, Soul Hoarder",
		rprB18LavalancheOracle: "Lavalanche",
		rprB18MaelstromOracle:  "Maelstrom Archangel",
		rprB18BoskOracle:       "Murmuring Bosk",
		rprB18NehebOracle:      "Neheb, Dreadhorde Champion",
		rprB18OdricOracle:      "Odric, Lunarch Marshal",
		rprB18PathOracle:       "Path to the World Tree",
		rprB18PrimevalsOracle:  "Primevals' Glorious Rebirth",
		rprB18FlamesOracle:     "Radiant Flames",
		rprB18RienneOracle:     "Rienne, Angel of Rebirth",
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

// rprB18Cast casts a card from hand with a floated pool, strictly.
func rprB18Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, pool string, colors []string, params game.CastSpellParams) error {
	t.Helper()
	id := effectstest.HandCardFull(me, name, typeLine, cost, oracle, colors)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, pool)
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

// --- Murmuring Bosk ---------------------------------------------------------

func TestRprB18MurmuringBoskRevealsATreefolk(t *testing.T) {
	effectstest.S12RevealLandCase(t, "Murmuring Bosk", rprB18BoskOracle, "Creature — Treefolk", "Creature — Elf")
}

func TestRprB18MurmuringBoskColouredManaDealsOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Life
	land := effectstest.SeedPermanentWithOracle(g, me.ID, "Murmuring Bosk", "Land — Forest", rprB18BoskOracle)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool %v, want one mana", me.ManaPool)
	}
	if c := me.ManaPool.Units()[0].Color; c != "W" && c != "B" {
		t.Errorf("mana is %s, want W or B", c)
	}
	if me.Life != before-1 {
		t.Errorf("life %d → %d, want one damage", before, me.Life)
	}
}

// --- Kothophed, Soul Hoarder --------------------------------------------------

func TestRprB18KothophedDrawsWhenAnotherPlayersPermanentIsPutIntoAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Kothophed, Soul Hoarder", "Legendary Creature — Demon", rprB18KothophedOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life

	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatalf("my own permanent triggered it: hand %d→%d life %d→%d", hand, me.Hand.Size(), life, me.Life)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d→%d life %d→%d, want a card and 1 life lost", hand, me.Hand.Size(), life, me.Life)
	}
}

// Ownership decides, not control: an opponent's card I control counts, and
// my card an opponent controls does not.
func TestRprB18KothophedReadsTheOwnerNotTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Kothophed, Soul Hoarder", "Legendary Creature — Demon", rprB18KothophedOracle, false)
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID})
	lent := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Lent", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, opp.ID, lent)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a card I own, controlled by an opponent, triggered it")
	}
	effectstest.KillCreature(t, g, opp.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an opponent's card under my control did not trigger it: hand %d→%d", hand, me.Hand.Size())
	}
}

// --- Primevals' Glorious Rebirth --------------------------------------------

func TestRprB18PrimevalsNeedsALegendaryCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	err := rprB18Cast(t, g, me, "Primevals' Glorious Rebirth", "Legendary Sorcery", "{5}{W}{B}", rprB18PrimevalsOracle,
		"WBWBWBW", []string{"W", "B"}, game.CastSpellParams{})
	if err == nil {
		t.Fatal("cast without a legendary creature or planeswalker")
	}
}

func TestRprB18PrimevalsReturnsEveryLegendaryPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Legend", "Legendary Creature — Human", "", false)
	king := effectstest.GraveyardCardOf(g, me, "Dead King", "Legendary Creature — Human", "{3}", "")
	land := effectstest.GraveyardCardOf(g, me, "Dead Land", "Legendary Land", "", "")
	sorc := effectstest.GraveyardCardOf(g, me, "Dead Sorcery", "Legendary Sorcery", "{1}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}", "")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Legend", "Legendary Creature — Elf", "{1}", "")
	if err := rprB18Cast(t, g, me, "Primevals' Glorious Rebirth", "Legendary Sorcery", "{5}{W}{B}", rprB18PrimevalsOracle,
		"WBWBWBW", []string{"W", "B"}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	onBF := func(id uuid.UUID) bool {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				return true
			}
		}
		return false
	}
	if !onBF(king) || !onBF(land) {
		t.Errorf("legendary permanent cards on the battlefield: creature %v land %v", onBF(king), onBF(land))
	}
	if onBF(sorc) || onBF(bear) || onBF(theirs) {
		t.Errorf("returned a non-legendary, non-permanent or opponent's card: sorcery %v bear %v theirs %v", onBF(sorc), onBF(bear), onBF(theirs))
	}
}

// --- Radiant Flames ----------------------------------------------------------

func TestRprB18RadiantFlamesDealsOneDamagePerColourSpent(t *testing.T) {
	for _, tc := range []struct {
		pool string
		want int
	}{{"RGW", 3}, {"RRG", 2}, {"RRR", 1}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wall", 0, 9)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 9)
		if err := rprB18Cast(t, g, me, "Radiant Flames", "Sorcery", "{2}{R}", rprB18FlamesOracle, tc.pool, []string{"R"}, game.CastSpellParams{}); err != nil {
			t.Fatalf("%s: cast: %v", tc.pool, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, id := range []uuid.UUID{mine, theirs} {
			if got := effectstest.DamageOn(t, g, id); got != tc.want {
				t.Errorf("%s: a creature has %d damage, want %d", tc.pool, got, tc.want)
			}
		}
	}
}

// --- Lavalanche --------------------------------------------------------------

func TestRprB18LavalancheHitsThePlayerAndTheirCreaturesForX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wall", 0, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 9)
	third := effectstest.B12Creature(g, other.ID, "Third", "Creature — Wall", 0, 9)
	life := opp.Life
	if err := rprB18Cast(t, g, me, "Lavalanche", "Sorcery", "{X}{B}{R}{G}", rprB18LavalancheOracle, "BRGCCC",
		[]string{"B", "R", "G"}, game.CastSpellParams{XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want 3 damage", life, opp.Life)
	}
	if got := effectstest.DamageOn(t, g, theirs); got != 3 {
		t.Errorf("their creature has %d damage, want 3", got)
	}
	if effectstest.DamageOn(t, g, mine) != 0 || effectstest.DamageOn(t, g, third) != 0 {
		t.Errorf("another player's creatures were damaged")
	}
}

func TestRprB18LavalancheAtAPlaneswalkerHitsItsControllersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 9}, Owner: opp.ID, Controller: opp.ID})
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wall", 0, 9)
	life := opp.Life
	if err := rprB18Cast(t, g, me, "Lavalanche", "Sorcery", "{X}{B}{R}{G}", rprB18LavalancheOracle, "BRGCC",
		[]string{"B", "R", "G"}, game.CastSpellParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, theirs); got != 2 {
		t.Errorf("the planeswalker controller's creature has %d damage, want 2", got)
	}
	if opp.Life != life {
		t.Errorf("the player took damage aimed at the planeswalker: %d → %d", life, opp.Life)
	}
}

// --- Rienne, Angel of Rebirth ---------------------------------------------

func TestRprB18RienneAnthemsOtherMulticolouredCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rienne := effectstest.PushCatalogPermanent(g, me.ID, "Rienne, Angel of Rebirth", "Legendary Creature — Angel", rprB18RienneOracle, false)
	gold := effectstest.B16Creature(g, me.ID, "Gold", "Creature — Bear", 2, 2, "R", "G")
	mono := effectstest.B16Creature(g, me.ID, "Mono", "Creature — Bear", 2, 2, "R")
	theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "R", "G")
	effectstest.AdvanceToMain(t, g)
	if got := effectstest.EffectivePower(t, g, gold); got != 3 {
		t.Errorf("multicoloured creature power %d, want 3", got)
	}
	if effectstest.EffectivePower(t, g, mono) != 2 || effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Errorf("a monocoloured or opposing creature was pumped")
	}
	_ = rienne
}

func TestRprB18RienneReturnsADeadMulticolouredCreatureAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Rienne, Angel of Rebirth", "Legendary Creature — Angel", rprB18RienneOracle, false)
	gold := effectstest.B16Creature(g, me.ID, "Gold", "Creature — Bear", 2, 2, "R", "G")
	mono := effectstest.B16Creature(g, me.ID, "Mono", "Creature — Bear", 2, 2, "R")
	theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "R", "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.KillCreature(t, g, me.ID, mono)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.KillCreature(t, g, me.ID, gold)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(gold) {
		t.Fatal("returned before the end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gold) {
		t.Error("the multicoloured creature did not return to hand at the next end step")
	}
	if me.Hand.Contains(mono) || opp.Hand.Contains(theirs) {
		t.Error("a monocoloured or opposing creature returned")
	}
}

// --- Odric, Lunarch Marshal -----------------------------------------------

func rprB18KeywordCreature(g *game.Game, owner uuid.UUID, name string, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: keywords, Owner: owner, Controller: owner})
}

func TestRprB18OdricSharesKeywordsAtTheBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Odric, Lunarch Marshal", "Legendary Creature — Human Soldier", rprB18OdricOracle, false)
	flyer := rprB18KeywordCreature(g, me.ID, "Flyer", "flying")
	tramp := rprB18KeywordCreature(g, me.ID, "Trampler", "trample")
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	foe := rprB18KeywordCreature(g, opp.ID, "Foe", "deathtouch")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{flyer, tramp, plain} {
		for _, kw := range []string{"flying", "trample"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("creature lacks %s after Odric's trigger", kw)
			}
		}
		if effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("an opponent's keyword was shared")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, "vigilance") {
		t.Error("a keyword no creature has was granted")
	}
	_ = foe
}

// --- Neheb, Dreadhorde Champion -----------------------------------------------

func TestRprB18NehebDiscardsDrawsAndAddsKeptRedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	neheb := effectstest.PushCatalogPermanent(g, me.ID, "Neheb, Dreadhorde Champion", "Legendary Creature — Zombie Minotaur Warrior", rprB18NehebOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	for i := 0; i < 3; i++ {
		effectstest.HandCardFull(me, "Filler", "Sorcery", "{9}", "", nil)
	}
	effectstest.AttackWith(t, g, opp.ID, neheb)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after Neheb connected")
	}
	if c.ChooseMax != 3 {
		t.Errorf("may discard up to %d, want the whole hand of 3", c.ChooseMax)
	}
	ids := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	if err := g.ResolveChooseCards(c.ID, me.ID, ids); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if me.Hand.Size() != 3 {
		t.Errorf("hand %d, want 3 (discarded 2 of 3, drew 2)", me.Hand.Size())
	}
	red := 0
	for _, u := range me.ManaPool.Units() {
		if u.Color == "R" {
			red++
		}
	}
	if red != 2 {
		t.Errorf("pool has %d {R}, want 2", red)
	}
	// The mana survives the end of the step.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	red = 0
	for _, u := range me.ManaPool.Units() {
		if u.Color == "R" {
			red++
		}
	}
	if red != 2 {
		t.Errorf("after the steps ended the pool has %d {R}, want 2 kept", red)
	}
}

func TestRprB18NehebMayDiscardNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	neheb := effectstest.PushCatalogPermanent(g, me.ID, "Neheb, Dreadhorde Champion", "Legendary Creature — Zombie Minotaur Warrior", rprB18NehebOracle, false)
	effectstest.HandCardFull(me, "Filler", "Sorcery", "{9}", "", nil)
	effectstest.AttackWith(t, g, opp.ID, neheb)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Hand.Size() != hand || me.ManaPool.Total() != 0 {
		t.Errorf("declining changed the hand (%d → %d) or added mana (%d)", hand, me.Hand.Size(), me.ManaPool.Total())
	}
}

// --- Maelstrom Archangel -----------------------------------------------------

func TestRprB18MaelstromArchangelCastsAFreeSpellOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	angel := effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom Archangel", "Creature — Angel", rprB18MaelstromOracle, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bear := effectstest.HandCardFull(me, "Big Bear", "Creature — Bear", "{5}{G}", "", []string{"G"})
	land := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	effectstest.AttackWith(t, g, opp.ID, angel)
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, bear)
	if effectstest.HasID(offer.DiscardOptions, land) {
		t.Error("a land was offered as a spell")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	if me.Hand.Contains(bear) {
		t.Error("the spell is still in hand")
	}
}

// --- Path to the World Tree ----------------------------------------------------

func TestRprB18PathSearchesForABasicLandOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Path to the World Tree", "Enchantment", rprB18PathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not go to hand")
	}
}

func TestRprB18PathSacrificeDoesEverythingPrinted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	path := effectstest.PushCatalogPermanent(g, me.ID, "Path to the World Tree", "Enchantment", rprB18PathOracle, false)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Wall", 0, 9)
	effectstest.FloatForTest(g, me, "CCWUBRG")
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, path, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: foe, Slot: 1},
		},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || opp.Life != oppLife-2 {
		t.Errorf("life: me %d→%d opp %d→%d, want +2 and -2", life, me.Life, oppLife, opp.Life)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d→%d, want 2 cards drawn", hand, me.Hand.Size())
	}
	if got := effectstest.DamageOn(t, g, foe); got != 2 {
		t.Errorf("the creature has %d damage, want 2", got)
	}
	bears := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Bear" {
			bears++
		}
		if c.InstanceID == path {
			t.Error("the enchantment was not sacrificed")
		}
	}
	if bears != 1 {
		t.Errorf("%d Bear tokens, want 1", bears)
	}
}

func TestRprB18PathSacrificeWithNoCreatureTargetStillDoesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	path := effectstest.PushCatalogPermanent(g, me.ID, "Path to the World Tree", "Enchantment", rprB18PathOracle, false)
	effectstest.FloatForTest(g, me, "CCWUBRG")
	oppLife := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, path, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d→%d, want 2 lost", oppLife, opp.Life)
	}
}
