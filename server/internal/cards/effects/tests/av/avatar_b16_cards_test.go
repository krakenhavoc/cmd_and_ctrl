package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b16_cards_test.go — Avatar: The Last Airbender batches 16 and 17
// (#4459).

const (
	avB16SeismicSenseOracle   = "9d586c74-b1dd-4fcb-a24a-c703d250fa63"
	avB16SparringDummyOracle  = "036b7b89-e49a-4560-9de9-1d4d71f3f666"
	avB16TophOracle           = "2ce20f43-9357-448f-bb40-7315feb08433"
	avB16TrueAncestryOracle   = "26c2dd96-07e7-4294-9431-96d5aeff3761"
	avB16TurtleDuckOracle     = "e93782e3-1b93-459a-a7de-06c3b6591e6f"
	avB16CabbageOracle        = "d8f7d4c8-bc50-4bcf-b930-797dcb3c2a21"
	avB16WalltopOracle        = "fb3a0910-a582-41ef-b5b9-3cda1f1cd5ce"
	avB16AirNomadOracle       = "9e830359-ab92-4bfd-82b3-d9c67ac80cf7"
	avB16CatOwlOracle         = "ba6840a6-b6fd-4b91-ac1a-a61b7b15849a"
	avB16AdministratorOracle  = "24a691f3-88dc-47a5-b5cd-e02bb8026470"
	avB16DragonflyOracle      = "83abf1d0-04e4-49e5-bf63-59f77829ffd1"
	avB16EKSoldierOracle      = "4289f2d4-3db2-4dd1-9c18-af3f78309844"
	avB16LieutenantOracle     = "9da9248d-1201-447f-b6c2-2b64af4f71c4"
	avB16WrestlersOracle      = "1de17747-27e8-40bb-b6f7-6d5402d4705c"
	avB16RuffiansOracle       = "0380e99c-18f9-4472-8223-bb208e5f0963"
	avB16FoggySwampOracle     = "faecfb8d-2f63-4b98-a842-b11448f0d538"
	avB16GuruPathikOracle     = "4dcd7b78-ce5a-43eb-a16d-3b9ea411795d"
	avB16HeiBaiOracle         = "8f709660-bb26-409f-9650-e5228ac8db7e"
	avB16HerbalistOracle      = "9928773d-3f35-4b0e-965d-4233e04398fb"
	avB16JetOracle            = "77cd2246-6906-4dce-8580-4c9d2e68c477"
	avB16LessonTypeLine       = "Sorcery — Lesson"
	avB16AllyTypeLine         = "Creature — Human Soldier Ally"
	avB16CabbageTypeLine      = "Creature — Human Citizen"
	avB16SpiritKeeperTypeLine = "Creature — Human Druid Ally"
)

func TestAvatarB16CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB16SeismicSenseOracle:  "Seismic Sense",
		avB16SparringDummyOracle: "Sparring Dummy",
		avB16TophOracle:          "Toph, the Blind Bandit",
		avB16TrueAncestryOracle:  "True Ancestry",
		avB16TurtleDuckOracle:    "Turtle-Duck",
		avB16CabbageOracle:       "Unlucky Cabbage Merchant",
		avB16WalltopOracle:       "Walltop Sentries",
		avB16AirNomadOracle:      "Air Nomad Legacy",
		avB16CatOwlOracle:        "Cat-Owl",
		avB16AdministratorOracle: "Cruel Administrator",
		avB16DragonflyOracle:     "Dragonfly Swarm",
		avB16EKSoldierOracle:     "Earth Kingdom Soldier",
		avB16LieutenantOracle:    "Earth King's Lieutenant",
		avB16WrestlersOracle:     "Earth Rumble Wrestlers",
		avB16RuffiansOracle:      "Earth Village Ruffians",
		avB16FoggySwampOracle:    "Foggy Swamp Spirit Keeper",
		avB16GuruPathikOracle:    "Guru Pathik",
		avB16HeiBaiOracle:        "Hei Bai, Spirit of Balance",
		avB16HerbalistOracle:     "Hermitic Herbalist",
		avB16JetOracle:           "Jet, Freedom Fighter",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func avB16Refs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

