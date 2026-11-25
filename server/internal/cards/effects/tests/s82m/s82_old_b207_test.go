package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b207_test.go — Dominaria Sagas, batch 207 (#5530): each test casts the
// real catalog Saga and lets its chapters trigger through the turn engine.

const (
	oldB207KeldOracle     = "26b41c78-c5ee-4c50-93be-e3acc35ab355"
	oldB207ThranOracle    = "341803d9-fd29-4721-81bf-ae9b7f1c01d2"
	oldB207ChainerOracle  = "3c18113f-8d1e-4c63-a759-8cfc94e90876"
	oldB207MirariOracle   = "56528234-46c0-4fba-972a-0c8011996f8d"
	oldB207IceOracle      = "93e0fd89-8a0d-4d97-aa22-aafe83d53515"
	oldB207FreyaliseOrcl  = "9fe1f3ef-2c30-4ccc-9a67-5349e98e2502"
	oldB207BelzenlokOrcl  = "a3b2e0a2-f4ad-457b-ae70-7732bc3479ff"
	oldB207AntiquitiesOrc = "ab5a9749-5c94-43e0-b9df-bd2c3601936f"
	oldB207EruptionOracle = "df7a3868-4f03-452c-8307-1493977d42ef"
)

const oldB207SagaType = "Enchantment — Saga"

func TestOldB207CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB207KeldOracle:     "The Flame of Keld",
		oldB207ThranOracle:    "Fall of the Thran",
		oldB207ChainerOracle:  "Chainer's Torment",
		oldB207MirariOracle:   "The Mirari Conjecture",
		oldB207IceOracle:      "Time of Ice",
		oldB207FreyaliseOrcl:  "Song of Freyalise",
		oldB207BelzenlokOrcl:  "Rite of Belzenlok",
		oldB207AntiquitiesOrc: "The Antiquities War",
		oldB207EruptionOracle: "The First Eruption",
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

// oldB207Cast casts the Saga for the active seat and lets chapter I resolve.
func oldB207Cast(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastCatalogSpell(t, g, name, oldB207SagaType, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// oldB207NextChapter walks to the controller's next precombat main, where the
// next lore counter is added, and lets the chapter go on the stack and resolve.
func oldB207NextChapter(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
}

// oldB207RedBolt casts a red Lightning Bolt from `who`'s hand at a player and
// resolves everything, answering any copy prompt by keeping the target.
func oldB207RedBolt(t *testing.T, g *game.Game, who *game.Player, victim uuid.UUID, copies int) {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: who.ID, Controller: who.ID})
	if err := g.CastSpell(who.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	answered := 0
	for i := 0; i < 48; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, who.ID); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, who.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, who.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, who.ID, game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered != copies {
		t.Fatalf("answered %d copy prompts, want %d", answered, copies)
	}
}

// oldB207PassAndPick lets the Saga (or the chapter trigger) get to its target
// prompt and answers it when one opens; a chapter with exactly one legal target
// may have chosen it already.
func oldB207PassAndPick(t *testing.T, g *game.Game, who uuid.UUID, target uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityUntilChoice(t, g)
	if effectstest.LatestPickTarget(g, who) != nil {
		effectstest.AnswerPickTarget(t, g, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB207OnBattlefield(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- The Flame of Keld ---

func TestOldB207FlameOfKeldDiscardsDrawsAndAddsTwoToRedDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AuraSeedLibrary(me, 10)
	g.WithWriteLock(func() {
		for range 3 {
			me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	oldB207Cast(t, g, "The Flame of Keld", oldB207KeldOracle)
	if len(me.Hand.Cards) != 0 {
		t.Errorf("hand is %d cards after chapter I, want 0", len(me.Hand.Cards))
	}
	oldB207NextChapter(t, g)
	hand := len(me.Hand.Cards)
	if hand < 2 {
		t.Errorf("hand is %d after chapter II, want at least the two cards drawn", hand)
	}
	oldB207NextChapter(t, g)
	before := effectstest.LifeOf(g, opp.ID)
	oldB207RedBolt(t, g, me, opp.ID, 0)
	if got := effectstest.LifeOf(g, opp.ID); got != before-5 {
		t.Errorf("a red Bolt dealt %d, want 5 (3 plus 2)", before-got)
	}
}

func TestOldB207FlameOfKeldDoesNotBoostAnOpponentsRedSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AuraSeedLibrary(me, 10)
	oldB207Cast(t, g, "The Flame of Keld", oldB207KeldOracle)
	oldB207NextChapter(t, g)
	oldB207NextChapter(t, g)
	before := effectstest.LifeOf(g, me.ID)
	oldB207RedBolt(t, g, opp, me.ID, 0)
	if got := effectstest.LifeOf(g, me.ID); got != before-3 {
		t.Errorf("an opponent's Bolt dealt %d, want 3", before-got)
	}
}

// --- Fall of the Thran ---

func TestOldB207FallOfTheThranDestroysLandsThenEachPlayerReturnsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	var myYard, oppYard []uuid.UUID
	for range 3 {
		myYard = append(myYard, effectstest.GyCard(me, "Mountain", "Basic Land — Mountain", "", ""))
		oppYard = append(oppYard, effectstest.GyCard(opp, "Swamp", "Basic Land — Swamp", "", ""))
	}
	effectstest.CastCatalogSpell(t, g, "Fall of the Thran", oldB207SagaType, oldB207ThranOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a land survived chapter I")
	}
	oldB207NextChapter(t, g)
	effectstest.SlChoose(t, g, me.ID, myYard[0], myYard[1])
	effectstest.SlChoose(t, g, opp.ID, oppYard[0], oppYard[1])
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{myYard[0], myYard[1], oppYard[0], oppYard[1]} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a chosen land did not return")
		}
	}
	if !me.Graveyard.Contains(myYard[2]) || !opp.Graveyard.Contains(oppYard[2]) {
		t.Error("a land that was not chosen left its graveyard")
	}
}

func TestOldB207FallOfTheThranReturnsOnlyWhatAPlayerHas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	one := effectstest.GyCard(me, "Mountain", "Basic Land — Mountain", "", "")
	effectstest.GyCard(opp, "Bear", "Creature — Bear", "", "")
	effectstest.CastCatalogSpell(t, g, "Fall of the Thran", oldB207SagaType, oldB207ThranOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB207NextChapter(t, g)
	effectstest.SlChoose(t, g, me.ID, one)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(one) {
		t.Error("the only land card did not return")
	}
	if len(opp.Graveyard.Cards) != 1 {
		t.Error("a nonland card left the graveyard")
	}
}

// --- Chainer's Torment ---

func TestOldB207ChainersTormentDrainsThenMakesAHalfLifeNightmareHorror(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	oldB207Cast(t, g, "Chainer's Torment", oldB207ChainerOracle)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d after chapter I, want +2", life, me.Life)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != life-2 {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, life-2)
		}
	}
	oldB207NextChapter(t, g)
	if me.Life != life+4 {
		t.Errorf("life is %d after chapter II, want %d", me.Life, life+4)
	}
	g.WithWriteLock(func() { me.Life = 11 })
	oldB207NextChapter(t, g)
	horrors := oldB207OnBattlefield(g, me.ID, "Nightmare Horror")
	if len(horrors) != 1 {
		t.Fatalf("%d Nightmare Horrors, want 1", len(horrors))
	}
	if p, tt := effectstest.PtOf(t, g, horrors[0].InstanceID); p != 6 || tt != 6 {
		t.Errorf("the Horror is %d/%d, want 6/6 (half of 11, rounded up)", p, tt)
	}
	if me.Life != 5 {
		t.Errorf("life is %d, want 5 after the token deals 6 to its controller", me.Life)
	}
}

// --- The Mirari Conjecture ---

func TestOldB207MirariConjectureReturnsAnInstantThenASorceryThenCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	instant := effectstest.GyCard(me, "Old Instant", "Instant", "", "")
	sorcery := effectstest.GyCard(me, "Old Sorcery", "Sorcery", "", "")
	effectstest.CastCatalogSpell(t, g, "The Mirari Conjecture", oldB207SagaType, oldB207MirariOracle, nil)
	oldB207PassAndPick(t, g, me.ID, instant)
	if !me.Hand.Contains(instant) || me.Hand.Contains(sorcery) {
		t.Error("chapter I did not return exactly the instant")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	oldB207PassAndPick(t, g, me.ID, sorcery)
	if !me.Hand.Contains(sorcery) {
		t.Error("chapter II did not return the sorcery")
	}
	oldB207NextChapter(t, g)
	before := effectstest.LifeOf(g, opp.ID)
	oldB207RedBolt(t, g, me, opp.ID, 1)
	oldB207RedBolt(t, g, me, opp.ID, 1)
	if got := effectstest.LifeOf(g, opp.ID); got != before-12 {
		t.Errorf("two copied Bolts dealt %d, want 12", before-got)
	}
}

func TestOldB207MirariConjectureChapterOneRefusesASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sorcery := effectstest.GyCard(me, "Old Sorcery", "Sorcery", "", "")
	instant := effectstest.GyCard(me, "Old Instant", "Instant", "", "")
	effectstest.CastCatalogSpell(t, g, "The Mirari Conjecture", oldB207SagaType, oldB207MirariOracle, nil)
	oldB207PassAndPick(t, g, me.ID, instant)
	if me.Hand.Contains(sorcery) || !me.Hand.Contains(instant) {
		t.Error("chapter I did not return exactly the instant, leaving the sorcery")
	}
}

