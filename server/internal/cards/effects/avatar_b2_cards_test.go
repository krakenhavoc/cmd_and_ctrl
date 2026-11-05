package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b2_cards_test.go — Avatar: The Last Airbender, batch 2 (#4459).

const (
	avB2FancyFootworkOracle  = "c5648a32-98b0-4625-aae1-ae56e0ef0197"
	avB2GatherOracle         = "db69c3c9-d809-4361-8055-8c2dd4c6dd67"
	avB2GliderKidsOracle     = "ff5ce28b-25be-4136-a7b2-8f25c08a6455"
	avB2GliderStaffOracle    = "823006a8-098c-457a-8546-77ae2457c459"
	avB2InvasionOracle       = "185316ff-6fb7-4b5c-8a60-3238674562a6"
	avB2JeongJeongOracle     = "6a06d281-3ca2-498f-91d0-e13e7b0869f9"
	avB2KyoshiWarriorsOracle = "b29abb09-75ea-431c-ab60-f2940d002e83"
	avB2MasterPiandaoOracle  = "e294d9d2-df91-4410-85d3-8ffb13fb2285"
	avB2MomoOracle           = "85ee8ab6-8db5-45d7-85ae-64838771146a"
	avB2RabarooTroopOracle   = "48252207-d345-455c-b655-c2027664e6a6"
)

func TestAvatarB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB2FancyFootworkOracle:  "Fancy Footwork",
		avB2GatherOracle:         "Gather the White Lotus",
		avB2GliderKidsOracle:     "Glider Kids",
		avB2GliderStaffOracle:    "Glider Staff",
		avB2InvasionOracle:       "Invasion Reinforcements",
		avB2JeongJeongOracle:     "Jeong Jeong's Deserters",
		avB2KyoshiWarriorsOracle: "Kyoshi Warriors",
		avB2MasterPiandaoOracle:  "Master Piandao",
		avB2MomoOracle:           "Momo, Playful Pet",
		avB2RabarooTroopOracle:   "Rabaroo Troop",
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

func avB2Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: owner, Controller: owner,
	})
}

func avB2Tapped(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, Tapped: true,
	})
}

// --- Fancy Footwork -----------------------------------------------------------

func TestAvatarB2FancyFootworkUntapsAndPumpsBothTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := avB2Tapped(g, me.ID, "My Bear")
	theirs := avB2Tapped(g, opp.ID, "Their Bear")
	other := avB2Tapped(g, me.ID, "Untargeted Bear")
	castCatalogSpell(t, g, "Fancy Footwork", "Instant — Lesson", avB2FancyFootworkOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if c, _ := battlefieldCard(g, id); c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
		if p, tough := p2PT(g, id); p != 4 || tough != 4 {
			t.Errorf("target is %d/%d, want 4/4", p, tough)
		}
	}
	if c, _ := battlefieldCard(g, other); !c.Tapped {
		t.Error("an untargeted creature was untapped")
	}
	if p, tough := p2PT(g, other); p != 2 || tough != 2 {
		t.Errorf("untargeted creature is %d/%d, want 2/2", p, tough)
	}
}

func TestAvatarB2FancyFootworkRefusesAThirdTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a, b, c := avB2Tapped(g, me.ID, "A"), avB2Tapped(g, me.ID, "B"), avB2Tapped(g, me.ID, "C")
	if err := castCatalogSpellErr(t, g, "Fancy Footwork", "Instant — Lesson", avB2FancyFootworkOracle, cardRefs(a, b, c)); err == nil {
		t.Fatal("Fancy Footwork was cast with three targets")
	}
}

func TestAvatarB2FancyFootworkStrictCastNeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := avB2Tapped(g, me.ID, "Bear")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Fancy Footwork", TypeLine: "Instant — Lesson", ManaCost: "{2}{W}",
		OracleID: avB2FancyFootworkOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a strict cast with an empty pool was allowed")
	}
	fillPoolColored(me, "W", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("a strict cast with {W}{W}{W} floating was refused: %v", err)
	}
}

// --- Gather the White Lotus ---------------------------------------------------