func avB16Lesson(g *game.Game, p *game.Player) uuid.UUID {
	return effectstest.PushToGraveyard(g, p, game.Card{Name: "Some Lesson", TypeLine: avB16LessonTypeLine, ManaCost: "{1}"})
}

func avB16InHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func avB16InLibrary(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Library.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// --- Seismic Sense --------------------------------------------------------------

func TestAvatarB16SeismicSenseLooksAtOneCardPerLandAndTakesACreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Shock", "Instant"},
		{"Bear", "Creature — Bear"},
		{"Mountain", "Basic Land — Mountain"},
	})

	effectstest.CastCatalogSpell(t, g, "Seismic Sense", avB16LessonTypeLine, avB16SeismicSenseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no Seismic Sense prompt")
	}
	// Two lands: the top two only, and only the creature card of them.
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[1] || prompt.ChooseMax != 1 || prompt.ChooseMin != 0 {
		t.Fatalf("offered %v (min %d, max %d), want the Bear alone, optional", prompt.ChooseCards, prompt.ChooseMin, prompt.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[1])
	if !avB16InHand(me, top[1]) {
		t.Error("the Bear is not in hand")
	}
	lib := me.Library.Cards
	if len(lib) == 0 || lib[0].InstanceID != top[0] || lib[len(lib)-1].InstanceID != top[2] {
		t.Error("the Shock should be on the bottom and the Mountain, never looked at, on top")
	}
}

// --- Sparring Dummy -------------------------------------------------------------

func TestAvatarB16SparringDummyMillsALandIntoHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dummy := effectstest.B27Push(g, me.ID, "Sparring Dummy", "Artifact Creature — Scarecrow", avB16SparringDummyOracle, "{1}{G}", 1, 3, "G")
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{{"Forest", "Basic Land — Forest"}})
	life := me.Life

	effectstest.B16Activate(t, g, me.ID, dummy, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil || len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != top[0] || prompt.ChooseMin != 0 {
		t.Fatalf("want an optional offer of the milled Forest: %+v", prompt)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[0])
	if !avB16InHand(me, top[0]) {
		t.Error("the milled land did not go to hand")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d (no Lesson milled)", me.Life, life)
	}
	if err := g.ActivateCatalogAbility(me.ID, dummy, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated again while tapped")
	}
}

func TestAvatarB16SparringDummyGainsTwoForAMilledLesson(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dummy := effectstest.B27Push(g, me.ID, "Sparring Dummy", "Artifact Creature — Scarecrow", avB16SparringDummyOracle, "{1}{G}", 1, 3, "G")
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{{"Seismic Sense", avB16LessonTypeLine}})
	life := me.Life

	effectstest.B16Activate(t, g, me.ID, dummy, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChooseCards(g, me.ID); p != nil {
		t.Fatalf("a milled Lesson offered a land pick: %+v", p)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if avB16InHand(me, top[0]) {
		t.Error("the Lesson went to hand")
	}
}

// --- Toph, the Blind Bandit -----------------------------------------------------

func TestAvatarB16TophEarthbendsTwoAndCountsCountersOnLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	mountain := effectstest.PushEarthbendLand(g, me.ID, "Mountain", "Basic Land — Mountain")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 3) })

	toph := effectstest.AvB16CastCreature(t, g, "Toph, the Blind Bandit", "Legendary Creature — Human Warrior Ally", avB16TophOracle, 0, 3)
	if p, _ := effectstest.PtOf(t, g, toph); p != 0 {
		t.Fatalf("Toph's power %d before the earthbend, want 0 (a creature's counters do not count)", p)
	}
	effectstest.PickCard(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	power, toughness, creatureLand, hasty := effectstest.EarthbentBody(t, g, forest)
	if power != 2 || toughness != 2 || !creatureLand || !hasty {
		t.Fatalf("earthbent land = %d/%d creature land %v haste %v, want a 2/2 hasty land creature", power, toughness, creatureLand, hasty)
	}
	if p, _ := effectstest.PtOf(t, g, toph); p != 2 {
		t.Errorf("Toph's power %d, want 2", p)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mountain, game.CounterPlusOne, 1) })
	if p, _ := effectstest.PtOf(t, g, toph); p != 3 {
		t.Errorf("Toph's power %d with a counter on a second land, want 3", p)
	}
}

