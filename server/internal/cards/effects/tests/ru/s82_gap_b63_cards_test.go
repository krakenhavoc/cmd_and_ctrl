package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b63_cards_test.go — GAP batch 63 (#5948, S82): Fallout
// creatures. Each test plays the real catalog entry.

const (
	gapb63MegaslothOracle  = "6492c381-0fa1-4d43-bf72-d8956d8a34b6"
	gapb63YaoGuaiOracle    = "d37e8f75-c7cb-4270-bb2f-0125e972ed15"
	gapb63CaitOracle       = "f24319a1-0662-44be-9027-11e7bc020e1d"
	gapb63DogmeatOracle    = "5232fcc0-5dba-4569-8e99-c0b2c230e9f0"
	gapb63PrestonOracle    = "08a0c661-5ed5-47a6-aa1d-3114bf5db3fb"
	gapb63MarcusOracle     = "c9db8f36-2c72-403e-b2b7-7cdfba5feaa4"
	gapb63CaravaneerOracle = "ab3360e5-6a60-4806-bbc2-b7b26c119856"
	gapb63DuchessOracle    = "b47d91a1-f2ac-4604-95a7-c247ac2fab18"
	gapb63IanOracle        = "0727ce6a-aae2-4ddc-9e6c-b18011c5bd37"
	gapb63StrangerOracle   = "4cd2e1c2-5f14-4ac9-9c88-899af6ad9f80"
	gapb63RoseOracle       = "5a721c27-3e02-4f60-995e-b29d678d93b3"
	gapb63MotherlodeOracle = "5ae50388-7e36-4a83-a618-0148944c5099"
)

func TestGapb63CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapb63MegaslothOracle:  "Lumbering Megasloth",
		gapb63YaoGuaiOracle:    "Rampaging Yao Guai",
		gapb63CaitOracle:       "Cait, Cage Brawler",
		gapb63DogmeatOracle:    "Dogmeat, Ever Loyal",
		gapb63PrestonOracle:    "Preston Garvey, Minuteman",
		gapb63MarcusOracle:     "Marcus, Mutant Mayor",
		gapb63CaravaneerOracle: "Crimson Caravaneer",
		gapb63DuchessOracle:    "Duchess, Wayward Tavernkeep",
		gapb63IanOracle:        "Ian the Reckless",
		gapb63StrangerOracle:   "Mysterious Stranger",
		gapb63RoseOracle:       "Rose, Cutthroat Raider",
		gapb63MotherlodeOracle: "The Motherlode, Excavator",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func gapb63Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func gapb63Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// gapb63Settled re-enables every permanent's summoning-sick flag off, so a
// pushed creature can attack or tap.
func gapb63Ready(g *game.Game) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
}

// --- Crimson Caravaneer ------------------------------------------------------------

func TestGapb63CaravaneerMakesAJunkPerCombatDamageStepToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	c := gapb63Creature(g, me.ID, "Crimson Caravaneer", "Creature — Human Scout", gapb63CaravaneerOracle, 1, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == c {
				g.Battlefield.Cards[i].Keywords = []string{"double strike", "trample"}
			}
		}
	})
	gapb63Ready(g)
	effectstest.Pr7bAttack(t, g, opp.ID, c)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb63Named(g, me.ID, "Junk")); n < 1 {
		t.Errorf("%d Junk tokens after combat damage to a player, want at least 1", n)
	}
}

// --- Marcus, Mutant Mayor -----------------------------------------------------------

func TestGapb63MarcusGivesACounterThenDrawsForAnAlreadyCounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb63Creature(g, me.ID, "Marcus, Mutant Mayor", "Legendary Creature — Mutant Advisor", gapb63MarcusOracle, 4, 4)
	bear := gapb63Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	gapb63Ready(g)
	hand := me.Hand.Size()
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("a creature with no counter has %d after connecting, want 1", n)
	}
	if me.Hand.Size() != hand {
		t.Error("drew a card for a creature that had no counter")
	}
}

