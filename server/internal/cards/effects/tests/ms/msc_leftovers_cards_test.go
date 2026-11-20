package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_leftovers_cards_test.go — the Marvel Super Heroes Commander leftovers
// that waited on the connive event, the Vibranium and Robot Villain tokens
// and the hand-attacker put (#4455), and the Heroic Return fix.

const (
	mscLoGloriousOracle  = "182baad0-ec6e-4e93-9dfc-7c80b37005de"
	mscLoValentinaOracle = "58ce3c79-8903-41be-8494-56d7953074ad"
	mscLoUltronOracle    = "436501ca-0998-43fc-975f-87c2d34caab3"
	mscLoMechOracle      = "1bea6d02-186e-46d8-b788-d4cc37dbb480"
	mscLoDroneOracle     = "51c1f21d-1512-4885-bfdc-58790cf77d51"
	mscLoFuryOracle      = "b20860da-2e71-414a-a512-1cce7ba64079"
)

func TestMscLeftoversEveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscLoGloriousOracle:  "Glorious Purpose",
		mscLoValentinaOracle: "Valentina Allegra de Fontaine",
		mscLoUltronOracle:    "Ultron, Unlimited",
		mscLoMechOracle:      "Vibranium Mining Mech",
		mscLoDroneOracle:     "Ultron Drone",
		mscLoFuryOracle:      "Nick Fury, Spymaster",
		mscLoKangOracle:      "Kang, Temporal Tyrant",
	}
	for oracle, name := range want {
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

// mscLoConnive makes `creature` connive for `controller`, as an effect does.
func mscLoConnive(t *testing.T, g *game.Game, controller, creature uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = effects.ConniveEach{Controller: controller, Creatures: []uuid.UUID{creature}}.Apply(effects.NewContext(g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: controller, SourceCardID: creature,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveEach: %v", err)
	}
}

func mscLoHasSubtype(g *game.Game, id uuid.UUID, subtype string) bool {
	has := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				has = g.Battlefield.Cards[i].HasSubtype(subtype)
			}
		}
	})
	return has
}

// --- Glorious Purpose --------------------------------------------------------------------

func TestMscLoGloriousPurposeCountersTheConniverAndItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	purpose := effectstest.PushCatalogPermanent(g, me.ID, "Glorious Purpose", "Enchantment — Plan", mscLoGloriousOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })

	land := effectstest.HandCardFull(opp, "Forest", "Basic Land — Forest", "", "", nil)
	mscLoConnive(t, g, opp.ID, theirs)
	effectstest.AnswerDiscard(t, g, opp.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, purpose); c.Counters["plan"] != 0 || effectstest.PlusOnes(g, theirs) != 0 {
		t.Fatalf("an opponent's connive counted: plan %d, their bear %d", c.Counters["plan"], effectstest.PlusOnes(g, theirs))
	}

	land = effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	mscLoConnive(t, g, me.ID, mine)
	effectstest.AnswerDiscard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mine); got != 1 {
		t.Errorf("the conniving creature has %d +1/+1 counters, want 1 (a land was discarded, so none from connive)", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, purpose); c.Counters["plan"] != 1 {
		t.Errorf("plan counters = %d, want 1", c.Counters["plan"])
	}
}

// mscLoSixthPlanCounter sets Glorious Purpose at five plan counters over a
// library whose top four after the connive's draw are known, then connives
// for the sixth.
func mscLoSixthPlanCounter(t *testing.T, g *game.Game) (purpose uuid.UUID, land, s1, s2, s3 uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	purpose = effectstest.PushCatalogPermanent(g, me.ID, "Glorious Purpose", "Enchantment — Plan", mscLoGloriousOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(purpose, "plan", 5) })
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	me.Library.Cards = nil
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "Drawn Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Exiled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Exiled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Exiled Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		game.Card{Name: "Exiled Imp", TypeLine: "Creature — Imp", ManaCost: "{B}", Power: 1, Toughness: 1},
		game.Card{Name: "Stays In The Library", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 1, Toughness: 1},
	)
	mscLoConnive(t, g, me.ID, bear)
	effectstest.AnswerDiscard(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	return purpose, ids[1], ids[2], ids[3], ids[4]
}