// --- True Ancestry --------------------------------------------------------------

func TestAvatarB16TrueAncestryReturnsAPermanentCardAndMakesAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushToGraveyard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	shock := effectstest.PushToGraveyard(g, me, game.Card{Name: "Shock", TypeLine: "Instant"})

	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "True Ancestry", TypeLine: avB16LessonTypeLine, OracleID: avB16TrueAncestryOracle},
		game.CastSpellParams{Strict: true, Targets: avB16Refs(shock)}); err == nil {
		t.Fatal("an instant card was a legal target")
	}
	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "True Ancestry", TypeLine: avB16LessonTypeLine, OracleID: avB16TrueAncestryOracle},
		game.CastSpellParams{Targets: avB16Refs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !avB16InHand(me, bear) {
		t.Error("the creature card did not return to hand")
	}
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestAvatarB16TrueAncestryWithNoTargetStillMakesAClue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	if _, err := effectstest.AvB16Cast(t, g, game.Card{Name: "True Ancestry", TypeLine: avB16LessonTypeLine, OracleID: avB16TrueAncestryOracle},
		game.CastSpellParams{}); err != nil {
		t.Fatalf("cast with no target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Turtle-Duck ----------------------------------------------------------------

func TestAvatarB16TurtleDuckSetsBasePowerFourAndGainsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	duck := effectstest.B27Push(g, me.ID, "Turtle-Duck", "Creature — Turtle Bird", avB16TurtleDuckOracle, "{G}", 0, 4, "G")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(duck, game.CounterPlusOne, 1) })

	avB7Pool(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, duck, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {2}")
	}
	avB7Pool(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, duck, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, duck); p != 5 || tough != 5 {
		t.Errorf("Turtle-Duck is %d/%d, want 5/5 (base 4/4 plus its counter)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, duck, "trample") {
		t.Error("no trample")
	}
}

// --- Unlucky Cabbage Merchant ---------------------------------------------------

func TestAvatarB16UnluckyCabbageMerchantRampsAndGoesIntoTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Filler", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	forest := ids[1]
	merchant := effectstest.AvB16CastCreature(t, g, "Unlucky Cabbage Merchant", avB16CabbageTypeLine, avB16CabbageOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	food := effectstest.BattlefieldIDNamed(g, me.ID, "Food")
	if food == uuid.Nil {
		t.Fatal("the Merchant made no Food")
	}
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, forest)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) || !effectstest.B16Tapped(t, g, forest) {
		t.Fatal("the basic land should be on the battlefield tapped")
	}
	if g.Battlefield.Contains(merchant) || !avB16InLibrary(me, merchant) {
		t.Error("the Merchant should be in its owner's library")
	}
}

func TestAvatarB16UnluckyCabbageMerchantDecliningKeepsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	merchant := effectstest.AvB16CastCreature(t, g, "Unlucky Cabbage Merchant", avB16CabbageTypeLine, avB16CabbageOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	food := effectstest.BattlefieldIDNamed(g, me.ID, "Food")
	effectstest.B06AddMana(me, "C", "C")
	effectstest.B16Activate(t, g, me.ID, food, 0, game.ActivateAbilityParams{})
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("declining still searched")
	}
	if !g.Battlefield.Contains(merchant) {
		t.Error("the Merchant left although nothing was searched")
	}
}

// --- Walltop Sentries and Dragonfly Swarm ---------------------------------------