func TestGapb63MarcusDrawsInsteadWhenTheCreatureAlreadyHasACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb63Creature(g, me.ID, "Marcus, Mutant Mayor", "Legendary Creature — Mutant Advisor", gapb63MarcusOracle, 4, 4)
	bear := gapb63Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if err := g.AddCounter(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	gapb63Ready(g)
	hand := me.Hand.Size()
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("a countered creature got another: %d", n)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// --- Duchess, Wayward Tavernkeep ----------------------------------------------------

func TestGapb63DuchessQuestCountersPayForJunk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duchess := gapb63Creature(g, me.ID, "Duchess, Wayward Tavernkeep", "Legendary Creature — Human Citizen", gapb63DuchessOracle, 4, 3)
	bear := gapb63Creature(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	gapb63Ready(g)
	effectstest.Pr7bAttack(t, g, opp.ID, bear)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, "quest"); n != 1 {
		t.Fatalf("the creature has %d quest counters, want 1", n)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, duchess, 0, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, "quest"); n != 0 {
		t.Errorf("the quest counter was not removed: %d", n)
	}
	if n := len(gapb63Named(g, me.ID, "Junk")); n != 1 {
		t.Errorf("%d Junk tokens, want 1", n)
	}
}

// --- Lumbering Megasloth ------------------------------------------------------------

func TestGapb63MegaslothCostsOneLessForEachCounterAmongPlayersAndPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := gapb63Creature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := g.AddCounter(bear, game.CounterPlusOne, 4); err != nil {
		t.Fatal(err)
	}
	if err := g.AddPlayerCounter(me.ID, game.CounterEnergy, 3); err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lumbering Megasloth", TypeLine: "Creature — Sloth Mutant",
		OracleID: gapb63MegaslothOracle, ManaCost: "{10}{G}{G}", Power: 8, Toughness: 8, Keywords: []string{"trample"},
		Owner: me.ID, Controller: me.ID})
	// 12 total - 7 counters = {5}{G}{G}: seven mana.
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with six mana when seven are needed")
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {5}{G}{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Megasloth did not resolve")
	}
	if !effectstest.B12Card(t, g, id).Tapped {
		t.Error("Megasloth did not enter tapped")
	}
}

// --- Rampaging Yao Guai -------------------------------------------------------------

func gapb63CastYaoGuai(t *testing.T, g *game.Game, x int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rampaging Yao Guai", TypeLine: "Creature — Bear Mutant",
		OracleID: gapb63YaoGuaiOracle, ManaCost: "{X}{G}{G}{G}", Power: 2, Toughness: 2,
		Keywords: []string{"vigilance", "trample"}, Owner: me.ID, Controller: me.ID})
	mana := "{G}{G}{G}"
	for i := 0; i < x; i++ {
		mana += "{C}"
	}
	effectstest.FloatMana(t, g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Strict: true}); err != nil {
		t.Fatalf("cast X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func gapb63Permanent(g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: owner, Controller: owner,
	})
}

func TestGapb63YaoGuaiEntersWithXCountersAndDestroysAnyNumberWithinTotalManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	one := gapb63Permanent(g, opp.ID, "One Rock", "Artifact", "{1}")
	two := gapb63Permanent(g, opp.ID, "Two Aura", "Enchantment — Aura", "{1}{U}")
	free := gapb63Permanent(g, opp.ID, "Free Token", "Token Artifact — Treasure", "")
	big := gapb63Permanent(g, opp.ID, "Big Rock", "Artifact", "{3}{C}")
	yao := gapb63CastYaoGuai(t, g, 3)
	if n := effectstest.CountersOn(g, yao, game.CounterPlusOne); n != 3 {
		t.Fatalf("enters with %d +1/+1 counters, want X = 3", n)
	}
	p := effectstest.LatestPickTarget(g, g.Seats[0].ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	// Total 1 + 2 + 0 = 3: fine. Adding Big Rock (4) is over X.
	if err := g.ResolvePickTargets(p.ID, g.Seats[0].ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: two}, {Kind: game.TargetCard, ID: big},
	}); err == nil {
		t.Fatal("a set with total mana value 7 was accepted for X = 3")
	}
	if err := g.ResolvePickTargets(p.ID, g.Seats[0].ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: two}, {Kind: game.TargetCard, ID: free},
	}); err != nil {
		t.Fatalf("a set with total mana value 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"One Rock": one, "Two Aura": two, "Free Token": free} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the untargeted artifact was destroyed")
	}
}

// --- Cait, Cage Brawler -------------------------------------------------------------

func TestGapb63CaitIsIndestructibleOnlyOnItsControllersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cait := gapb63Creature(g, me.ID, "Cait, Cage Brawler", "Legendary Creature — Human Warrior", gapb63CaitOracle, 1, 1)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, cait), "indestructible") {
		t.Error("not indestructible on my turn")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, cait), "indestructible") {
		t.Error("indestructible on an opponent's turn")
	}
}

