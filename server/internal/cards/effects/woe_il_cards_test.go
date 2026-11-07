package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_il_cards_test.go — Wilds of Eldraine, the I–L batch (#4471). Each
// test plays the real catalog entry and asserts what its text says.

const (
	woeILIceOutOracle     = "fed3bc35-4269-48f0-9133-a595ce49c2d1"
	woeILRecruiterOracle  = imodanesRecruiterOracleID
	woeILFaeCourtOracle   = "8db0a379-4cf2-43a7-932b-c5646a604578"
	woeILTruffleOracle    = intrepidTrufflesnoutOracleID
	woeILStopgapOracle    = "c7b235e1-746e-4467-be54-4e11e2bf6362"
	woeILLightbladeOracle = "4b803bf2-025a-4bf6-9d3c-a28d714b4781"
	woeILKellanOracle     = kellanTheFaeBloodedOracleID
	woeILHeroismOracle    = "c1ed20c0-f75d-4aca-8874-6955bd9fb21b"
	woeILDovesOracle      = "d2e6fbcb-ebbd-40e8-8a01-b675fcacaf8e"
	woeILKorvoldOracle    = "bcae7f20-b601-4669-b3a7-c32ed0c6b9d3"
	woeILLadyOracle       = "356a85d9-2570-448c-a455-f4ce915a604f"
	woeILAmbushOracle     = "e920c955-66f5-4c6e-a2a8-8b282f9e456e"
	woeILButcherOracle    = "467f756f-75db-4748-99d3-62651776adfe"
	woeILSkitterOracle    = "f5bed085-bb7a-4250-9599-b2a8a95e6b6f"
)

func TestWoeILEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeILIceOutOracle:           "Ice Out",
		woeILRecruiterOracle:        "Imodane's Recruiter",
		woeILRecruiterOracle + "#1": "Train Troops",
		woeILFaeCourtOracle:         "Into the Fae Court",
		woeILTruffleOracle:          "Intrepid Trufflesnout",
		woeILTruffleOracle + "#1":   "Go Hog Wild",
		woeILStopgapOracle:          "Johann's Stopgap",
		woeILLightbladeOracle:       "Kellan's Lightblades",
		woeILKellanOracle:           "Kellan, the Fae-Blooded",
		woeILKellanOracle + "#1":    "Birthright Boon",
		woeILHeroismOracle:          "Kindled Heroism",
		woeILDovesOracle:            "Knight of Doves",
		woeILKorvoldOracle:          "Korvold and the Noble Thief",
		woeILLadyOracle:             "Lady of Laughter",
		woeILAmbushOracle:           "Leaping Ambush",
		woeILButcherOracle:          "Lord Skitter's Butcher",
		woeILSkitterOracle:          "Lord Skitter, Sewer King",
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

// --- fixtures ------------------------------------------------------------------

// woeILToken puts an artifact token under owner's control: a permanent
// bargain can sacrifice.
func woeILToken(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: owner, Controller: owner,
	})
}

// woeILPrice is what a card's spell costs from hand with the given
// optional costs announced, after its own cost modifiers.
func woeILPrice(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, cost, oracle string, optional []int) int {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: owner, Controller: owner}
	base, err := game.ParseCost(cost)
	if err != nil {
		t.Fatal(err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: owner, FromZone: game.ZoneHand, OptionalCosts: optional})
	if err != nil {
		t.Fatal(err)
	}
	return out.ManaValue()
}

// woeILAttacking marks a creature as attacking `defender`, the
// designation an "attacking or blocking" clause reads.
func woeILAttacking(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].AttackingTarget = defender
			}
		}
	})
}

// woeILBlockRefused reports whether blocker may not block attacker.
func woeILBlockRefused(g *game.Game, attacker, blocker uuid.UUID) bool {
	var r game.BlockRefusal
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		var a, b *game.Card
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case attacker:
				a = &g.Battlefield.Cards[i]
			case blocker:
				b = &g.Battlefield.Cards[i]
			}
		}
		r = g.BlockPairRefusalLocked(a, b)
	})
	return r.Reason != ""
}