func TestAvatarB16WalltopSentriesGainsTwoOnlyWithALesson(t *testing.T) {
	for _, lesson := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if lesson {
			avB16Lesson(g, me)
		}
		sentries := effectstest.B27Push(g, me.ID, "Walltop Sentries", avB16AllyTypeLine, avB16WalltopOracle, "{2}{G}", 2, 3, "G")
		life := me.Life
		effectstest.KillCreature(t, g, me.ID, sentries)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if lesson {
			want += 2
		}
		if me.Life != want {
			t.Errorf("lesson %v: life %d, want %d", lesson, me.Life, want)
		}
	}
}

func TestAvatarB16WalltopSentriesRechecksTheLessonAsItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lesson := avB16Lesson(g, me)
	sentries := effectstest.B27Push(g, me.ID, "Walltop Sentries", avB16AllyTypeLine, avB16WalltopOracle, "{2}{G}", 2, 3, "G")
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, sentries)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, lesson); err != nil {
		t.Fatalf("exile the Lesson: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: the Lesson was gone as it resolved", me.Life, life)
	}
}

func TestAvatarB16DragonflySwarmCountsNoncreatureNonlandCardsAndDrawsWithALesson(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	swarm := effectstest.B27Push(g, me.ID, "Dragonfly Swarm", "Creature — Dragon Insect", avB16DragonflyOracle, "{1}{U}{R}", 0, 3, "U", "R")
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Shock", TypeLine: "Instant"})
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PushToGraveyard(g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushToGraveyard(g, g.Seats[1], game.Card{Name: "Their Shock", TypeLine: "Instant"})
	avB16Lesson(g, me)
	if p, _ := effectstest.PtOf(t, g, swarm); p != 2 {
		t.Fatalf("power %d, want 2 (my instant and my Lesson; not creatures, lands or another player's cards)", p)
	}
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, swarm)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Air Nomad Legacy -----------------------------------------------------------

func TestAvatarB16AirNomadLegacyMakesAClueAndPumpsFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	flier := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	walker := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Air Nomad Legacy", "Enchantment", avB16AirNomadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Clue", me.ID); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, flier); p != 2 || tough != 2 {
		t.Errorf("flier %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, walker); p != 2 || tough != 2 {
		t.Errorf("non-flier %d/%d, want 2/2", p, tough)
	}
}

// --- Cat-Owl --------------------------------------------------------------------

func TestAvatarB16CatOwlUntapsTargetArtifactOrCreatureOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	owl := effectstest.B27Push(g, me.ID, "Cat-Owl", "Creature — Cat Bird", avB16CatOwlOracle, "{3}{W}", 3, 3, "W", "U")
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.TapForTest(t, g, relic)
	effectstest.DeclareAttack(t, g, opp.ID, owl)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, relic)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, relic) {
		t.Error("the artifact is still tapped")
	}
}

// --- Cruel Administrator --------------------------------------------------------

func TestAvatarB16CruelAdministratorRaidCounterAndSoldierOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.AvB16CastCreature(t, g, "Cruel Administrator", "Creature — Human Soldier", avB16AdministratorOracle, 5, 4)
	if n := effectstest.CountersOn(g, first, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters without an attack this turn, want 0", n)
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == first {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	soldier := effectstest.BattlefieldIDNamed(g, me.ID, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier token on attack")
	}
	c, _ := g.LookupCardForEffect(soldier)
	if c.Power != 2 || c.Toughness != 2 || len(c.Colors) != 1 || c.Colors[0] != "R" || c.AttackingTarget != uuid.Nil {
		t.Errorf("Soldier %d/%d %v attacking %v, want a 2/2 red token not attacking", c.Power, c.Toughness, c.Colors, c.AttackingTarget)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	second := effectstest.AvB16CastCreature(t, g, "Cruel Administrator", "Creature — Human Soldier", avB16AdministratorOracle, 5, 4)
	if n := effectstest.CountersOn(g, second, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after attacking this turn, want 1", n)
	}
}

func TestAvatarB16CruelAdministratorsSoldierFirebends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.AvB16SoldierFirebendingToken(), 1) })
	soldier := effectstest.BattlefieldIDNamed(g, me.ID, "Soldier")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == soldier {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	effectstest.DeclareAttack(t, g, opp.ID, soldier)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v after the Soldier attacked, want {R}", got)
	}
}