func TestAvatarB2GatherTheWhiteLotusMakesAnAllyPerPlainsThenScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avB2Land(g, me.ID, "Plains", "Basic Land — Plains")
	avB2Land(g, me.ID, "Plains", "Basic Land — Plains")
	avB2Land(g, me.ID, "Prairie Stream", "Land — Plains Island")
	avB2Land(g, me.ID, "Island", "Basic Land — Island")
	avB2Land(g, opp.ID, "Plains", "Basic Land — Plains")
	seedLibrary(me, "One", "Two", "Three")
	castCatalogSpell(t, g, "Gather the White Lotus", "Sorcery", avB2GatherOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(c.ScryCards))
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Ally", me.ID); got != 3 {
		t.Errorf("Ally tokens = %d, want 3 (one per Plains you control)", got)
	}
	if got := tokensNamedControlledBy(g, "Ally", opp.ID); got != 0 {
		t.Errorf("the opponent got %d Ally tokens", got)
	}
}

// --- Glider Kids --------------------------------------------------------------

func TestAvatarB2GliderKidsFliesAndScriesOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two")
	kids := castCatalogSpell(t, g, "Glider Kids", "Creature — Human Pilot Ally", avB2GliderKidsOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt on entry")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry looked at %d cards, want 1", len(c.ScryCards))
	}
	if !hasEffectiveKeyword(t, g, kids, "flying") {
		t.Error("Glider Kids lacks flying")
	}
}

// --- Glider Staff -------------------------------------------------------------

func TestAvatarB2GliderStaffAirbendsThenEquipsForPlusOneAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	staff := castCatalogSpell(t, g, "Glider Staff", "Artifact — Equipment", avB2GliderStaffOracle, nil)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(victim)...)
	passPriorityAroundTable(t, g)
	assertAirbent(t, g, victim, opp.ID)

	if err := hobB4Activate(t, g, me.ID, staff, 0, "{C}{C}", cardRefs(bear)); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := p2PT(g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("equipped creature lacks flying")
	}
}

func TestAvatarB2GliderStaffMayAirbendNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")
	castCatalogSpell(t, g, "Glider Staff", "Artifact — Equipment", avB2GliderStaffOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTargetRefs(t, g, me.ID, nil)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("a creature was airbent although no target was chosen")
	}
}

func TestAvatarB2GliderStaffEquipNeedsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	staff := p2Push(g, me.ID, "Glider Staff", "Artifact — Equipment", avB2GliderStaffOracle, 0, 0)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	if err := hobB4Activate(t, g, me.ID, staff, 0, "{C}", cardRefs(bear)); err == nil {
		t.Fatal("equip {2} was paid with one mana")
	}
}

// --- Invasion Reinforcements / Kyoshi Warriors --------------------------------

func TestAvatarB2InvasionReinforcementsHasFlashAndMakesAnAlly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepBeginCombat)
	castFromHandForTest(t, g, me, "Invasion Reinforcements", "Creature — Human Warrior Ally", "", avB2InvasionOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Ally", me.ID); got != 1 {
		t.Errorf("Ally tokens = %d, want 1", got)
	}
}

func TestAvatarB2KyoshiWarriorsMakesAnAllyAndHasNoFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Kyoshi Warriors", "Creature — Human Warrior Ally", avB2KyoshiWarriorsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Ally", me.ID); got != 1 {
		t.Errorf("Ally tokens = %d, want 1", got)
	}

	advanceTo(t, g, game.StepBeginCombat)
	id := pushCatalogHandCard(me, "Kyoshi Warriors", "Creature — Human Warrior Ally", avB2KyoshiWarriorsOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("Kyoshi Warriors was cast at instant speed")
	}
}

// --- Jeong Jeong's Deserters --------------------------------------------------

func TestAvatarB2JeongJeongsDesertersPutsACounterOnTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Jeong Jeong's Deserters", "Creature — Human Rebel Ally", avB2JeongJeongOracle, nil)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(theirs)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestAvatarB2JeongJeongsDesertersCannotTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := avB2Land(g, me.ID, "Plains", "Basic Land — Plains")
	castCatalogSpell(t, g, "Jeong Jeong's Deserters", "Creature — Human Rebel Ally", avB2JeongJeongOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(land)); err == nil {
		t.Error("a land was accepted as the target creature")
	}
}

// --- Master Piandao -----------------------------------------------------------

func avB2PiandaoAttacks(t *testing.T, g *game.Game) (me *game.Player, ids []uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	ids = samiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Staff", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Lesson", TypeLine: "Instant — Lesson"},
		game.Card{Name: "Friend", TypeLine: "Creature — Human Ally"},
		game.Card{Name: "Fifth", TypeLine: "Artifact — Equipment"},
	)
	piandao := p2Push(g, me.ID, "Master Piandao", "Legendary Creature — Human Warrior Ally", avB2MasterPiandaoOracle, 4, 4, "W")
	assertKeywords(t, g, piandao, "first strike")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(piandao, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	return me, ids
}