func TestGapb63CaitBothDrawDiscardAndGetsCountersForTheGreatestManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cait := gapb63Creature(g, me.ID, "Cait, Cage Brawler", "Legendary Creature — Human Warrior", gapb63CaitOracle, 1, 1)
	gapb63Ready(g)
	// Hands: mine holds one big card, theirs one small one.
	me.Hand.Cards = nil
	opp.Hand.Cards = nil
	big := uuid.New()
	small := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})
	opp.Hand.PushTop(game.Card{InstanceID: small, Name: "Small Spell", TypeLine: "Instant", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	myHand, theirHand := me.Hand.Size(), opp.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, cait)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		progressed := false
		for _, p := range []*game.Player{me, opp} {
			if c := effectstest.SacrificeChoiceFor(g, p.ID); c != nil {
				_ = c
			}
			if pc := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID); pc != nil {
				pick := pc.ChooseCards[0]
				for _, id := range pc.ChooseCards {
					if (p == me && id == big) || (p == opp && id == small) {
						pick = id
					}
				}
				if err := g.ResolveChooseCards(pc.ID, p.ID, []uuid.UUID{pick}); err != nil {
					t.Fatalf("discard pick: %v", err)
				}
				progressed = true
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !progressed {
			break
		}
	}
	if me.Hand.Size() != myHand || opp.Hand.Size() != theirHand {
		t.Errorf("hands %d→%d and %d→%d, want each to draw one and discard one", myHand, me.Hand.Size(), theirHand, opp.Hand.Size())
	}
	if !me.Graveyard.Contains(big) || !opp.Graveyard.Contains(small) {
		t.Fatal("the chosen cards were not discarded")
	}
	if n := effectstest.CountersOn(g, cait, game.CounterPlusOne); n != 2 {
		t.Errorf("Cait has %d counters after discarding the greatest mana value, want 2", n)
	}
}

func TestGapb63CaitGetsNoCountersWhenTheDefendersCardIsBigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cait := gapb63Creature(g, me.ID, "Cait, Cage Brawler", "Legendary Creature — Human Warrior", gapb63CaitOracle, 1, 1)
	gapb63Ready(g)
	me.Hand.Cards = nil
	opp.Hand.Cards = nil
	small := uuid.New()
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: small, Name: "Small Spell", TypeLine: "Instant", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	opp.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{5}", Owner: opp.ID, Controller: opp.ID})
	effectstest.DeclareAttack(t, g, opp.ID, cait)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		progressed := false
		for _, p := range []*game.Player{me, opp} {
			if pc := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID); pc != nil {
				pick := pc.ChooseCards[0]
				for _, id := range pc.ChooseCards {
					if (p == me && id == small) || (p == opp && id == big) {
						pick = id
					}
				}
				if err := g.ResolveChooseCards(pc.ID, p.ID, []uuid.UUID{pick}); err != nil {
					t.Fatalf("discard pick: %v", err)
				}
				progressed = true
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !progressed {
			break
		}
	}
	if n := effectstest.CountersOn(g, cait, game.CounterPlusOne); n != 0 {
		t.Errorf("Cait got %d counters though the defender discarded the greater mana value", n)
	}
}

// --- Dogmeat, Ever Loyal ------------------------------------------------------------

func TestGapb63DogmeatMillsFiveThenReturnsAnAuraOrEquipmentToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sword := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sword, Name: "Old Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	lib := me.Library.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dogmeat, Ever Loyal", TypeLine: "Legendary Creature — Dog",
		OracleID: gapb63DogmeatOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{R}{G}{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != lib-5 {
		t.Errorf("library %d → %d, want 5 milled", lib, me.Library.Size())
	}
	pc := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pc == nil {
		t.Fatal("no pick of an Aura or Equipment card")
	}
	if effectstest.HasID(pc.ChooseCards, bear) {
		t.Error("a creature card was offered")
	}
	if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{sword}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == sword {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the Equipment did not return to hand")
	}
}

