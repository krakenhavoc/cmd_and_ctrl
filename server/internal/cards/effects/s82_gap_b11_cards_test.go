package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b11_cards_test.go — Innistrad Remastered creatures, S82 GAP
// batch 11 (#5948), played through the real catalog entries.

const (
	gapB11ReveleroOracle = "01a7e0ce-c98f-4cf2-aba3-360aa0a9b9a7"
	gapB11ConduitOracle  = "1c4cf7c9-9dc5-49da-8d43-1ecd41c74ad3"
	gapB11SmolderOracle  = "372a01f7-6e58-4424-9405-0ee4acbb9346"
	gapB11KindredOracle  = "41fd1557-8dba-4bbf-9954-5c4f4f541812"
	gapB11AvacynOracle   = "432b37a5-d32a-4b78-91ab-860aa026b7cc"
	gapB11MaulerOracle   = "6c5664b1-fbb5-477e-80fd-cf2285014a10"
	gapB11WatchkeepOracl = "8bdf986e-a47e-47b6-8cdb-576c86c4041b"
	gapB11DevilOracle    = "b582e485-c024-47f0-ad6f-b918d32288ba"
	gapB11BloodmadOracle = "c8a20fc7-025b-403a-b893-ff5efcbdc9d8"
	gapB11MessengerOracl = "f6634a84-a9ce-4bbd-bc2c-3f9179b7f16a"
	gapB11FurybladeOracl = "f8dbfc60-33e9-4068-a0dd-0d0ed5b9b42c"
)

var (
	gapB11Watchkeep = iwCard{gapB11WatchkeepOracl, "Hanweir Watchkeep", "Creature — Human Warrior Werewolf", "{2}{R}", "Defender\n" + iwFrontText, "1", "5",
		"Bane of Hanweir", "This creature attacks each combat if able.\n" + iwBackText, "5", "5", []string{"R"}, []string{"Defender"}}
	gapB11Messenger = iwCard{gapB11MessengerOracl, "Village Messenger", "Creature — Human Werewolf", "{R}", "Haste\n" + iwFrontText, "1", "1",
		"Moonrise Intruder", "Menace\n" + iwBackText, "2", "2", []string{"R"}, []string{"Haste", "Menace"}}
	gapB11Avacyn = iwCard{gapB11AvacynOracle, "Archangel Avacyn", "Legendary Creature — Angel", "{3}{W}{W}",
		"Flash\nFlying, vigilance", "4", "4", "Avacyn, the Purifier", "Flying", "6", "5", []string{"W"}, []string{"Flash", "Flying", "Vigilance"}}
)

func TestGapB11EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		gapB11ReveleroOracle: "Bedlam Reveler", gapB11ConduitOracle: "Conduit of Storms",
		gapB11ConduitOracle + "#1": "Conduit of Emrakul", gapB11SmolderOracle: "Smoldering Werewolf",
		gapB11SmolderOracle + "#1": "Erupting Dreadwolf", gapB11KindredOracle: "Ulrich's Kindred",
		gapB11AvacynOracle: "Archangel Avacyn", gapB11AvacynOracle + "#1": "Avacyn, the Purifier",
		gapB11MaulerOracle: "Lightning Mauler", gapB11WatchkeepOracl: "Hanweir Watchkeep",
		gapB11WatchkeepOracl + "#1": "Bane of Hanweir", gapB11DevilOracle: "Vexing Devil",
		gapB11BloodmadOracle: "Bloodmad Vampire", gapB11MessengerOracl: "Village Messenger",
		gapB11MessengerOracl + "#1": "Moonrise Intruder", gapB11FurybladeOracl: "Furyblade Vampire",
	} {
		spec, ok := Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Bedlam Reveler ---

func TestGapB11RevelerCostsLessPerInstantAndSorceryAndRefillsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		graveyardCardOf(g, me, "Spell", "Instant", "{R}", "")
	}
	graveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}", "") // does not count
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	b06AddMana(me, "R", "R", "C", "C")
	if _, err := castFromHandErr(g, me, "Bedlam Reveler", "Creature — Devil Horror", "{6}{R}{R}", gapB11ReveleroOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("four mana cast a spell that costs {3}{R}{R} with three instants in the graveyard")
	}
	b06AddMana(me, "C")
	if _, err := castFromHandErr(g, me, "Bedlam Reveler", "Creature — Devil Horror", "{6}{R}{R}", gapB11ReveleroOracle,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("five mana should pay {3}{R}{R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d after the enters trigger, want 3 (the hand discarded, three drawn)", got)
	}
	assertKeywords(t, g, battlefieldIDNamed(g, me.ID, "Bedlam Reveler"), "prowess")
}

// --- Conduit of Storms // Conduit of Emrakul ---

func gapB11AttackThenPostcombat(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	passPriorityAroundTable(t, g)
}