// woeILPickMode answers the open mode prompt with one bullet.
func woeILPickMode(t *testing.T, g *game.Game, chooser uuid.UUID, mode int) {
	t.Helper()
	pick := modePickChoiceFor(g, chooser)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

// --- Ice Out ---------------------------------------------------------------------

func TestWoeILIceOutCountersASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := me.Life
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	castCatalogSpell(t, g, "Ice Out", "Instant", woeILIceOutOracle, b16TargetCard(bolt))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || me.Life != life {
		t.Errorf("the Bolt was not countered: in graveyard %v, life %d -> %d", opp.Graveyard.Contains(bolt), life, me.Life)
	}
}

func TestWoeILIceOutCostsOneLessBargained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := woeILPrice(t, g, me.ID, "Ice Out", "Instant", "{1}{U}{U}", woeILIceOutOracle, nil); got != 3 {
		t.Errorf("unbargained price %d, want 3", got)
	}
	if got := woeILPrice(t, g, me.ID, "Ice Out", "Instant", "{1}{U}{U}", woeILIceOutOracle, []int{0}); got != 2 {
		t.Errorf("bargained price %d, want 2", got)
	}
}

func TestWoeILIceOutBargainedIsCastForTwoBlueAndSacrificesTheToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	token := woeILToken(g, me.ID)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))
	floatForTest(g, me, "UU")
	p1g2Cast(t, g, "Ice Out", "Instant", woeILIceOutOracle, "{1}{U}{U}", game.CastSpellParams{
		Strict: true, Targets: b16TargetCard(bolt), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the Bolt was not countered")
	}
}

// --- Imodane's Recruiter // Train Troops ------------------------------------------

func woeILRecruiter(p *game.Player) game.Card {
	return woeB1Adventure(p, woeILRecruiterOracle,
		game.Face{Name: "Imodane's Recruiter", TypeLine: "Creature — Human Knight", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 2},
		game.Face{Name: "Train Troops", TypeLine: "Sorcery — Adventure", ManaCost: "{4}{W}", Colors: []string{"W"}})
}

func TestWoeILImodanesRecruiterPumpsAndHastesYourCreaturesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	card := woeILRecruiter(me)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Fatal("Imodane's Recruiter did not resolve")
	}
	for name, id := range map[string]uuid.UUID{"your other creature": bear, "the Recruiter": card.InstanceID} {
		if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 2 {
			t.Errorf("%s is %d/%d, want 3/2", name, p, tough)
		}
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", name)
		}
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Errorf("an opponent's creature is %d/%d (haste %v), want an untouched 2/2", p, tough, hasEffectiveKeyword(t, g, theirs, "haste"))
	}
	endTurn(t, g)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

func TestWoeILTrainTroopsMakesTwoVigilantKnightsAndGoesOnAnAdventure(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	woeB1CastAdventure(t, g, me, woeILRecruiter(me), nil)
	if got := b27CountNamed(g, me.ID, "Knight"); got != 2 {
		t.Fatalf("%d Knight tokens, want 2", got)
	}
	knight := findBattlefieldByName(g, "Knight")
	if c, _ := battlefieldCard(g, knight); c.Power != 2 || c.Toughness != 2 {
		t.Errorf("the Knight is %d/%d, want 2/2", c.Power, c.Toughness)
	}
	assertKeywords(t, g, knight, "vigilance")
}

// --- Into the Fae Court ----------------------------------------------------------