// --- Earth Kingdom Soldier ------------------------------------------------------

func TestAvatarB16EarthKingdomSoldierPutsACounterOnUpToTwoOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 1)
	effectstest.AvB16CastCreature(t, g, "Earth Kingdom Soldier", "Creature — Human Soldier", avB16EKSoldierOracle, 3, 4)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, c := range p.ChooseCards {
		if c == theirs {
			t.Error("an opponent's creature is offered")
		}
	}
	effectstest.B17PickCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, a, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, b, game.CounterPlusOne) != 1 {
		t.Error("each target should have one +1/+1 counter")
	}
	if effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("the opponent's creature got a counter")
	}
}

// --- Earth King's Lieutenant ----------------------------------------------------

func TestAvatarB16EarthKingsLieutenantCountersOtherAlliesAndGrowsOnAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Ally", TypeLine: avB16AllyTypeLine, Power: 1, Toughness: 1})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	lt := effectstest.AvB16CastCreature(t, g, "Earth King's Lieutenant", avB16AllyTypeLine, avB16LieutenantOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, ally, game.CounterPlusOne) != 1 {
		t.Error("the other Ally got no counter")
	}
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 0 || effectstest.CountersOn(g, lt, game.CounterPlusOne) != 0 {
		t.Error("a non-Ally or the Lieutenant itself got a counter from the enters trigger")
	}
	effectstest.AvB16CastCreature(t, g, "Another Ally", avB16AllyTypeLine, "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, lt, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after another Ally entered, want 1", n)
	}
	effectstest.AvB16CastCreature(t, g, "Another Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, lt, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a non-Ally entered, want 1", n)
	}
}

// --- Earth Rumble Wrestlers -----------------------------------------------------

func TestAvatarB16EarthRumbleWrestlersNeedALandCreatureOrALandThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wrestlers := effectstest.B27Push(g, me.ID, "Earth Rumble Wrestlers", "Creature — Human Warrior Performer", avB16WrestlersOracle, "{3}{R}", 3, 4, "R", "G")
	effectstest.PushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	if p, _ := effectstest.PtOf(t, g, wrestlers); p != 3 || effectstest.HasEffectiveKeyword(t, g, wrestlers, "trample") {
		t.Fatalf("power %d with no land creature and no land this turn, want 3 and no trample", p)
	}
	arbor := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if p, _ := effectstest.PtOf(t, g, wrestlers); p != 4 || !effectstest.HasEffectiveKeyword(t, g, wrestlers, "trample") {
		t.Errorf("power %d with a land creature, want 4 and trample", p)
	}
	effectstest.KillCreature(t, g, me.ID, arbor)
	if p, _ := effectstest.PtOf(t, g, wrestlers); p != 3 {
		t.Fatalf("power %d once the land creature is gone, want 3", p)
	}
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, land); err != nil {
		t.Fatalf("put the land onto the battlefield: %v", err)
	}
	if p, _ := effectstest.PtOf(t, g, wrestlers); p != 4 || !effectstest.HasEffectiveKeyword(t, g, wrestlers, "trample") {
		t.Errorf("power %d after a land entered this turn, want 4 and trample", p)
	}
}

// --- Earth Village Ruffians -----------------------------------------------------