// --- Time of Ice ---

func TestOldB207TimeOfIceHoldsTappedCreaturesThenBouncesEveryTappedOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	first := effectstest.B12Creature(g, opp.ID, "First Bear", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second Bear", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Time of Ice", oldB207SagaType, oldB207IceOracle, nil)
	oldB207PassAndPick(t, g, me.ID, first)
	effectstest.AssertTapped(t, g, true, first)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	oldB207PassAndPick(t, g, me.ID, second)
	effectstest.AssertTapped(t, g, true, first, second)
	// The held creatures stay tapped through their controller's untap step.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AssertTapped(t, g, true, first, second)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == mine {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{first, second} {
		if !opp.Hand.Contains(id) {
			t.Errorf("a tapped creature was not returned to its owner's hand")
		}
	}
	if !me.Hand.Contains(mine) {
		t.Error("my own tapped creature was not returned (all tapped creatures)")
	}
	if !g.Battlefield.Contains(free) {
		t.Error("an untapped creature was returned")
	}
}

// --- Song of Freyalise ---

func TestOldB207SongOfFreyaliseGrantsManaThenCountersAndKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Song of Freyalise", oldB207SagaType, oldB207FreyaliseOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateManaAbility(me.ID, bear, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("the creature's granted mana ability: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool is %d, want 1", me.ManaPool.Total())
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, late, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a creature that arrived after the chapter has the mana ability")
	}
	oldB207NextChapter(t, g)
	oldB207NextChapter(t, g)
	for _, id := range []uuid.UUID{bear, late} {
		if effectstest.PlusOnes(g, id) != 1 {
			t.Error("a creature did not get its +1/+1 counter")
		}
		for _, kw := range []string{"vigilance", "trample", "indestructible"} {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("a creature lacks %s", kw)
			}
		}
	}
}