func TestGapb63DogmeatMakesJunkForAnEnchantedOrEquippedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb63Creature(g, me.ID, "Dogmeat, Ever Loyal", "Legendary Creature — Dog", gapb63DogmeatOracle, 3, 3)
	bare := gapb63Creature(g, me.ID, "Bare", "Creature — Bear", "", 2, 2)
	armed := gapb63Creature(g, me.ID, "Armed", "Creature — Bear", "", 2, 2)
	sword := gapb63Creature(g, me.ID, "Sword", "Artifact — Equipment", "", 0, 0)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: armed}); err != nil {
			t.Fatal(err)
		}
	})
	gapb63Ready(g)
	effectstest.Pr7bAttack(t, g, opp.ID, bare, armed)
	if n := len(gapb63Named(g, me.ID, "Junk")); n != 1 {
		t.Errorf("%d Junk tokens, want 1 (only the equipped attacker)", n)
	}
}

// --- Preston Garvey, Minuteman ------------------------------------------------------

func TestGapb63PrestonEnchantsALandWithASettlementThatTapsForAnyColorAndUntapsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	preston := gapb63Creature(g, me.ID, "Preston Garvey, Minuteman", "Legendary Creature — Human Soldier", gapb63PrestonOracle, 4, 4)
	land := gapb63Creature(g, me.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	gapb63Ready(g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	settlements := gapb63Named(g, me.ID, "Settlement")
	if len(settlements) != 1 {
		t.Fatalf("%d Settlements, want 1", len(settlements))
	}
	var host uuid.UUID
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(settlements[0])
		host = c.AttachedTo.ID
	})
	if host != land {
		t.Fatalf("the Settlement is attached to %v, want the land", host)
	}
	var mana int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(land)
		mana = len(game.ManaAbilitiesForCard(c))
	})
	if mana == 0 {
		t.Error("the enchanted land has no mana ability")
	}
	// Tap the enchanted land, attack with Preston: it untaps.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, preston)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, land).Tapped {
		t.Error("the enchanted land stayed tapped after Preston attacked")
	}
}

func TestGapb63PrestonWithNoLandTargetMakesNoLastingSettlement(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb63Creature(g, me.ID, "Preston Garvey, Minuteman", "Legendary Creature — Human Soldier", gapb63PrestonOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb63Named(g, me.ID, "Settlement")); n != 0 {
		t.Errorf("%d unattached Settlements stayed on the battlefield", n)
	}
}

// --- Ian the Reckless ---------------------------------------------------------------

func TestGapb63IanDamagesYouAndATargetOnlyWhenModified(t *testing.T) {
	for _, modified := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ian := gapb63Creature(g, me.ID, "Ian the Reckless", "Legendary Creature — Human Warrior", gapb63IanOracle, 2, 1)
		if modified {
			if err := g.AddCounter(ian, game.CounterPlusOne, 1); err != nil {
				t.Fatal(err)
			}
		}
		gapb63Ready(g)
		life, oppLife := me.Life, opp.Life
		effectstest.DeclareAttack(t, g, opp.ID, ian)
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if !modified {
				t.Fatal("an unmodified Ian asked for a target")
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[2].ID}); err != nil {
				t.Fatal(err)
			}
		} else if modified {
			t.Fatal("a modified Ian asked for no target")
		}
		if modified {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if modified {
			if me.Life != life-3 || g.Seats[2].Life != 40-3 {
				t.Errorf("modified Ian (3 power): my life %d→%d, target's %d, want both to take 3", life, me.Life, g.Seats[2].Life)
			}
		} else if me.Life != life || opp.Life != oppLife {
			t.Error("an unmodified Ian dealt damage")
		}
	}
}

// --- Rose, Cutthroat Raider ---------------------------------------------------------