func TestAvatarB16EarthVillageRuffiansEarthbendTwoWhenTheyDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ruffians := effectstest.B27Push(g, me.ID, "Earth Village Ruffians", "Creature — Human Soldier Rogue", avB16RuffiansOracle, "{2}{B}", 3, 1, "B", "G")
	land := effectstest.PushEarthbendLand(g, me.ID, "Swamp", "Basic Land — Swamp")
	effectstest.KillCreature(t, g, me.ID, ruffians)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	power, toughness, creatureLand, hasty := effectstest.EarthbentBody(t, g, land)
	if power != 2 || toughness != 2 || !creatureLand || !hasty {
		t.Errorf("earthbent land = %d/%d creature land %v haste %v, want a 2/2 hasty land creature", power, toughness, creatureLand, hasty)
	}
}

// --- Foggy Swamp Spirit Keeper --------------------------------------------------

func TestAvatarB16FoggySwampSpiritKeeperMakesASpiritOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B27Push(g, me.ID, "Foggy Swamp Spirit Keeper", avB16SpiritKeeperTypeLine, avB16FoggySwampOracle, "{1}{U}{B}", 2, 4, "U", "B")
	for i, want := range []int{0, 1, 1} {
		effectstest.HobA5Draw(t, g, me, nil)
		if got := effectstest.CountOnBattlefield(g, "Spirit", me.ID); got != want {
			t.Fatalf("%d Spirits after draw %d, want %d", got, i+1, want)
		}
	}
	spirit, _ := g.LookupCardForEffect(effectstest.BattlefieldIDNamed(g, me.ID, "Spirit"))
	if spirit.TokenKey != effects.KurukSpiritToken().TokenKey {
		t.Errorf("the Spirit is %q, want Kuruk's printed Spirit", spirit.TokenKey)
	}
}

// --- Guru Pathik ----------------------------------------------------------------

func TestAvatarB16GuruPathikDigsForALessonSagaOrShrine(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Bear", "Creature — Bear"},
		{"Some Lesson", avB16LessonTypeLine},
		{"A Saga", "Enchantment — Saga"},
		{"A Shrine", "Enchantment — Shrine"},
		{"Forest", "Basic Land — Forest"},
	})
	effectstest.AvB16CastCreature(t, g, "Guru Pathik", "Legendary Creature — Human Monk Ally", avB16GuruPathikOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no Guru Pathik prompt")
	}
	want := map[uuid.UUID]bool{top[1]: true, top[2]: true, top[3]: true}
	if len(prompt.ChooseCards) != 3 || prompt.ChooseMax != 1 {
		t.Fatalf("offered %v (max %d), want the Lesson, Saga and Shrine, one of them", prompt.ChooseCards, prompt.ChooseMax)
	}
	for _, id := range prompt.ChooseCards {
		if !want[id] {
			t.Errorf("offered %s", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[3])
	if !avB16InHand(me, top[3]) {
		t.Error("the Shrine is not in hand")
	}
}

func TestAvatarB16GuruPathikCountersAnotherCreatureOnALessonCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	guru := effectstest.B27Push(g, me.ID, "Guru Pathik", "Legendary Creature — Human Monk Ally", avB16GuruPathikOracle, "{2}{G}{U}", 2, 4, "G", "U")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 0 {
		t.Fatal("a non-Lesson sorcery triggered Guru Pathik")
	}
	effectstest.CastCatalogSpell(t, g, "Some Lesson", avB16LessonTypeLine, "", nil)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.ChooseCards, guru) {
			t.Error("Guru Pathik offered himself as the other target")
		}
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, guru, game.CounterPlusOne) != 0 {
		t.Error("the counter should be on the other creature")
	}
}

// --- Hei Bai, Spirit of Balance -------------------------------------------------