func TestMscLoGloriousPurposeSixthCounterSacrificesExilesFourCastsAnyNumberAndTheRestGoToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	purpose, land, s1, s2, s3 := mscLoSixthPlanCounter(t, g)
	if g.Battlefield.Contains(purpose) {
		t.Fatal("Glorious Purpose was not sacrificed at the sixth plan counter")
	}
	for _, id := range []uuid.UUID{land, s1, s2, s3} {
		if !effectstest.InExile(g, id) {
			t.Fatalf("card %s was not exiled from the library", id)
		}
	}
	if me.Library.Size() < 1 {
		t.Fatal("the library is empty: more than four cards were exiled")
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, s1, s2, s3)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, s1, game.CastSpellParams{}); err != nil {
		t.Fatalf("first free cast: %v", err)
	}
	// Any number: a second cast is offered, and the Wurm costs {4}{G} free.
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, s2, game.CastSpellParams{}); err != nil {
		t.Fatalf("second free cast: %v", err)
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	// The rest: the land and the spell not cast, in hand; the two cast are not.
	for _, id := range []uuid.UUID{land, s3} {
		if !effectstest.InHandOf(me, id) {
			t.Errorf("card %s is not in hand", id)
		}
	}
	for _, id := range []uuid.UUID{s1, s2} {
		if effectstest.InHandOf(me, id) || effectstest.InExile(g, id) {
			t.Errorf("cast card %s went to hand or stayed in exile", id)
		}
	}
}

func TestMscLoGloriousPurposeDecliningEveryCastPutsAllFourIntoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, land, s1, s2, s3 := mscLoSixthPlanCounter(t, g)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	for _, id := range []uuid.UUID{land, s1, s2, s3} {
		if !effectstest.InHandOf(me, id) {
			t.Errorf("card %s is not in hand", id)
		}
	}
}

func TestMscLoGloriousPurposeFifthCounterDoesNothingYet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	purpose := effectstest.PushCatalogPermanent(g, me.ID, "Glorious Purpose", "Enchantment — Plan", mscLoGloriousOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(purpose, "plan", 4) })
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	libBefore := me.Library.Size()
	mscLoConnive(t, g, me.ID, bear)
	effectstest.AnswerDiscard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, purpose); !ok || c.Counters["plan"] != 5 {
		t.Fatalf("Glorious Purpose after the fifth counter: on battlefield %v, %v", ok, c.Counters)
	}
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d -> %d, want only the connive's draw", libBefore, got)
	}
}

// --- Valentina Allegra de Fontaine ---------------------------------------------------------

func TestMscLoValentinaMakesOtherVillainsYouControlHeroesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	val := effectstest.PushCatalogPermanent(g, me.ID, "Valentina Allegra de Fontaine", "Legendary Creature — Human Spy Villain", mscLoValentinaOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	if !mscLoHasSubtype(g, mine, "Hero") || !mscLoHasSubtype(g, mine, "Villain") {
		t.Error("my other Villain is not a Hero in addition to a Villain")
	}
	if mscLoHasSubtype(g, val, "Hero") {
		t.Error("Valentina made herself a Hero ('other' Villains only)")
	}
	if mscLoHasSubtype(g, bear, "Hero") || mscLoHasSubtype(g, theirs, "Hero") {
		t.Error("a non-Villain or an opponent's Villain became a Hero")
	}
}

func mscLoValentinaTable(t *testing.T) (g *game.Game, me *game.Player, val, villain uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	val = effectstest.PushCatalogPermanent(g, me.ID, "Valentina Allegra de Fontaine", "Legendary Creature — Human Spy Villain", mscLoValentinaOracle, false)
	villain = effectstest.B12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, val, villain
}

func TestMscLoValentinaTapsToMakeAHeroConniveAtSorcerySpeed(t *testing.T) {
	g, me, val, villain := mscLoValentinaTable(t)
	spell := effectstest.HandCardFull(me, "Discarded Spell", "Instant", "{R}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, val, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(villain), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, villain); got != 1 {
		t.Errorf("the Hero has %d +1/+1 counters, want 1 (a nonland card was discarded)", got)
	}
	if c, _ := effectstest.BattlefieldCard(g, val); !c.Tapped {
		t.Error("Valentina's cost did not tap her")
	}
}

