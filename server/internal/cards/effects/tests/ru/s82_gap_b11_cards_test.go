package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	gapB11Watchkeep = effectstest.IwCard{Oracle: gapB11WatchkeepOracl, Front: "Hanweir Watchkeep", FrontType: "Creature — Human Warrior Werewolf", Cost: "{2}{R}", FrontText: "Defender\n" + effectstest.IwFrontText, Fp: "1", Ft: "5",
		Back: "Bane of Hanweir", BackText: "This creature attacks each combat if able.\n" + effectstest.IwBackText, Bp: "5", Bt: "5", Colors: []string{"R"}, Keywords: []string{"Defender"}}
	gapB11Messenger = effectstest.IwCard{Oracle: gapB11MessengerOracl, Front: "Village Messenger", FrontType: "Creature — Human Werewolf", Cost: "{R}", FrontText: "Haste\n" + effectstest.IwFrontText, Fp: "1", Ft: "1",
		Back: "Moonrise Intruder", BackText: "Menace\n" + effectstest.IwBackText, Bp: "2", Bt: "2", Colors: []string{"R"}, Keywords: []string{"Haste", "Menace"}}
	gapB11Avacyn = effectstest.IwCard{Oracle: gapB11AvacynOracle, Front: "Archangel Avacyn", FrontType: "Legendary Creature — Angel", Cost: "{3}{W}{W}",
		FrontText: "Flash\nFlying, vigilance", Fp: "4", Ft: "4", Back: "Avacyn, the Purifier", BackText: "Flying", Bp: "6", Bt: "5", Colors: []string{"W"}, Keywords: []string{"Flash", "Flying", "Vigilance"}}
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
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// --- Bedlam Reveler ---

func TestGapB11RevelerCostsLessPerInstantAndSorceryAndRefillsTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.GraveyardCardOf(g, me, "Spell", "Instant", "{R}", "")
	}
	effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}", "") // does not count
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.B06AddMana(me, "R", "R", "C", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bedlam Reveler", "Creature — Devil Horror", "{6}{R}{R}", gapB11ReveleroOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("four mana cast a spell that costs {3}{R}{R} with three instants in the graveyard")
	}
	effectstest.B06AddMana(me, "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Bedlam Reveler", "Creature — Devil Horror", "{6}{R}{R}", gapB11ReveleroOracle,
		game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("five mana should pay {3}{R}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d after the enters trigger, want 3 (the hand discarded, three drawn)", got)
	}
	effectstest.AssertKeywords(t, g, effectstest.BattlefieldIDNamed(g, me.ID, "Bedlam Reveler"), "prowess")
}

// --- Conduit of Storms // Conduit of Emrakul ---

func gapB11AttackThenPostcombat(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB11ConduitOfStormsAddsRedAtTheNextMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	conduit := effectstest.B12Push(g, me.ID, "Conduit of Storms", "Creature — Werewolf Horror", gapB11ConduitOracle, 2, 3)
	before := me.ManaPool.CountColor("R")
	gapB11AttackThenPostcombat(t, g, conduit, opp.ID)
	if got := me.ManaPool.CountColor("R") - before; got != 1 {
		t.Errorf("{R} added: %d, want 1", got)
	}
}

func TestGapB11ConduitOfEmrakulAddsTwoColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	conduit := effectstest.B12Push(g, me.ID, "Conduit of Emrakul", "Creature — Eldrazi Werewolf", gapB11ConduitOracle+"#1", 5, 4)
	gapB11AttackThenPostcombat(t, g, conduit, opp.ID)
	if got := me.ManaPool.CountColor("C"); got != 2 {
		t.Errorf("{C} added: %d, want 2", got)
	}
}

func TestGapB11ConduitOfStormsTransformsForFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	conduit := effectstest.B12Push(g, me.ID, "Conduit of Storms", "Creature — Werewolf Horror", gapB11ConduitOracle, 2, 3)
	if err := effectstest.HobB4Activate(t, g, me.ID, conduit, 0, effectstest.BroB04Pool("{3}{R}{R}"), nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, conduit); c.ActiveFace == 0 && c.Layout == game.LayoutTransform {
		t.Error("the Conduit did not transform")
	}
}