// --- Rite of Belzenlok ---

func TestOldB207RiteOfBelzenlokMakesClericsThenADemonThatEatsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AuraSeedLibrary(me, 10)
	oldB207Cast(t, g, "Rite of Belzenlok", oldB207BelzenlokOrcl)
	if n := len(oldB207OnBattlefield(g, me.ID, "Cleric")); n != 2 {
		t.Fatalf("%d Clerics after chapter I, want 2", n)
	}
	oldB207NextChapter(t, g)
	clerics := oldB207OnBattlefield(g, me.ID, "Cleric")
	if len(clerics) != 4 {
		t.Fatalf("%d Clerics after chapter II, want 4", len(clerics))
	}
	if p, tt := effectstest.PtOf(t, g, clerics[0].InstanceID); p != 0 || tt != 1 {
		t.Errorf("a Cleric is %d/%d, want 0/1", p, tt)
	}
	oldB207NextChapter(t, g)
	demons := oldB207OnBattlefield(g, me.ID, "Demon")
	if len(demons) != 1 {
		t.Fatalf("%d Demons, want 1", len(demons))
	}
	if p, tt := effectstest.PtOf(t, g, demons[0].InstanceID); p != 6 || tt != 6 {
		t.Errorf("the Demon is %d/%d, want 6/6", p, tt)
	}
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, demons[0].InstanceID, kw) {
			t.Errorf("the Demon lacks %s", kw)
		}
	}
	// Upkeep: sacrifice another creature of your choice.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.SlChoose(t, g, me.ID, clerics[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(clerics[0].InstanceID) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(demons[0].InstanceID) {
		t.Error("the Demon itself was sacrificed")
	}
}