func TestMscLoValentinaRefusesWhatIsNotAHeroYouControl(t *testing.T) {
	g, me, val, villain := mscLoValentinaTable(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirHero := effectstest.B12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	for name, target := range map[string]uuid.UUID{"a Bear": bear, "an opponent's Hero": theirHero, "Valentina herself": val} {
		if err := g.ActivateCatalogAbility(me.ID, val, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target), Strict: true}); err == nil {
			t.Errorf("targeted %s", name)
		}
	}
	// Outside a main phase (sorcery speed).
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, val, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(villain), Strict: true}); err == nil {
		t.Error("activated at the beginning of combat")
	}
}

// --- Ultron, Unlimited -------------------------------------------------------------------

func TestMscLoUltronConnivesOnAttackAndPaysOneForARobotVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ultron := effectstest.B12Push(g, me.ID, "Ultron, Unlimited", "Legendary Artifact Creature — Robot Villain", mscLoUltronOracle, 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	effectstest.DeclareAttack(t, g, opp.ID, ultron)
	effectstest.PassPriorityAroundTable(t, g) // the attack trigger: Ultron connives
	effectstest.AnswerDiscard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g) // the connive trigger asks for {1}
	effectstest.B06AddMana(me, "G")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Robot Villain"); n != 1 {
		t.Errorf("%d Robot Villain tokens, want 1", n)
	}
}

func TestMscLoUltronDecliningTheOneMakesNoToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ultron, Unlimited", "Legendary Artifact Creature — Robot Villain", mscLoUltronOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	land := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	mscLoConnive(t, g, me.ID, bear)
	effectstest.AnswerDiscard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Robot Villain"); n != 0 {
		t.Errorf("%d Robot Villain tokens after declining, want 0", n)
	}
}

func TestMscLoUltronIgnoresAnOpponentsCreatureConniving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Ultron, Unlimited", "Legendary Artifact Creature — Robot Villain", mscLoUltronOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	land := effectstest.HandCardFull(opp, "Forest", "Basic Land — Forest", "", "", nil)
	mscLoConnive(t, g, opp.ID, theirs)
	effectstest.AnswerDiscard(t, g, opp.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PayUnlessFor(g, me.ID) != nil {
		t.Error("Ultron asked for {1} when an opponent's creature connived")
	}
}

// --- Vibranium Mining Mech ---------------------------------------------------------------

func TestMscLoMechMakesATappedVibraniumWhenItEntersAndWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Vibranium Mining Mech", TypeLine: "Artifact — Vehicle", ManaCost: "{4}",
		OracleID: mscLoMechOracle, Power: 6, Toughness: 6})
	vib := msc17NamedOnBattlefield(g, "Vibranium")
	if len(vib) != 1 || !vib[0].Tapped || vib[0].Controller != me.ID {
		t.Fatalf("Vibranium after entering: %+v, want one tapped token under me", vib)
	}
	var mech uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == mscLoMechOracle {
			mech = c.InstanceID
		}
	}
	crew := effectstest.PushCrewerForTest(g, me.ID, "Crew", 2)
	effectstest.Unsick(g, mech)
	if err := g.ActivateCatalogAbility(me.ID, mech, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 2 with a power-2 creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, mech)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(msc17NamedOnBattlefield(g, "Vibranium")); got != 2 {
		t.Errorf("%d Vibranium after attacking, want 2", got)
	}
}

func TestMscLoMechPumpsItselfForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mech := effectstest.PushVehicleForTest(g, me.ID, "Vibranium Mining Mech", mscLoMechOracle, 6, 6)
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the pump activated without {2}")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, mech, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("pump: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	power := 0
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mech {
				power = g.Battlefield.Cards[i].CurrentPower()
			}
		}
	})
	if power != 7 {
		t.Errorf("power = %d, want 7", power)
	}
}

// --- Ultron Drone ------------------------------------------------------------------------

func TestMscLoUltronDronePowerUpGivesTwoCountersAndARobotVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	drone := effectstest.CastPowerUpCreature(t, g, me, "Ultron Drone", "Artifact Creature — Robot Villain", "{3}", "{C}{C}{C}", mscLoDroneOracle, 2)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, drone); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3} ({6} less its {3})", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, drone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, drone); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Robot Villain"); n != 1 {
		t.Errorf("%d Robot Villain tokens, want 1", n)
	}
}

// --- Nick Fury, Spymaster ----------------------------------------------------------------

type mscLoFuryTable struct {
	g              *game.Game
	me, opp        *game.Player
	fury, attacker uuid.UUID
}