// --- Smoldering Werewolf // Erupting Dreadwolf ---

func TestGapB11SmolderingWerewolfPingsUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 1)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 1)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 1)
	effectstest.CastCatalogSpell(t, g, "Smoldering Werewolf", "Creature — Werewolf Horror", gapB11SmolderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	g.RunStateChecksForTest()
	dead := 0
	for _, id := range []uuid.UUID{a, b} {
		if zone, _ := effectstest.CardWhere(g, opp, id); zone != "battlefield" {
			dead++
		}
	}
	if dead != 2 {
		t.Errorf("%d of the two chosen 2/1s died, want 2", dead)
	}
	if zone, _ := effectstest.CardWhere(g, opp, c); zone != "battlefield" {
		t.Error("the third, unchosen creature died")
	}
}

func TestGapB11ErruptingDreadwolfDealsTwoWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wolf := effectstest.B12Push(g, me.ID, "Erupting Dreadwolf", "Creature — Eldrazi Werewolf", gapB11SmolderOracle+"#1", 6, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(wolf, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	before := opp.Life
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life > before-2 {
		t.Errorf("life %d → %d, want at least 2 from the trigger", before, opp.Life)
	}
}

// --- Ulrich's Kindred ---

func TestGapB11UlrichsKindredGivesAnAttackingWolfIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kindred := effectstest.B12Push(g, me.ID, "Ulrich's Kindred", "Creature — Wolf", gapB11KindredOracle, 3, 2)
	effectstest.AssertKeywords(t, g, kindred, "trample")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kindred, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, kindred, 0, effectstest.BroB04Pool("{3}{G}"), gapB11Target(bear)); err == nil {
		t.Fatal("a non-attacking Bear is not a legal target")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, kindred, 0, "", gapB11Target(kindred)); err != nil {
		t.Fatalf("activate on the attacking Wolf: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, kindred, "indestructible") {
		t.Error("the attacking Wolf did not gain indestructible")
	}
}

// --- Lightning Mauler ---

func TestGapB11LightningMaulerGivesBothCreaturesHasteWhilePaired(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := effectstest.SbondHand(me, "Lightning Mauler", "Creature — Human Berserker", gapB11MaulerOracle, 2, 1, game.KeywordSoulbond)
	effectstest.SbondCast(t, g, me, id)
	effectstest.NghcPassOnce(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") || !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the paired creatures do not both have haste")
	}
	effectstest.KillCreature(t, g, me.ID, id)
	g.RunStateChecksForTest()
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the partner kept haste after the pair ended")
	}
}

func TestGapB11LightningMaulerUnpairedHasNoHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mauler := effectstest.B12Push(g, me.ID, "Lightning Mauler", "Creature — Human Berserker", gapB11MaulerOracle, 2, 1)
	if effectstest.HasEffectiveKeyword(t, g, mauler, "haste") {
		t.Error("an unpaired Mauler has haste")
	}
}

// --- Vexing Devil ---

func TestGapB11VexingDevilIsSacrificedOnlyIfAnOpponentTakesFour(t *testing.T) {
	for _, taken := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.CastCatalogSpell(t, g, "Vexing Devil", "Creature — Devil", gapB11DevilOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if taken {
			who = opp.ID
		}
		before := opp.Life
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who))
		devil := effectstest.BattlefieldIDNamed(g, me.ID, "Vexing Devil")
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vamp := effectstest.B12Push(g, me.ID, "Bloodmad Vampire", "Creature — Vampire Berserker", gapB11BloodmadOracle, 4, 1)
	effectstest.AttackWith(t, g, opp.ID, vamp)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LayeredCard(t, g, vamp).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if spec, _ := effects.Lookup(gapB11BloodmadOracle); spec.Madness != "{1}{R}" {
		t.Errorf("madness %q", spec.Madness)
	}
}

// --- Furyblade Vampire ---

func TestGapB11FurybladeVampireDiscardsForPlusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vamp := effectstest.B12Push(g, me.ID, "Furyblade Vampire", "Creature — Vampire Berserker", gapB11FurybladeOracl, 1, 2)
	effectstest.AssertKeywords(t, g, vamp, "trample")
	card := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{card}); err != nil {
		t.Fatalf("discard to pay: %v", err)
	}
	if got := effectstest.LayeredCard(t, g, vamp).CurrentPower(); got != 4 {
		t.Errorf("power %d, want 4 after discarding", got)
	}
	if me.Hand.Contains(card) {
		t.Error("the card was not discarded")
	}
}

// --- werewolves ---

func TestGapB11WerewolvesTransformAndBack(t *testing.T) {
	for _, w := range []effectstest.IwCard{gapB11Watchkeep, gapB11Messenger} {
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, w, me)
		effectstest.IwNextUpkeep(t, g, nil)
		if effectstest.IwFace(g, id) != 1 {
			t.Fatalf("%s did not transform after a quiet turn (face %d)", w.Front, effectstest.IwFace(g, id))
		}
		if w.Front == "Village Messenger" {
			effectstest.AssertKeywords(t, g, id, "menace")
		}
		effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 2})
		if effectstest.IwFace(g, id) != 0 {
			t.Errorf("%s did not turn back after a player cast two spells", w.Back)
		}
	}
}

func TestGapB11BaneOfHanweirMustAttack(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPutBack(t, g, gapB11Watchkeep, me)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	// Declaring no attackers at all must be refused while Bane can attack.
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("Bane of Hanweir can't attack: %v", err)
	}
}

// --- Archangel Avacyn // Avacyn, the Purifier ---

func TestGapB11AvacynGivesYourCreaturesIndestructibleWhenSheEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Archangel Avacyn", "Legendary Creature — Angel", gapB11AvacynOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("your creature is not indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("an opponent's creature is indestructible")
	}
}

func TestGapB11AvacynTransformsAtTheNextUpkeepAndBurnsTheBoard(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avacyn := effectstest.IwPut(t, g, gapB11Avacyn, me)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Tough", "Creature — Wall", 0, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, avacyn) != 0 {
		t.Fatal("Avacyn transformed immediately; it is at the beginning of the next upkeep")
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, avacyn) != 1 {
		t.Fatalf("Avacyn is on face %d at the next upkeep, want the Purifier", effectstest.IwFace(g, avacyn))
	}
	if zone, _ := effectstest.CardWhere(g, opp, theirs); zone == "battlefield" {
		t.Error("the opponent's 2/2 survived 3 damage")
	}
	if c := effectstest.LayeredCard(t, g, other); c.DamageMarked != 3 {
		t.Errorf("your 0/9 has %d damage, want 3", c.DamageMarked)
	}
	if c := effectstest.LayeredCard(t, g, avacyn); c.DamageMarked != 0 {
		t.Errorf("Avacyn took %d damage from her own trigger", c.DamageMarked)
	}
	if opp.Life != oppLife-3 || me.Life != myLife {
		t.Errorf("life: opponent %d → %d, you %d → %d; want -3 and unchanged", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestGapB11AvacynTransformsOnlyOnceForSeveralDeaths(t *testing.T) {
	g, me := effectstest.IwTable(t)
	avacyn := effectstest.IwPut(t, g, gapB11Avacyn, me)
	for i := 0; i < 2; i++ {
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		effectstest.KillCreature(t, g, me.ID, bear)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, avacyn) != 1 {
		t.Errorf("two deaths left Avacyn on face %d, want the Purifier (transformed once, not twice)", effectstest.IwFace(g, avacyn))
	}
}

func TestGapB11AvacynIgnoresAnAngelOrAnOpponentsCreatureDying(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	avacyn := effectstest.IwPut(t, g, gapB11Avacyn, me)
	angel := effectstest.B12Creature(g, me.ID, "Angel", "Creature — Angel", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, me.ID, angel)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.IwNextUpkeep(t, g, map[int]int{g.Turn.ActiveSeat: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IwFace(g, avacyn) != 0 {
		t.Errorf("Avacyn transformed (face %d) with only an Angel and an opponent's creature dead", effectstest.IwFace(g, avacyn))
	}
}

func gapB11Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}
