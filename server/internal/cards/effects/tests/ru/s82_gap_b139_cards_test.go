package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b139_cards_test.go — S82 GAP batch 139 (Dominaria Remastered,
// #5948): Sawtooth Loon, Cleric of the Forward Order, Glory, Icatian
// Javelineers, Kjeldoran Gargoyle, Lieutenant Kirtar, Nomad Decoy, Vigilant
// Sentry, Voice of All, Dralnu's Crusade, Invigorating Boon.

const (
	gapB139LoonOracle      = "b917e472-2550-4381-b756-45487c2c187a"
	gapB139ClericOracle    = "84a2a592-0e31-434f-b3b4-f2013d86f29d"
	gapB139GloryOracle     = "3853ef6a-6771-4708-8ed5-d667f646de5a"
	gapB139JavelinOracle   = "26112ba7-2859-4af3-93c3-e62c31177216"
	gapB139GargoyleOracle  = "b82d1104-8a54-4745-ba8b-eeddcb657338"
	gapB139KirtarOracle    = "ad39e859-e4a6-49cc-b738-4578a5250538"
	gapB139DecoyOracle     = "73c40477-7061-47db-af8f-2a5cf6c22c9b"
	gapB139SentryOracle    = "9bb66d53-a83c-4313-91c3-4df09c248558"
	gapB139VoiceOracle     = "b1631e9c-a993-4d44-b009-b2624d79560c"
	gapB139CrusadeOracle   = "ff48cf80-4950-4ae4-9f7c-8d826b2f26f7"
	gapB139BoonOracle      = "d00329c2-e65c-4774-83cd-d68e1d1a6e31"
	gapB139BoonCreatureTxt = "Creature — Test"
)

func gapB139OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func gapB139InHandNamed(p *game.Player, name string) bool {
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			return true
		}
	}
	return false
}

// gapB139AnswerPicks answers every open pick of `who` (a choose-cards prompt
// or an own-permanents choice), naming `prefer` when it is offered and the
// first card otherwise.
func gapB139AnswerPicks(t *testing.T, g *game.Game, who uuid.UUID, prefer uuid.UUID) {
	t.Helper()
	choose := func(cards []uuid.UUID) uuid.UUID {
		for _, c := range cards {
			if c == prefer {
				return c
			}
		}
		return cards[0]
	}
	for i := 0; i < 8; i++ {
		effectstest.AnswerOpenOrderingIfAny(t, g, who)
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, who); p != nil {
			if err := g.ResolveOwnPermanents(p.ID, who, []uuid.UUID{choose(p.ChooseCards)}); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
			continue
		}
		if p := effectstest.ChooseCardsChoiceFor(g, who); p != nil {
			// A put-back from the hand: anything but the card kept.
			pick := p.ChooseCards[0]
			for _, c := range p.ChooseCards {
				if c != prefer {
					pick = c
					break
				}
			}
			effectstest.AnswerChooseCards(t, g, who, pick)
			continue
		}
		return
	}
}

func gapB139Coloured(g *game.Game, owner uuid.UUID, name, color string, typeLine ...string) uuid.UUID {
	tl := gapB139BoonCreatureTxt
	if len(typeLine) > 0 {
		tl = typeLine[0]
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: tl, Power: 2, Toughness: 2,
		Colors: []string{color}, Owner: owner, Controller: owner,
	})
}

func gapB139CastColoured(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Power: power, Toughness: toughness, Colors: colors, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return id
}

func TestGapB139SawtoothLoonBouncesAWhiteCreatureAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "White Bear", gapB139BoonCreatureTxt, 2, 2)
	green := effectstest.B12Creature(g, me.ID, "Green Bear", gapB139BoonCreatureTxt, 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, bear).Colors = []string{"W"}
		effectstest.FindBattlefieldCardForTest(g, green).Colors = []string{"G"}
		g.RecomputeLayersIfStaleLocked()
	})
	lib := me.Library.Size()
	hand := me.Hand.Size()
	loon := gapB139CastColoured(t, g, "Sawtooth Loon", "Creature — Bird", gapB139LoonOracle, 2, 2, "W", "U")
	gapB139AnswerPicks(t, g, me.ID, bear)
	if gapB139OnBattlefield(g, bear) || !gapB139InHandNamed(me, "White Bear") {
		t.Error("the white Bear was not returned to its owner's hand")
	}
	if !gapB139OnBattlefield(g, green) || !gapB139OnBattlefield(g, loon) {
		t.Error("the green creature or the Loon left")
	}
	if me.Library.Size() != lib {
		t.Errorf("library = %d, want %d (two drawn, two put back)", me.Library.Size(), lib)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d (the Bear joins it; two drawn, two put back)", me.Hand.Size(), hand+1)
	}
}

func TestGapB139SawtoothLoonCanReturnItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	green := effectstest.B12Creature(g, me.ID, "Green Bear", gapB139BoonCreatureTxt, 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, green).Colors = []string{"G"}
		g.RecomputeLayersIfStaleLocked()
	})
	loon := gapB139CastColoured(t, g, "Sawtooth Loon", "Creature — Bird", gapB139LoonOracle, 2, 2, "W", "U")
	gapB139AnswerPicks(t, g, me.ID, loon)
	if gapB139OnBattlefield(g, loon) {
		t.Error("the Loon was not returned although it was the only white or blue creature")
	}
	if !gapB139OnBattlefield(g, green) {
		t.Error("the green creature was returned")
	}
}

func TestGapB139ClericGainsTwoForEachClericYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	effectstest.CastAndResolveCreature(t, g, "Cleric of the Forward Order", "Creature — Human Cleric", gapB139ClericOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("first Cleric: life %d, want %d", me.Life, life+2)
	}
	effectstest.CastAndResolveCreature(t, g, "Cleric of the Forward Order", "Creature — Human Cleric", gapB139ClericOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2+4 {
		t.Errorf("second Cleric: life %d, want %d (two Clerics)", me.Life, life+6)
	}
}

func TestGapB139GloryFliesAndProtectsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", gapB139BoonCreatureTxt, 2, 2)
	glory := effectstest.PushGraveyardCardWithTypeLineID(me, "Glory", "Creature — Incarnation")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == glory {
				me.Graveyard.Cards[i].OracleID = gapB139GloryOracle
			}
		}
	})
	effectstest.ApaMana(me, "W", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, glory, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, bear, "protection from red") {
		t.Errorf("the Bear has no protection from red")
	}
	if err := g.ActivateCatalogAbility(me.ID, glory, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with an empty pool")
	}
}

func TestGapB139GloryCantBeActivatedOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	glory := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Glory", gapB139GloryOracle, 3, 3))
	if !effectstest.HasKeywordOf(g, glory, "flying") {
		t.Error("no flying")
	}
	effectstest.ApaMana(me, "W", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, glory, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard-only ability was activated from the battlefield")
	}
}

func TestGapB139IcatianJavelineersShootOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jav := gapB139CastColoured(t, g, "Icatian Javelineers", "Creature — Human Soldier", gapB139JavelinOracle, 1, 1, "W")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, jav); c == nil || c.Counters["javelin"] != 1 {
		t.Fatalf("no javelin counter on entry: %+v", c)
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, jav).SummonedThisTurn = false })
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, jav, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
	if c := effectstest.FindBattlefieldCardForTest(g, jav); c == nil || c.Counters["javelin"] != 0 {
		t.Errorf("the javelin counter was not removed: %+v", c)
	}
	if err := g.ActivateCatalogAbility(me.ID, jav, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("activated without a javelin counter")
	}
}

func TestGapB139KjeldoranGargoyleGainsLifeEqualToDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garg := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Kjeldoran Gargoyle", gapB139GargoyleOracle, 3, 3))
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasKeywordOf(g, garg, kw) {
			t.Errorf("missing %s", kw)
		}
	}
	life := me.Life
	effectstest.AttackWith(t, g, opp.ID, garg)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life after dealing 3 = %d, want %d", me.Life, life+3)
	}
}

func TestGapB139LieutenantKirtarExilesAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirtar := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Lieutenant Kirtar", gapB139KirtarOracle, 2, 2))
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Raider", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.ApaMana(opp, "W", "C")
	if err := g.ActivateCatalogAbility(opp.ID, kirtar, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gapB139OnBattlefield(g, kirtar) {
		t.Error("Kirtar was not sacrificed")
	}
	if gapB139OnBattlefield(g, attacker) {
		t.Error("the attacker was not exiled")
	}
}

func TestGapB139LieutenantKirtarNeedsAnAttackingTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirtar := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Lieutenant Kirtar", gapB139KirtarOracle, 2, 2))
	idle := effectstest.PushVanillaCreature(g, me.ID, "Idle", 3, 3)
	effectstest.ApaMana(opp, "W", "C")
	if err := g.ActivateCatalogAbility(opp.ID, kirtar, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Error("targeted a creature that is not attacking")
	}
}