func newMscLoFuryTable(t *testing.T) mscLoFuryTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fury := effectstest.B12Push(g, me.ID, "Nick Fury, Spymaster", "Legendary Creature — Human Spy Hero", mscLoFuryOracle, 4, 4)
	attacker := effectstest.B12Creature(g, me.ID, "Lone Attacker", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return mscLoFuryTable{g: g, me: me, opp: opp, fury: fury, attacker: attacker}
}

func TestMscLoNickFuryDrawsThenPutsACheapCreatureInTappedAttackingAndIndestructible(t *testing.T) {
	f := newMscLoFuryTable(t)
	small := effectstest.HandCardFull(f.me, "Small Spy", "Creature — Human Spy", "{2}", "", nil)
	three := effectstest.HandCardFull(f.me, "Three Spy", "Creature — Human Spy", "{1}{U}{B}", "", nil)
	big := effectstest.HandCardFull(f.me, "Big Spy", "Creature — Human Spy", "{3}{U}", "", nil)
	rock := effectstest.HandCardFull(f.me, "Rock", "Artifact", "{1}", "", nil)
	handBefore := f.me.Hand.Size()
	effectstest.DeclareAttack(t, f.g, f.opp.ID, f.attacker)
	effectstest.PassPriorityAroundTable(t, f.g)
	if got := f.me.Hand.Size(); got != handBefore+1 {
		t.Fatalf("hand %d -> %d, want one card drawn", handBefore, got)
	}
	prompt := effectstest.ChooseCardsChoiceFor(f.g, f.me.ID)
	if prompt == nil {
		t.Fatal("no put prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range prompt.ChooseCards {
		offered[id] = true
	}
	if !offered[small] || !offered[three] || offered[big] || offered[rock] {
		t.Fatalf("offered %v: want the MV 2 and 3 creatures only (small %v three %v big %v rock %v)",
			prompt.ChooseCards, small, three, big, rock)
	}
	effectstest.AnswerChooseCards(t, f.g, f.me.ID, three)
	effectstest.AnswerOptionPick(t, f.g, f.me.ID, 0) // the lone attacker's own defender first
	effectstest.PassPriorityAroundTable(t, f.g)
	c, ok := effectstest.BattlefieldCard(f.g, three)
	if !ok {
		t.Fatal("the creature card did not enter")
	}
	if !c.Tapped || c.AttackingTarget != f.opp.ID {
		t.Errorf("it is tapped=%v attacking %v, want tapped and attacking the opponent", c.Tapped, c.AttackingTarget)
	}
	if !effectstest.HasEffectiveKeyword(t, f.g, three, "indestructible") {
		t.Error("it did not gain indestructible")
	}
	if effectstest.InHandOf(f.me, small) == false || effectstest.InHandOf(f.me, big) == false {
		t.Error("a card that was not chosen left the hand")
	}
}

func TestMscLoNickFuryDecliningThePutGainsNothing(t *testing.T) {
	f := newMscLoFuryTable(t)
	small := effectstest.HandCardFull(f.me, "Small Spy", "Creature — Human Spy", "{2}", "", nil)
	effectstest.DeclareAttack(t, f.g, f.opp.ID, f.attacker)
	effectstest.PassPriorityAroundTable(t, f.g)
	effectstest.AnswerChooseCards(t, f.g, f.me.ID)
	effectstest.PassPriorityAroundTable(t, f.g)
	if f.g.Battlefield.Contains(small) {
		t.Error("a declined put still put the creature in")
	}
}

func TestMscLoNickFuryDoesNothingWhenTwoCreaturesAttack(t *testing.T) {
	f := newMscLoFuryTable(t)
	other := effectstest.B12Creature(f.g, f.me.ID, "Second Attacker", "Creature — Bear", 2, 2)
	effectstest.HandCardFull(f.me, "Small Spy", "Creature — Human Spy", "{2}", "", nil)
	handBefore := f.me.Hand.Size()
	effectstest.DeclareAttack(t, f.g, f.opp.ID, f.attacker, other)
	effectstest.PassPriorityAroundTable(t, f.g)
	if got := f.me.Hand.Size(); got != handBefore {
		t.Errorf("hand %d -> %d: Nick Fury drew although two creatures attacked", handBefore, got)
	}
	if effectstest.ChooseCardsChoiceFor(f.g, f.me.ID) != nil {
		t.Error("a put was offered although two creatures attacked")
	}
}

func TestMscLoNickFuryOffersNothingWithNoCheapCreatureInHand(t *testing.T) {
	f := newMscLoFuryTable(t)
	effectstest.HandCardFull(f.me, "Big Spy", "Creature — Human Spy", "{3}{U}", "", nil)
	effectstest.DeclareAttack(t, f.g, f.opp.ID, f.attacker)
	effectstest.PassPriorityAroundTable(t, f.g)
	if effectstest.ChooseCardsChoiceFor(f.g, f.me.ID) != nil {
		t.Error("a put was offered with no creature of mana value 3 or less in hand")
	}
}

// --- Heroic Return (CR 614.12) ----------------------------------------------------------

// The card is judged as it would exist on the battlefield: a Villain card
// that Valentina's static makes a Hero there enters with the counters.
func TestMscLoHeroicReturnJudgesHeroAsItWouldExistOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Valentina Allegra de Fontaine", "Legendary Creature — Human Spy Villain", mscLoValentinaOracle, false)
	villain := effectstest.PushGraveyardPermanent(me, "Fallen Villain", "Creature — Human Villain", "{2}")
	msc15CastHeroicReturn(t, g, me, villain)
	if !g.Battlefield.Contains(villain) {
		t.Fatal("the Villain did not return")
	}
	if got := effectstest.PlusOnes(g, villain); got != 2 {
		t.Errorf("a Villain that is a Hero on the battlefield entered with %d +1/+1 counters, want 2", got)
	}
}