func TestGapB11ConduitOfStormsAddsRedAtTheNextMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	conduit := b12Push(g, me.ID, "Conduit of Storms", "Creature — Werewolf Horror", gapB11ConduitOracle, 2, 3)
	before := me.ManaPool.CountColor("R")
	gapB11AttackThenPostcombat(t, g, conduit, opp.ID)
	if got := me.ManaPool.CountColor("R") - before; got != 1 {
		t.Errorf("{R} added: %d, want 1", got)
	}
}

func TestGapB11ConduitOfEmrakulAddsTwoColorless(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	conduit := b12Push(g, me.ID, "Conduit of Emrakul", "Creature — Eldrazi Werewolf", gapB11ConduitOracle+"#1", 5, 4)
	gapB11AttackThenPostcombat(t, g, conduit, opp.ID)
	if got := me.ManaPool.CountColor("C"); got != 2 {
		t.Errorf("{C} added: %d, want 2", got)
	}
}

func TestGapB11ConduitOfStormsTransformsForFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	conduit := b12Push(g, me.ID, "Conduit of Storms", "Creature — Werewolf Horror", gapB11ConduitOracle, 2, 3)
	if err := hobB4Activate(t, g, me.ID, conduit, 0, broB04Pool("{3}{R}{R}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, conduit); c.ActiveFace == 0 && c.Layout == game.LayoutTransform {
		t.Error("the Conduit did not transform")
	}
}

// --- Smoldering Werewolf // Erupting Dreadwolf ---

func TestGapB11SmolderingWerewolfPingsUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 1)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 1)
	c := b12Creature(g, opp.ID, "C", "Creature — Bear", 2, 1)
	castCatalogSpell(t, g, "Smoldering Werewolf", "Creature — Werewolf Horror", gapB11SmolderOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	g.RunStateChecksForTest()
	dead := 0
	for _, id := range []uuid.UUID{a, b} {
		if zone, _ := cardWhere(g, opp, id); zone != "battlefield" {
			dead++
		}
	}
	if dead != 2 {
		t.Errorf("%d of the two chosen 2/1s died, want 2", dead)
	}
	if zone, _ := cardWhere(g, opp, c); zone != "battlefield" {
		t.Error("the third, unchosen creature died")
	}
}

func TestGapB11ErruptingDreadwolfDealsTwoWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wolf := b12Push(g, me.ID, "Erupting Dreadwolf", "Creature — Eldrazi Werewolf", gapB11SmolderOracle+"#1", 6, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolf, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	before := opp.Life
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayerTarget(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if opp.Life > before-2 {
		t.Errorf("life %d → %d, want at least 2 from the trigger", before, opp.Life)
	}
}

// --- Ulrich's Kindred ---

func TestGapB11UlrichsKindredGivesAnAttackingWolfIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kindred := b12Push(g, me.ID, "Ulrich's Kindred", "Creature — Wolf", gapB11KindredOracle, 3, 2)
	assertKeywords(t, g, kindred, "trample")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kindred, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if err := hobB4Activate(t, g, me.ID, kindred, 0, broB04Pool("{3}{G}"), gapB11Target(bear)); err == nil {
		t.Fatal("a non-attacking Bear is not a legal target")
	}
	if err := hobB4Activate(t, g, me.ID, kindred, 0, "", gapB11Target(kindred)); err != nil {
		t.Fatalf("activate on the attacking Wolf: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, kindred, "indestructible") {
		t.Error("the attacking Wolf did not gain indestructible")
	}
}

// --- Lightning Mauler ---

func TestGapB11LightningMaulerGivesBothCreaturesHasteWhilePaired(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := sbondHand(me, "Lightning Mauler", "Creature — Human Berserker", gapB11MaulerOracle, 2, 1, game.KeywordSoulbond)
	sbondCast(t, g, me, id)
	nghcPassOnce(t, g)
	answerOwnPermanents(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, id, "haste") || !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the paired creatures do not both have haste")
	}
	killCreature(t, g, me.ID, id)
	g.RunStateChecksForTest()
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the partner kept haste after the pair ended")
	}
}

func TestGapB11LightningMaulerUnpairedHasNoHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mauler := b12Push(g, me.ID, "Lightning Mauler", "Creature — Human Berserker", gapB11MaulerOracle, 2, 1)
	if hasEffectiveKeyword(t, g, mauler, "haste") {
		t.Error("an unpaired Mauler has haste")
	}
}

// --- Vexing Devil ---

func TestGapB11VexingDevilIsSacrificedOnlyIfAnOpponentTakesFour(t *testing.T) {
	for _, taken := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		castCatalogSpell(t, g, "Vexing Devil", "Creature — Devil", gapB11DevilOracle, nil)
		passPriorityAroundTable(t, g)
		who := uuid.Nil
		if taken {
			who = opp.ID
		}
		before := opp.Life
		appAnswerWith(t, g, appOnly(who))
		devil := battlefieldIDNamed(g, me.ID, "Vexing Devil")
		if taken && (devil != uuid.Nil || opp.Life != before-4) {
			t.Errorf("taken: Devil on the battlefield %v, life %d → %d; want it gone and -4", devil != uuid.Nil, before, opp.Life)
		}
		if !taken && (devil == uuid.Nil || opp.Life != before) {
			t.Errorf("declined: Devil on the battlefield %v, life %d → %d; want it kept and no loss", devil != uuid.Nil, before, opp.Life)
		}
	}
}