func TestGapb63RoseMakesAJunkForEachOpponentAttackedAndMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapb63Creature(g, me.ID, "Rose, Cutthroat Raider", "Legendary Artifact Creature — Robot", gapb63RoseOracle, 3, 2)
	a := gapb63Creature(g, me.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := gapb63Creature(g, me.ID, "Bear B", "Creature — Bear", "", 2, 2)
	c := gapb63Creature(g, me.ID, "Bear C", "Creature — Bear", "", 2, 2)
	gapb63Ready(g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, pair := range []struct{ id, def uuid.UUID }{{a, opp.ID}, {b, opp.ID}, {c, g.Seats[2].ID}} {
		if err := g.DeclareAttacker(pair.id, pair.def); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	junk := gapb63Named(g, me.ID, "Junk")
	if len(junk) != 2 {
		t.Fatalf("%d Junk tokens after attacking two opponents with three creatures, want 2", len(junk))
	}
	// Sacrificing a Junk adds {R}.
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	before := me.ManaPool.Total()
	if err := g.ActivateCatalogAbility(me.ID, junk[0], 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Junk: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total() - before; got != 1 {
		t.Errorf("mana gained %d from sacrificing a Junk, want 1", got)
	}
}

func TestGapb63RoseMakesNothingIfYouDidNotAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapb63Creature(g, me.ID, "Rose, Cutthroat Raider", "Legendary Artifact Creature — Robot", gapb63RoseOracle, 3, 2)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gapb63Named(g, me.ID, "Junk")); n != 0 {
		t.Errorf("%d Junk without an attack", n)
	}
}

// --- The Motherlode, Excavator ------------------------------------------------------

func TestGapb63MotherlodeGetsEnergyPerNonbasicLandOfTheChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gapb63Permanent(g, opp.ID, "Nonbasic A", "Land", "")
	gapb63Permanent(g, opp.ID, "Nonbasic B", "Land", "")
	gapb63Permanent(g, opp.ID, "Plains", "Basic Land — Plains", "")
	gapb63Permanent(g, g.Seats[2].ID, "Other Nonbasic", "Land", "")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "The Motherlode, Excavator", TypeLine: "Legendary Artifact Creature — Robot",
		OracleID: gapb63MotherlodeOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters[game.CounterEnergy]; got != 2 {
		t.Errorf("%d energy, want 2 (two nonbasic lands)", got)
	}
}

func TestGapb63MotherlodeAttackPaysFourEnergyToDestroyALandAndStopGroundBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lode := gapb63Creature(g, me.ID, "The Motherlode, Excavator", "Legendary Artifact Creature — Robot", gapb63MotherlodeOracle, 5, 5)
	nonbasic := gapb63Permanent(g, opp.ID, "Nonbasic", "Land", "")
	basic := gapb63Permanent(g, opp.ID, "Plains", "Basic Land — Plains", "")
	ground := gapb63Creature(g, opp.ID, "Ground", "Creature — Bear", "", 2, 2)
	flier := gapb63Creature(g, opp.ID, "Flier", "Creature — Bird", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	effectstest.SetEnergy(t, g, me, 4)
	gapb63Ready(g)
	effectstest.AttackAndAnswerEnergy(t, g, lode, true)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, basic) {
			t.Error("a basic land was offered")
		}
		effectstest.PickCard(t, g, me.ID, nonbasic)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nonbasic) {
		t.Error("the nonbasic land survived")
	}
	if !g.Battlefield.Contains(basic) {
		t.Error("the basic land was destroyed")
	}
	if me.Counters[game.CounterEnergy] != 0 {
		t.Errorf("%d energy left after paying four", me.Counters[game.CounterEnergy])
	}
	canBlock := func(id uuid.UUID) bool {
		var r game.Restriction
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(id)
			r = c.Effective().Restrictions
		})
		return r&game.CantBlock == 0
	}
	if canBlock(ground) {
		t.Error("the ground creature can still block")
	}
	if !canBlock(flier) {
		t.Error("the flier was stopped from blocking")
	}
}

// --- Mysterious Stranger ------------------------------------------------------------

func TestGapb63StrangerExilesOneInstantOrSorceryPerGraveyardAndCopiesOneOfThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := uuid.New()
	theirs := uuid.New()
	extra := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: extra, Name: "My Other Spell", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Spell", TypeLine: "Sorcery", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mysterious Stranger", TypeLine: "Creature — Human Rogue",
		OracleID: gapb63StrangerOracle, Keywords: []string{"flash"}, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no targets asked for")
	}
	// Two cards from one graveyard are refused: one per graveyard.
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: extra},
	}); err == nil {
		t.Fatal("two cards from the same graveyard were accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	}); err != nil {
		t.Fatalf("one per graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Fatal("the targeted cards were not exiled")
	}
	if !me.Graveyard.Contains(extra) {
		t.Error("the untargeted spell left the graveyard")
	}
	// Two were exiled, so one of them is copied and offered for a free cast.
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID) == nil {
		t.Error("no free-cast offer for the copy of the randomly chosen card")
	}
}

func TestGapb63StrangerWithNoInstantOrSorceryInAnyGraveyardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mysterious Stranger", TypeLine: "Creature — Human Rogue",
		OracleID: gapb63StrangerOracle, Keywords: []string{"flash"}, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.FloatMana(t, g, me, "{C}{C}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("asked for a target with no instant or sorcery card in any graveyard")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("Stranger did not enter")
	}
}