// Without the static the same card is no Hero: nothing in the graveyard or
// on the battlefield makes it one.
func TestMscLoHeroicReturnGivesAVillainNothingWithoutAnythingMakingItAHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	villain := effectstest.PushGraveyardPermanent(me, "Fallen Villain", "Creature — Human Villain", "{2}")
	msc15CastHeroicReturn(t, g, me, villain)
	if got := effectstest.PlusOnes(g, villain); got != 0 {
		t.Errorf("a plain Villain entered with %d +1/+1 counters, want 0", got)
	}
}

// --- Kang, Temporal Tyrant ---------------------------------------------------------------

const mscLoKangOracle = "b43350a8-fc4d-41ed-becd-f0277512bb5f"

func mscLoDraw(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { _ = g.DrawNForEffect(p.ID, 1) })
}

func mscLoCardsDrawn(g *game.Game, p *game.Player) int {
	n := 0
	g.WithWriteLock(func() { n = g.TurnTallyFor(p.ID).CardsDrawn })
	return n
}

func TestMscLoKangDrainsOnYourSecondDrawOfTheTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Kang, Temporal Tyrant", "Legendary Creature — Human Villain", mscLoKangOracle, 3, 4)
	myLife, theirLife := me.Life, opp.Life
	for mscLoCardsDrawn(g, me) < 1 {
		mscLoDraw(g, me)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != myLife || opp.Life != theirLife {
		t.Fatalf("life changed on the first draw of the turn: me %d->%d, opp %d->%d", myLife, me.Life, theirLife, opp.Life)
	}
	// An opponent's second draw is not mine.
	mscLoDraw(g, opp)
	mscLoDraw(g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != theirLife {
		t.Fatalf("an opponent's draws moved life: me %d->%d, opp %d->%d", myLife, me.Life, theirLife, opp.Life)
	}
	mscLoDraw(g, me) // the second
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 || g.Seats[2].Life != theirLife-1 {
		t.Errorf("after the second draw: me %d (want %d), opponents %d/%d/%d (want %d)",
			me.Life, myLife+1, opp.Life, g.Seats[2].Life, g.Seats[3].Life, theirLife-1)
	}
	mscLoDraw(g, me) // the third
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 {
		t.Errorf("the third draw drained again: life %d, want %d", me.Life, myLife+1)
	}
}

func TestMscLoKangConnivesWhenHeAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kang := effectstest.B12Push(g, me.ID, "Kang, Temporal Tyrant", "Legendary Creature — Human Villain", mscLoKangOracle, 3, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := effectstest.HandCardFull(me, "Discarded Spell", "Instant", "{R}", "", nil)
	effectstest.DeclareAttack(t, g, opp.ID, kang)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, kang); got != 1 {
		t.Errorf("Kang has %d +1/+1 counters, want 1 (a nonland card was discarded)", got)
	}
}