func TestWoeILIntoTheFaeCourtDrawsThreeAndMakesAFaerieThatBlocksOnlyFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	hand := handSize(me)
	castCatalogSpell(t, g, "Into the Fae Court", "Sorcery", woeILFaeCourtOracle, nil)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+3 {
		t.Errorf("hand %d, want %d", got, hand+3)
	}
	if got := b27CountNamed(g, me.ID, "Faerie"); got != 1 {
		t.Fatalf("%d Faerie tokens, want 1", got)
	}
	faerie := findBattlefieldByName(g, "Faerie")
	if c, _ := battlefieldCard(g, faerie); c.Power != 1 || c.Toughness != 1 {
		t.Errorf("the Faerie is %d/%d, want 1/1", c.Power, c.Toughness)
	}
	assertKeywords(t, g, faerie, "flying")
	ground := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	if !woeILBlockRefused(g, ground, faerie) {
		t.Error("the Faerie could block a creature without flying")
	}
	if woeILBlockRefused(g, flier, faerie) {
		t.Error("the Faerie could not block a creature with flying")
	}
}

// --- Intrepid Trufflesnout // Go Hog Wild ----------------------------------------

func TestWoeILTrufflesnoutMakesAFoodOnlyWhenItAttacksAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	snout := b12Push(g, me.ID, "Intrepid Trufflesnout", "Creature — Boar", woeILTruffleOracle, 3, 1)
	declareAttack(t, g, opp.ID, snout)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after attacking alone, want 1", got)
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	snout2 := b12Push(g2, me2.ID, "Intrepid Trufflesnout", "Creature — Boar", woeILTruffleOracle, 3, 1)
	bear := b12Creature(g2, me2.ID, "My Bear", woeMPBearType, 2, 2)
	declareAttack(t, g2, opp2.ID, snout2, bear)
	passPriorityAroundTable(t, g2)
	if got := b27CountNamed(g2, me2.ID, "Food"); got != 0 {
		t.Errorf("%d Food after attacking beside another creature, want 0", got)
	}
}

func TestWoeILGoHogWildPumpsTwoTwoAndGoesOnAnAdventure(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	card := woeB1Adventure(me, woeILTruffleOracle,
		game.Face{Name: "Intrepid Trufflesnout", TypeLine: "Creature — Boar", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 3, Toughness: 1},
		game.Face{Name: "Go Hog Wild", TypeLine: "Instant — Adventure", ManaCost: "{1}{G}", Colors: []string{"G"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(bear))
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	endTurn(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the pump outlasted the turn: %d/%d", p, tough)
	}
}

// --- Johann's Stopgap ------------------------------------------------------------

func TestWoeILJohannsStopgapBouncesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	hand := handSize(me)
	castCatalogSpell(t, g, "Johann's Stopgap", "Sorcery", woeILStopgapOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the permanent was not returned to its owner's hand")
	}
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d after the draw, want %d", got, hand+1)
	}
}

func TestWoeILJohannsStopgapRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := woeB1Seats(g)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Johann's Stopgap", "Sorcery", woeILStopgapOracle, cardRefs(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
}

func TestWoeILJohannsStopgapCostsTwoLessBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	if got := woeILPrice(t, g, me.ID, "Johann's Stopgap", "Sorcery", "{3}{U}", woeILStopgapOracle, nil); got != 4 {
		t.Errorf("unbargained price %d, want 4", got)
	}
	if got := woeILPrice(t, g, me.ID, "Johann's Stopgap", "Sorcery", "{3}{U}", woeILStopgapOracle, []int{0}); got != 2 {
		t.Errorf("bargained price %d, want 2", got)
	}
	seedLibrary(me, "A", "B")
	token := woeILToken(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	floatForTest(g, me, "UC")
	p1g2Cast(t, g, "Johann's Stopgap", "Sorcery", woeILStopgapOracle, "{3}{U}", game.CastSpellParams{
		Strict: true, Targets: cardRefs(theirs), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("the bargained Stopgap did not bounce")
	}
}

// --- Kellan's Lightblades --------------------------------------------------------

func TestWoeILKellansLightbladesDealsThreeToAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	attacker := b12Creature(g, opp.ID, "Their Giant", woeMPBearType, 4, 4)
	if err := castCatalogSpellErr(t, g, "Kellan's Lightblades", "Instant", woeILLightbladeOracle, cardRefs(attacker)); err == nil {
		t.Fatal("Kellan's Lightblades targeted a creature that is neither attacking nor blocking")
	}
	woeILAttacking(g, attacker, me.ID)
	castCatalogSpell(t, g, "Kellan's Lightblades", "Instant", woeILLightbladeOracle, cardRefs(attacker))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, attacker)
	if !ok {
		t.Fatal("an unbargained Lightblades destroyed a 4/4")
	}
	if c.DamageMarked != 3 {
		t.Errorf("%d damage marked, want 3", c.DamageMarked)
	}
}