func TestAvatarB2MasterPiandaoTakesAnEquipmentFromTheTopFour(t *testing.T) {
	g := newCatalogGame(t)
	me, ids := avB2PiandaoAttacks(t, g)
	lib := me.Library.Size()
	answerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Fatal("the Equipment did not reach the hand")
	}
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card taken", lib, me.Library.Size())
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[4] {
		t.Error("the fifth card must now be on top; the other three went to the bottom")
	}
}

func TestAvatarB2MasterPiandaoOffersAlliesAndLessonsButNotOtherCreatures(t *testing.T) {
	for _, pick := range []int{2, 3} {
		g := newCatalogGame(t)
		me, ids := avB2PiandaoAttacks(t, g)
		answerChooseCards(t, g, me.ID, ids[pick])
		if !me.Hand.Contains(ids[pick]) {
			t.Errorf("card %d did not reach the hand", pick)
		}
	}
	g := newCatalogGame(t)
	me, ids := avB2PiandaoAttacks(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]}); err == nil && me.Hand.Contains(ids[0]) {
		t.Error("a non-Ally creature was taken")
	}
}

// --- Momo, Playful Pet --------------------------------------------------------

func avB2MomoLeaves(t *testing.T, g *game.Game, mode int) (me *game.Player, bear uuid.UUID) {
	t.Helper()
	me = g.Seats[0]
	advanceToMain(t, g)
	momo := p2Push(g, me.ID, "Momo, Playful Pet", "Legendary Creature — Lemur Bat Ally", avB2MomoOracle, 1, 1, "W")
	assertKeywords(t, g, momo, "flying", "vigilance")
	bear = pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	seedLibrary(me, "One", "Two", "Three")
	b25Destroy(g, momo)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Momo leaving asks which mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("mode: %v", err)
	}
	if mode == 1 {
		pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	}
	passPriorityAroundTable(t, g)
	return me, bear
}

func TestAvatarB2MomoMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := avB2MomoLeaves(t, g, 0)
	if got := tokensNamedControlledBy(g, "Food", me.ID); got != 1 {
		t.Errorf("Food tokens = %d, want 1", got)
	}
}

func TestAvatarB2MomoPutsACounterOnYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, bear := avB2MomoLeaves(t, g, 1)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

func TestAvatarB2MomoScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := avB2MomoLeaves(t, g, 2)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(c.ScryCards))
	}
}

// --- Rabaroo Troop ------------------------------------------------------------

func TestAvatarB2RabarooTroopLandfallGrantsFlyingAndOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	troop := p2Push(g, me.ID, "Rabaroo Troop", "Creature — Rabbit Kangaroo", avB2RabarooTroopOracle, 3, 5, "W")
	if hasEffectiveKeyword(t, g, troop, "flying") {
		t.Fatal("Rabaroo Troop flies before any land entered")
	}
	life := me.Life
	b31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, troop, "flying") {
		t.Error("landfall did not grant flying")
	}
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestAvatarB2RabarooTroopIgnoresAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	troop := p2Push(g, me.ID, "Rabaroo Troop", "Creature — Rabbit Kangaroo", avB2RabarooTroopOracle, 3, 5, "W")
	life := me.Life
	aangAdvanceToMain(t, g, 1)
	b31PlayLand(t, g, "Plains", "Basic Land — Plains", "")
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, troop, "flying") || me.Life != life {
		t.Error("an opponent's land triggered Rabaroo Troop's landfall")
	}
	_ = opp
}

func TestAvatarB2RabarooTroopPlainscyclesForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	id, _ := cycleFromHand(t, g, "Rabaroo Troop", "Creature — Rabbit Kangaroo", avB2RabarooTroopOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the plainscycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 {
		t.Fatalf("search offers %d cards, want only the Plains", len(c.SearchCards))
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Plains did not reach the hand")
	}
}

func TestAvatarB2RabarooTroopPlainscyclingNeedsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Rabaroo Troop", "Creature — Rabbit Kangaroo", avB2RabarooTroopOracle)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}", nil); err == nil {
		t.Fatal("plainscycling {2} was paid with one mana")
	}
}