func TestOldB207RiteOfBelzenlokDemonHurtsYouWhenItHasNothingToEat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	demon := effectstest.PushToken(g, me.ID, effects.OldB207BelzenlokDemonToken())
	life := me.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-6 {
		t.Errorf("life %d -> %d, want 6 damage from the Demon", life, me.Life)
	}
	if !g.Battlefield.Contains(demon) {
		t.Error("the Demon left the battlefield")
	}
}

// --- The Antiquities War ---

func TestOldB207AntiquitiesWarTakesAnArtifactThenMakesArtifactsFiveFives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	var filler []uuid.UUID
	for range 4 {
		filler = append(filler, effectstest.PushLibraryCard(me, game.Card{Name: "Filler", TypeLine: "Sorcery"}))
	}
	art := effectstest.PushLibraryCard(me, game.Card{Name: "Fine Bauble", TypeLine: "Artifact"})
	rock := effectstest.B12Permanent(g, me.ID, "Mind Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "The Antiquities War", oldB207SagaType, oldB207AntiquitiesOrc, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, art)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(art) {
		t.Error("the revealed artifact card was not put into hand")
	}
	for _, id := range filler {
		if !me.Library.Contains(id) {
			t.Error("a card that was not taken left the library")
		}
	}
	oldB207NextChapter(t, g)
	oldB207NextChapter(t, g)
	if p, tt := effectstest.PtOf(t, g, rock); p != 5 || tt != 5 {
		t.Errorf("the artifact is %d/%d, want 5/5", p, tt)
	}
	if c, ok := effectstest.BattlefieldCard(g, rock); !ok || !c.IsCreature() {
		t.Error("the artifact is not a creature")
	}
}

// --- The First Eruption ---

func TestOldB207FirstEruptionDamagesGroundCreaturesAddsManaAndSweeps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ground := effectstest.B12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 1, 1)
	flier := effectstest.B12Push(g, opp.ID, "Flier", "Creature — Bird", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 3, 3)
	mountain := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.CastCatalogSpell(t, g, "The First Eruption", oldB207SagaType, oldB207EruptionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ground) {
		t.Error("the creature without flying survived 1 damage")
	}
	if !g.Battlefield.Contains(flier) {
		t.Error("the flier was dealt damage")
	}
	oldB207NextChapter(t, g)
	if me.ManaPool.Total() != 2 {
		t.Errorf("mana pool is %d after chapter II, want {R}{R}", me.ManaPool.Total())
	}
	oldB207NextChapter(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("chapter III did not ask which Mountain to sacrifice")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{mountain}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mountain) {
		t.Error("the Mountain was not sacrificed")
	}
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(flier) {
		t.Error("3 damage to each creature left a creature alive")
	}
}

func TestOldB207FirstEruptionChapterThreeDoesNothingWithoutAMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 3, 3)
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "The First Eruption", oldB207SagaType, oldB207EruptionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB207NextChapter(t, g)
	oldB207NextChapter(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("the 3 damage was dealt without a Mountain sacrificed")
	}
}