func TestWoeILKellansLightbladesBargainedDestroysInstead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	token := woeILToken(g, me.ID)
	attacker := b12Creature(g, opp.ID, "Their Giant", woeMPBearType, 6, 6)
	woeILAttacking(g, attacker, me.ID)
	p1g2Cast(t, g, "Kellan's Lightblades", "Instant", woeILLightbladeOracle, "", game.CastSpellParams{
		Targets: cardRefs(attacker), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) || !opp.Graveyard.Contains(attacker) {
		t.Error("a bargained Lightblades did not destroy the 6/6")
	}
}

// --- Kellan, the Fae-Blooded // Birthright Boon ----------------------------------

func TestWoeILKellanPumpsOthersForEachAuraAndEquipmentOnHim(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	kellan := b12Push(g, me.ID, "Kellan, the Fae-Blooded", "Legendary Creature — Human Faerie", woeILKellanOracle, 2, 2)
	assertKeywords(t, g, kellan, "double strike")
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	if p, _ := effectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("with nothing attached your creature has power %d, want 2", p)
	}
	aura := b12Permanent(g, me.ID, "Some Aura", "Enchantment — Aura")
	woeB1Attach(t, g, aura, kellan)
	sword := b12Permanent(g, me.ID, "Some Sword", "Artifact — Equipment")
	woeB1Attach(t, g, sword, kellan)
	elsewhere := b12Permanent(g, me.ID, "Other Aura", "Enchantment — Aura")
	woeB1Attach(t, g, elsewhere, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("with an Aura and an Equipment on Kellan your creature is %d/%d, want 4/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, kellan); p != 2 || tough != 2 {
		t.Errorf("Kellan pumped himself: %d/%d", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestWoeILBirthrightBoonFindsAnAuraOrEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	aura := stapleLibraryCard(me, "Rancor", "Enchantment — Aura")
	anthem := stapleLibraryCard(me, "Glorious Anthem", "Enchantment")
	card := woeB1Adventure(me, woeILKellanOracle,
		game.Face{Name: "Kellan, the Fae-Blooded", TypeLine: "Legendary Creature — Human Faerie", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 2},
		game.Face{Name: "Birthright Boon", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{W}", Colors: []string{"W"}})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast the Adventure: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{anthem}); err == nil {
		t.Fatal("a plain enchantment was a legal pick")
	}
	answerSearchByID(t, g, me.ID, aura)
	if !inHand(me, aura) {
		t.Error("the Aura is not in your hand")
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure card is not in exile after it resolved")
	}
}

// --- Kindled Heroism -------------------------------------------------------------

func TestWoeILKindledHeroismPumpsGrantsFirstStrikeAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	castCatalogSpell(t, g, "Kindled Heroism", "Instant", woeILHeroismOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the target is %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target has no first strike")
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("scry prompt %v, want one looking at one card", c)
	}
	answerScryKeepAll(t, g, me.ID)
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
}

// --- Knight of Doves -------------------------------------------------------------

func TestWoeILKnightOfDovesMakesABirdWhenYourEnchantmentGoesToAGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Knight of Doves", "Creature — Human Knight", woeILDovesOracle, false)
	theirs := b12Permanent(g, opp.ID, "Their Anthem", "Enchantment")
	woeB1Destroy(t, g, theirs)
	if got := b27CountNamed(g, me.ID, "Bird"); got != 0 {
		t.Fatalf("%d Birds after an opponent's enchantment died, want 0", got)
	}
	mine := b12Permanent(g, me.ID, "My Anthem", "Enchantment")
	woeB1Destroy(t, g, mine)
	if got := b27CountNamed(g, me.ID, "Bird"); got != 1 {
		t.Fatalf("%d Birds after your enchantment died, want 1", got)
	}
	bird := findBattlefieldByName(g, "Bird")
	if c, _ := battlefieldCard(g, bird); c.Power != 1 || c.Toughness != 1 {
		t.Errorf("the Bird is %d/%d, want 1/1", c.Power, c.Toughness)
	}
	assertKeywords(t, g, bird, "flying")
}

// --- Korvold and the Noble Thief -------------------------------------------------

func TestWoeILKorvoldMakesTwoTreasuresThenImpulsesThreeOffAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	for _, p := range g.Seats {
		hobB10Filler(p)
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	saga := castCatalogSpell(t, g, "Korvold and the Noble Thief", "Enchantment — Saga", woeILKorvoldOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("chapter I: %d Treasures, want 1", n)
	}
	hobB10NextChapter(t, g)
	if n := b27CountNamed(g, me.ID, "Treasure"); n != 2 {
		t.Fatalf("chapter II: %d Treasures, want 2", n)
	}
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	var top []uuid.UUID
	for i := len(opp.Library.Cards) - 1; i >= 0 && len(top) < 3; i-- {
		top = append(top, opp.Library.Cards[i].InstanceID)
	}
	if latestPickTarget(g, me.ID) != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	for _, id := range top {
		if !g.Exile.Contains(id) {
			t.Fatalf("chapter III: %v of the opponent's top three is not in exile", id)
		}
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("casting an exiled card this turn: %v", err)
	}
}

// --- Lady of Laughter ------------------------------------------------------------

func TestWoeILLadyOfLaughterDrawsAtYourEndStepOnlyWhileCelebrating(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	lady := pushCatalogPermanent(g, me.ID, "Lady of Laughter", "Creature — Faerie Noble", woeILLadyOracle, false)
	assertKeywords(t, g, lady, "flying")
	hand := handSize(me)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand {
		t.Fatalf("without celebration the hand went %d -> %d", hand, got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	seedLibrary(me2, "A", "B", "C")
	pushCatalogPermanent(g2, me2.ID, "Lady of Laughter", "Creature — Faerie Noble", woeILLadyOracle, false)
	advanceTo(t, g2, game.StepPrecombatMain)
	celebrationTokens(t, g2, me2.ID, 2)
	hand2 := handSize(me2)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if got := handSize(me2); got != hand2+1 {
		t.Errorf("while celebrating the hand went %d -> %d, want +1", hand2, got)
	}
}

func TestWoeILLadyOfLaughterIgnoresAnotherPlayersEndStep(t *testing.T) {
	g := newCatalogGame(t)
	active, other := woeB1Seats(g)
	seedLibrary(other, "A", "B")
	pushCatalogPermanent(g, other.ID, "Lady of Laughter", "Creature — Faerie Noble", woeILLadyOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	celebrationTokens(t, g, other.ID, 2)
	hand := handSize(other)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := handSize(other); got != hand {
		t.Errorf("the Lady drew on %v's end step: hand %d -> %d", active.ID, hand, got)
	}
}

// --- Leaping Ambush --------------------------------------------------------------

func TestWoeILLeapingAmbushPumpsGrantsReachAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	b16Tap(g, bear)
	castCatalogSpell(t, g, "Leaping Ambush", "Instant", woeILAmbushOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 5 {
		t.Errorf("the target is %d/%d, want 3/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("the target has no reach")
	}
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the target was not untapped")
	}
}

// --- Lord Skitter's Butcher ------------------------------------------------------

func woeILButcherEnters(t *testing.T, g *game.Game, me *game.Player, mode int) uuid.UUID {
	t.Helper()
	butcher := castCatalogSpell(t, g, "Lord Skitter's Butcher", "Creature — Rat Peasant", woeILButcherOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(butcher) {
		t.Fatal("Lord Skitter's Butcher did not resolve")
	}
	woeILPickMode(t, g, me.ID, mode)
	passPriorityAroundTable(t, g)
	return butcher
}

func TestWoeILLordSkittersButcherMakesARatThatCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	woeILButcherEnters(t, g, me, 0)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rat tokens, want 1", got)
	}
	assertRestrictions(t, g, findBattlefieldByName(g, "Rat"), game.CantBlock)
}

func TestWoeILLordSkittersButcherSacrificesAnotherToScryThenDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	fodder := b12Creature(g, me.ID, "Fodder", woeMPBearType, 2, 2)
	hand := handSize(me)
	butcher := castCatalogSpell(t, g, "Lord Skitter's Butcher", "Creature — Rat Peasant", woeILButcherOracle, nil)
	passPriorityAroundTable(t, g)
	woeILPickMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{butcher}); err == nil {
		t.Fatal("the Butcher sacrificed itself")
	}
	answerOwnPermanents(t, g, me.ID, fodder)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	s := scryChoiceFor(g, me.ID)
	if s == nil || len(s.ScryCards) != 2 {
		t.Fatalf("scry prompt %v, want one looking at two cards", s)
	}
	if got := handSize(me); got != hand {
		t.Fatalf("drew before the scry was answered: hand %d -> %d", hand, got)
	}
	answerScryKeepAll(t, g, me.ID)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d after the scry, want %d", got, hand+1)
	}
}