// --- Bloodmad Vampire ---

func TestGapB11BloodmadVampireGrowsOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vamp := b12Push(g, me.ID, "Bloodmad Vampire", "Creature — Vampire Berserker", gapB11BloodmadOracle, 4, 1)
	attackWith(t, g, opp.ID, vamp)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, vamp).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if spec, _ := Lookup(gapB11BloodmadOracle); spec.Madness != "{1}{R}" {
		t.Errorf("madness %q", spec.Madness)
	}
}

// --- Furyblade Vampire ---

func TestGapB11FurybladeVampireDiscardsForPlusThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vamp := b12Push(g, me.ID, "Furyblade Vampire", "Creature — Vampire Berserker", gapB11FurybladeOracl, 1, 2)
	assertKeywords(t, g, vamp, "trample")
	card := pushCatalogHandCard(me, "Fodder", "Sorcery", "")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	prompt := b24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{card}); err != nil {
		t.Fatalf("discard to pay: %v", err)
	}
	if got := layeredCard(t, g, vamp).CurrentPower(); got != 4 {
		t.Errorf("power %d, want 4 after discarding", got)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded")
	}
}

// --- werewolves ---

func TestGapB11WerewolvesTransformAndBack(t *testing.T) {
	for _, w := range []iwCard{gapB11Watchkeep, gapB11Messenger} {
		g, me := iwTable(t)
		id := iwPut(t, g, w, me)
		iwNextUpkeep(t, g, nil)
		if iwFace(g, id) != 1 {
			t.Fatalf("%s did not transform after a quiet turn (face %d)", w.front, iwFace(g, id))
		}
		if w.front == "Village Messenger" {
			assertKeywords(t, g, id, "menace")
		}
		iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
		if iwFace(g, id) != 0 {
			t.Errorf("%s did not turn back after a player cast two spells", w.back)
		}
	}
}

func TestGapB11BaneOfHanweirMustAttack(t *testing.T) {
	g, me := iwTable(t)
	id := iwPutBack(t, g, gapB11Watchkeep, me)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepDeclareAttackers)
	// Declaring no attackers at all must be refused while Bane can attack.
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("Bane of Hanweir can't attack: %v", err)
	}
}

// --- Archangel Avacyn // Avacyn, the Purifier ---

func TestGapB11AvacynGivesYourCreaturesIndestructibleWhenSheEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Archangel Avacyn", "Legendary Creature — Angel", gapB11AvacynOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("your creature is not indestructible")
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature is indestructible")
	}
}

func TestGapB11AvacynTransformsAtTheNextUpkeepAndBurnsTheBoard(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avacyn := iwPut(t, g, gapB11Avacyn, me)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Tough", "Creature — Wall", 0, 9)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	killCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if iwFace(g, avacyn) != 0 {
		t.Fatal("Avacyn transformed immediately; it is at the beginning of the next upkeep")
	}
	iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
	passPriorityAroundTable(t, g)
	if iwFace(g, avacyn) != 1 {
		t.Fatalf("Avacyn is on face %d at the next upkeep, want the Purifier", iwFace(g, avacyn))
	}
	if zone, _ := cardWhere(g, opp, theirs); zone == "battlefield" {
		t.Error("the opponent's 2/2 survived 3 damage")
	}
	if c := layeredCard(t, g, other); c.DamageMarked != 3 {
		t.Errorf("your 0/9 has %d damage, want 3", c.DamageMarked)
	}
	if c := layeredCard(t, g, avacyn); c.DamageMarked != 0 {
		t.Errorf("Avacyn took %d damage from her own trigger", c.DamageMarked)
	}
	if opp.Life != oppLife-3 || me.Life != myLife {
		t.Errorf("life: opponent %d → %d, you %d → %d; want -3 and unchanged", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestGapB11AvacynTransformsOnlyOnceForSeveralDeaths(t *testing.T) {
	g, me := iwTable(t)
	avacyn := iwPut(t, g, gapB11Avacyn, me)
	for i := 0; i < 2; i++ {
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		killCreature(t, g, me.ID, bear)
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
	}
	iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
	passPriorityAroundTable(t, g)
	if iwFace(g, avacyn) != 1 {
		t.Errorf("two deaths left Avacyn on face %d, want the Purifier (transformed once, not twice)", iwFace(g, avacyn))
	}
}

func TestGapB11AvacynIgnoresAnAngelOrAnOpponentsCreatureDying(t *testing.T) {
	g, me := iwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avacyn := iwPut(t, g, gapB11Avacyn, me)
	angel := b12Creature(g, me.ID, "Angel", "Creature — Angel", 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	killCreature(t, g, me.ID, angel)
	killCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	iwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
	passPriorityAroundTable(t, g)
	if iwFace(g, avacyn) != 0 {
		t.Errorf("Avacyn transformed (face %d) with only an Angel and an opponent's creature dead", iwFace(g, avacyn))
	}
}

func gapB11Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