func TestGapB139NomadDecoyTapsOneAndThresholdTapsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	decoy := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Nomad Decoy", gapB139DecoyOracle, 1, 2))
	a := effectstest.B12Creature(g, opp.ID, "Target A", gapB139BoonCreatureTxt, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Target B", gapB139BoonCreatureTxt, 2, 2)
	effectstest.ApaMana(me, "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, decoy, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}}); err != nil {
		t.Fatalf("tap one: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, a); c == nil || !c.Tapped {
		t.Error("Target A is not tapped")
	}
	// The threshold ability is refused below seven cards.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, decoy).Tapped = false })
	two := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.ActivateCatalogAbility(me.ID, decoy, 1, game.ActivateAbilityParams{Strict: true, Targets: two}); err == nil {
		t.Fatal("the threshold ability was activated with an empty graveyard")
	}
	for i := 0; i < 7; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Filler", "Sorcery")
	}
	if err := g.ActivateCatalogAbility(me.ID, decoy, 1, game.ActivateAbilityParams{Strict: true, Targets: two}); err != nil {
		t.Fatalf("threshold ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, b); c == nil || !c.Tapped {
		t.Error("Target B is not tapped by the threshold ability")
	}
}

func TestGapB139VigilantSentryThreshold(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.B12Push(g, me.ID, "Vigilant Sentry", "Creature — Human Nomad", gapB139SentryOracle, 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, sentry); p != 2 || tough != 2 {
		t.Fatalf("without threshold = %d/%d, want 2/2", p, tough)
	}
	for i := 0; i < 7; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Filler", "Sorcery")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, sentry); p != 3 || tough != 3 {
		t.Fatalf("with threshold = %d/%d, want 3/3", p, tough)
	}
	_ = opp
}

func TestGapB139VigilantSentryPumpsAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sentry := effectstest.B12Push(g, me.ID, "Vigilant Sentry", "Creature — Human Nomad", gapB139SentryOracle, 2, 2)
	for i := 0; i < 7; i++ {
		effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Filler", "Sorcery")
	}
	raider := effectstest.PushVanillaCreature(g, me.ID, "Raider", 2, 2)
	opp := g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sentry, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: raider}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, raider); p != 5 || tough != 5 {
		t.Errorf("the attacker = %d/%d, want 5/5", p, tough)
	}
}

func TestGapB139VoiceOfAllHasProtectionFromTheChosenColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	voice := gapB139CastColoured(t, g, "Voice of All", "Creature — Angel", gapB139VoiceOracle, 2, 2, "W")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "B")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOf(g, voice, "flying") {
		t.Error("no flying")
	}
	if !effectstest.HasKeywordOf(g, voice, "protection from black") {
		t.Error("no protection from the chosen colour (black)")
	}
	if effectstest.HasKeywordOf(g, voice, "protection from red") {
		t.Error("protection from a colour that was not chosen")
	}
}

func TestGapB139DralnusCrusadeMakesEveryGoblinBlackAndAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := gapB139Coloured(g, me.ID, "My Goblin", "R", "Creature — Goblin")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, mine).Colors = []string{"R"}
		g.RecomputeLayersIfStaleLocked()
	})
	effectstest.B12Push(g, me.ID, "Dralnu's Crusade", "Enchantment", gapB139CrusadeOracle, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my Goblin = %d/%d, want 3/3 (a 2/2 and +1/+1)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("their Goblin = %d/%d, want 3/3 (a 2/2 and +1/+1)", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear = %d/%d, want 2/2", p, tough)
	}
	g.ReadSnapshot(func() {
		c := effectstest.FindBattlefieldCardForTest(g, mine)
		e := c.Effective()
		if len(e.Colors) != 1 || e.Colors[0] != "B" {
			t.Errorf("my Goblin's colours = %v, want [B]", e.Colors)
		}
		if !c.HasSubtype("Zombie") || !c.HasSubtype("Goblin") {
			t.Errorf("my Goblin's subtypes = %v, want Goblin and Zombie", e.Subtypes)
		}
		if effectstest.FindBattlefieldCardForTest(g, bear).HasSubtype("Zombie") {
			t.Error("the Bear is a Zombie")
		}
	})
}

func TestGapB139InvigoratingBoonCountersOnAnyPlayersCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Invigorating Boon", OracleID: gapB139BoonOracle, TypeLine: "Enchantment"})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CycleFromHand(t, g, "Akroma's Blessing", "Instant", "a3fc6888-0a4b-4e1d-b709-042b8b6b1a04", "{W}")
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Bear has no +1/+1 counter: %+v", c)
	}
}