func TestWoeILLordSkittersButcherDeclinedSacrificeDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	fodder := b12Creature(g, me.ID, "Fodder", woeMPBearType, 2, 2)
	hand := handSize(me)
	castCatalogSpell(t, g, "Lord Skitter's Butcher", "Creature — Rat Peasant", woeILButcherOracle, nil)
	passPriorityAroundTable(t, g)
	woeILPickMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) || handSize(me) != hand || scryChoiceFor(g, me.ID) != nil {
		t.Error("declining still sacrificed, scried or drew")
	}
}

func TestWoeILLordSkittersButcherGivesYourCreaturesMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeMPBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeMPBearType, 2, 2)
	butcher := woeILButcherEnters(t, g, me, 2)
	for name, id := range map[string]uuid.UUID{"your creature": bear, "the Butcher": butcher} {
		if !hasEffectiveKeyword(t, g, id, "menace") {
			t.Errorf("%s has no menace", name)
		}
	}
	if hasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("an opponent's creature gained menace")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("menace outlasted the turn")
	}
}

// --- Lord Skitter, Sewer King ----------------------------------------------------

func TestWoeILLordSkitterExilesFromAnOpponentsGraveyardWhenAnotherRatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Lord Skitter, Sewer King", "Legendary Creature — Rat Noble", woeILSkitterOracle, false)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: woeMPBearType, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Dead Bear", TypeLine: woeMPBearType, Owner: me.ID, Controller: me.ID})

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Bear", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2}, 1)
	})
	if latestPickTarget(g, me.ID) != nil || triggerOnStackFrom(g, "Lord Skitter, Sewer King") {
		t.Fatal("a creature that is not a Rat triggered Lord Skitter")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, CantBlockRatToken(), 1) })
	b04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(mine)); err == nil {
		t.Fatal("Lord Skitter targeted a card in your own graveyard")
	}
	pickCard(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead) || opp.Graveyard.Contains(dead) {
		t.Error("the opponent's card was not exiled")
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("your own card left your graveyard")
	}
}

func TestWoeILLordSkitterMakesARatAtTheBeginningOfYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	pushCatalogPermanent(g, me.ID, "Lord Skitter, Sewer King", "Legendary Creature — Rat Noble", woeILSkitterOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Fatalf("%d Rat tokens at the beginning of combat, want 1", got)
	}
	assertRestrictions(t, g, findBattlefieldByName(g, "Rat"), game.CantBlock)
}