func TestAvatarB16HeiBaiSacrificesToGrowAndPassesItsCountersOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hei := effectstest.AvB16CastCreature(t, g, "Hei Bai, Spirit of Balance", "Legendary Creature — Bear Spirit", avB16HeiBaiOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || effectstest.HasID(prompt.ChooseCards, hei) || !effectstest.HasID(prompt.ChooseCards, relic) {
		t.Fatalf("want a may-sacrifice of another creature or artifact: %+v", prompt)
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Fatal("the artifact was not sacrificed")
	}
	if n := effectstest.CountersOn(g, hei, game.CounterPlusOne); n != 2 {
		t.Fatalf("%d counters on Hei Bai, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hei, "stun", 1) })

	effectstest.KillCreature(t, g, me.ID, hei)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 2 || effectstest.CountersOn(g, bear, "stun") != 1 {
		t.Errorf("the Bear has %d +1/+1 and %d stun counters, want Hei Bai's 2 and 1",
			effectstest.CountersOn(g, bear, game.CounterPlusOne), effectstest.CountersOn(g, bear, "stun"))
	}
}

func TestAvatarB16HeiBaiDecliningAddsNoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	hei := effectstest.AvB16CastCreature(t, g, "Hei Bai, Spirit of Balance", "Legendary Creature — Bear Spirit", avB16HeiBaiOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || effectstest.CountersOn(g, hei, game.CounterPlusOne) != 0 {
		t.Error("declining should sacrifice nothing and add no counters")
	}
}

// --- Hermitic Herbalist ---------------------------------------------------------

func TestAvatarB16HermiticHerbalistsTwoManaPaysOnlyForLessonSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herbalist := effectstest.B27Push(g, me.ID, "Hermitic Herbalist", avB16SpiritKeeperTypeLine, avB16HerbalistOracle, "{G}{U}", 2, 3, "G", "U")
	if err := g.ActivateManaAbility(me.ID, herbalist, 1, game.ManaAbilityParams{Colors: []string{"G", "U"}}); err != nil {
		t.Fatalf("tap for two: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 {
		t.Fatalf("pool %v, want two mana", got)
	}
	cost, _ := game.ParseCost("{G}{U}")
	lesson := game.Card{Name: "Some Lesson", TypeLine: avB16LessonTypeLine, ManaCost: "{G}{U}"}
	plain := game.Card{Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{G}{U}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(lesson)) {
		t.Error("refused a Lesson spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(plain)) {
		t.Error("paid for a spell that is not a Lesson")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(lesson)) {
		t.Error("paid for an ability")
	}
}

func TestAvatarB16HermiticHerbalistsOneManaIsUnrestricted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	herbalist := effectstest.B27Push(g, me.ID, "Hermitic Herbalist", avB16SpiritKeeperTypeLine, avB16HerbalistOracle, "{G}{U}", 2, 3, "G", "U")
	if err := g.ActivateManaAbility(me.ID, herbalist, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap for one: %v", err)
	}
	cost, _ := game.ParseCost("{R}")
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(game.Card{Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"})) {
		t.Error("the any-color mana refused an ordinary spell")
	}
	if err := g.ActivateManaAbility(me.ID, herbalist, 1, game.ManaAbilityParams{Colors: []string{"G", "U"}}); err == nil {
		t.Error("tapped for mana a second time")
	}
}

// --- Jet, Freedom Fighter -------------------------------------------------------

func TestAvatarB16JetDealsDamagePerCreatureAndCountersTwoOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 1, 1)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 1, 4)
	bigger := effectstest.PushVanillaCreature(g, opp.ID, "Bigger", 1, 3)
	jet := effectstest.AvB16CastCreature(t, g, "Jet, Freedom Fighter", "Legendary Creature — Human Rebel Ally", avB16JetOracle, 3, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Jet")
	}
	if effectstest.HasID(p.ChooseCards, a) || effectstest.HasID(p.ChooseCards, jet) {
		t.Error("a creature I control is offered")
	}
	effectstest.PickCard(t, g, me.ID, bigger)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bigger) {
		t.Error("the 1/3 survived 3 damage (three creatures)")
	}
	_ = big

	effectstest.KillCreature(t, g, me.ID, jet)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.B17PickCards(t, g, me.ID, a, big)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, a, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, big, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, b, game.CounterPlusOne) != 0 {
		t.Error("the two targets (any controller) should have one counter each")
	}
}
