package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_gl_cards_test.go — Bloomburrow cards starting with G and I to L
// (#4467). Each test plays the real catalog entry and asserts what its
// text says.

const (
	blbGLMooseOracle        = "b0747a4f-d041-4750-aef5-a5d101696f65"
	blbGLGevOracle          = "77b832a8-a3af-44fe-bc67-72510d292667"
	blbGLGlidediveOracle    = "33c06c11-8d0a-45ce-adea-c9a7a8ac8779"
	blbGLRabbitOracle       = "7ed4afc9-959a-4868-8f4a-2221bea1ba32"
	blbGLVinelasherOracle   = "4f8c4927-46c2-4c18-b5af-c6288dbaff6d"
	blbGLBruiserOracle      = "0a2fbdd3-4940-4bbf-ae19-f6d00d7ff294"
	blbGLKindlesparkOracle  = "980cd7d5-a511-49ab-bc52-08b71c65db76"
	blbGLKitnapOracle       = "a1bd1e83-40ac-449f-838c-ef308f275a5d"
	blbGLKitsaOracle        = "16414460-11ce-481d-ab54-6f0e0ea7d83d"
	blbGLKnightfisherOracle = "cb2d222c-df42-41a7-a517-0a41e47a4280"
	blbGLLightshellOracle   = "ef8fab12-d7d2-4c21-8b8c-41627048ac8b"
	blbGLMentorOracle       = "20194297-f2a4-4473-95bc-7c6835e5b6b3"
	blbGLBrawlOracle        = "adc8976a-0871-4fd6-9c4c-822ba9549b4b"
	blbGLVillageOracle      = "b6c7c708-5212-4100-b954-b77855b27915"
	blbGLRabbitTypeLine     = "Creature — Rabbit Soldier"
	blbGLVinelasherType     = "Creature — Lizard Assassin"
)

func TestBlbGLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		blbGLMooseOracle:        "Galewind Moose",
		blbGLGevOracle:          "Gev, Scaled Scorch",
		blbGLGlidediveOracle:    "Glidedive Duo",
		blbGLRabbitOracle:       "Intrepid Rabbit",
		blbGLVinelasherOracle:   "Iridescent Vinelasher",
		blbGLBruiserOracle:      "Junkblade Bruiser",
		blbGLKindlesparkOracle:  "Kindlespark Duo",
		blbGLKitnapOracle:       "Kitnap",
		blbGLKitsaOracle:        "Kitsa, Otterball Elite",
		blbGLKnightfisherOracle: "Knightfisher",
		blbGLLightshellOracle:   "Lightshell Duo",
		blbGLMentorOracle:       "Lilysplash Mentor",
		blbGLBrawlOracle:        "Longstalk Brawl",
		blbGLVillageOracle:      "Lupinflower Village",
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

// blbGLCastCreature casts a creature with real P/T from the active
// player's hand, paying the optional costs at `optional`, and resolves
// it, leaving its entry triggers on the stack.
func blbGLCastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, optional []int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{OptionalCosts: optional}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// blbGLSettleTriggerOrder answers an open trigger-order prompt for
// chooser in the order offered (CR 603.3b), if there is one. It passes
// no priority, so nothing else moves.
func blbGLSettleTriggerOrder(g *game.Game, chooser uuid.UUID) {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder && c.Chooser == chooser {
			_ = g.ResolveTriggerOrder(c.ID, chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...))
			return
		}
	}
}

// --- Galewind Moose -----------------------------------------------------------

func TestBlbGLGalewindMooseHasItsFourKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	moose := blbB1Push(g, me.ID, "Galewind Moose", "Creature — Elemental Elk", blbGLMooseOracle, 6, 6)
	for _, kw := range []string{"flash", "reach", "vigilance", "trample"} {
		if !hasKeywordOnBattlefield(t, g, moose, kw) {
			t.Errorf("the Moose lacks %s", kw)
		}
	}
}

func TestBlbGLGalewindMooseCastsDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	advanceTo(t, g, game.StepBeginCombat)
	id := pushCatalogHandCard(me, "Galewind Moose", "Creature — Elemental Elk", blbGLMooseOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Errorf("flash: casting during combat was refused: %v", err)
	}
}

// --- Gev, Scaled Scorch -------------------------------------------------------

func TestBlbGLGevGivesOthersACounterPerOpponentWhoLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	gev := blbB1Push(g, me.ID, "Gev, Scaled Scorch", "Legendary Creature — Lizard Mercenary", blbGLGevOracle, 3, 2)
	first := castHoldCreature(t, g, "First Bear", "Creature — Bear", "", 2, 2)
	if got := plusOnes(g, first); got != 0 {
		t.Fatalf("no opponent lost life: the Bear has %d counters, want 0", got)
	}
	blbDFOpponentLosesLife(t, g, opp.ID)
	blbDFOpponentLosesLife(t, g, opp.ID)
	blbDFOpponentLosesLife(t, g, other.ID)
	blbDFOpponentLosesLife(t, g, me.ID)
	second := castHoldCreature(t, g, "Second Bear", "Creature — Bear", "", 2, 2)
	if got := plusOnes(g, second); got != 2 {
		t.Errorf("two opponents lost life: the Bear has %d counters, want 2", got)
	}
	if got := plusOnes(g, gev); got != 0 {
		t.Errorf("Gev has %d counters; it gives only other creatures counters", got)
	}
}

func TestBlbGLGevPingsAnOpponentForALizardSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Gev, Scaled Scorch", "Legendary Creature — Lizard Mercenary", blbGLGevOracle, 3, 2)
	if _, err := blbHCastWithMana(t, g, "Bear", "Creature — Bear", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Lizard spell triggered Gev")
	}
	passPriorityAroundTable(t, g)
	if _, err := blbHCastWithMana(t, g, "Lizard", "Creature — Lizard Warrior", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Lizard: %v", err)
	}
	if err := blbHAnswerPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("Gev's controller was accepted as target opponent")
	}
	life := opp.Life
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestBlbGLGevHasWardPayTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gev := pushCatalogPermanent(g, me.ID, "Gev, Scaled Scorch", "Legendary Creature — Lizard Mercenary", blbGLGevOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, gev)
	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the opponent who targeted Gev")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("the ward charges %d life, want 2", prompt.LifeCost)
	}
}

// --- Glidedive Duo ------------------------------------------------------------

func TestBlbGLGlidediveDuoDrainsEachOpponentForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	myLife, oppLife, otherLife := lifeOf(g, me.ID), lifeOf(g, opp.ID), lifeOf(g, other.ID)
	duo := castHoldCreature(t, g, "Glidedive Duo", "Creature — Bat Lizard", blbGLGlidediveOracle, 3, 3)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-2 {
		t.Errorf("an opponent is at %d, want %d", got, oppLife-2)
	}
	if got := lifeOf(g, other.ID); got != otherLife-2 {
		t.Errorf("another opponent is at %d, want %d", got, otherLife-2)
	}
	if got := lifeOf(g, me.ID); got != myLife+2 {
		t.Errorf("I am at %d, want %d", got, myLife+2)
	}
	if !hasKeywordOnBattlefield(t, g, duo, "flying") {
		t.Error("the Duo lacks flying")
	}
}

// --- Intrepid Rabbit ----------------------------------------------------------

func TestBlbGLIntrepidRabbitPumpsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castHoldCreature(t, g, "Intrepid Rabbit", blbGLRabbitTypeLine, blbGLRabbitOracle, 3, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the pump")
	}
	if hasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered")
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if pw, tough := ptOf(t, g, bear); pw != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", pw, tough)
	}
}

func TestBlbGLIntrepidRabbitOffspringMakesAOneOneCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	rabbit := blbGLCastCreature(t, g, "Intrepid Rabbit", blbGLRabbitTypeLine, blbGLRabbitOracle, 3, 2, []int{0})
	for i := 0; i < 6; i++ {
		blbGLSettleTriggerOrder(g, me.ID)
		if p := latestPickTarget(g, me.ID); p != nil {
			answerPickTarget(t, g, rabbit)
		}
		passPriorityAroundTable(t, g)
	}
	rabbits := blbMP2Named(g, "Intrepid Rabbit")
	if len(rabbits) != 2 {
		t.Fatalf("%d Intrepid Rabbits, want 2 (offspring paid)", len(rabbits))
	}
	for _, c := range rabbits {
		if c.IsToken() && (c.Power != 1 || c.Toughness != 1) {
			t.Errorf("the offspring token is %d/%d, want 1/1", c.Power, c.Toughness)
		}
	}
}

// --- Iridescent Vinelasher ----------------------------------------------------

func TestBlbGLIridescentVinelasherPingsAnOpponentOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Iridescent Vinelasher", blbGLVinelasherType, blbGLVinelasherOracle, 1, 2)
	playLandFromHand(t, g, "Forest", "")
	if err := blbHAnswerPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("the Vinelasher's controller was accepted as target opponent")
	}
	life := opp.Life
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestBlbGLIridescentVinelasherHasOffspringTwo(t *testing.T) {
	spec, _ := Lookup(blbGLVinelasherOracle)
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}
}

// --- Junkblade Bruiser --------------------------------------------------------

func TestBlbGLJunkbladeBruiserGrowsOnExpendFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bruiser := blbB1Push(g, me.ID, "Junkblade Bruiser", "Creature — Raccoon Berserker", blbGLBruiserOracle, 4, 5)
	toMainForCost(t, g)
	castPaid(t, g, me, "Three", "{2}{G}", "CCG")
	passPriorityAroundTable(t, g)
	if pw, tough := ptOf(t, g, bruiser); pw != 4 || tough != 5 {
		t.Fatalf("three mana spent: the Bruiser is %d/%d, want 4/5", pw, tough)
	}
	castPaid(t, g, me, "One", "{G}", "G")
	passPriorityAroundTable(t, g)
	if pw, tough := ptOf(t, g, bruiser); pw != 6 || tough != 6 {
		t.Errorf("the fourth mana was spent: the Bruiser is %d/%d, want 6/6", pw, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bruiser, "trample") {
		t.Error("the Bruiser lacks trample")
	}
}

// --- Kindlespark Duo ----------------------------------------------------------

func TestBlbGLKindlesparkDuoPingsAndUntapsOnANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	duo := blbB1Push(g, me.ID, "Kindlespark Duo", "Creature — Lizard Otter", blbGLKindlesparkOracle, 1, 3)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, duo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("the Duo's controller was accepted as target opponent")
	}
	life := opp.Life
	b16Activate(t, g, me.ID, duo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	if c, _ := battlefieldCard(g, duo); !c.Tapped {
		t.Fatal("the Duo should be tapped")
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, duo); c.Tapped {
		t.Error("a noncreature spell did not untap the Duo")
	}
}

// --- Kitnap -------------------------------------------------------------------

func TestBlbGLKitnapStealsTapsAndStunsWithoutTheGift(t *testing.T) {
	for _, promised := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		bear := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
		oppHand := opp.Hand.Size()
		to := uuid.Nil
		if promised {
			to = opp.ID
		}
		mustCastWithGift(t, g, "Kitnap", "Enchantment — Aura", blbGLKitnapOracle, cardTarget(bear), to)
		for i := 0; i < 3; i++ {
			blbGLSettleTriggerOrder(g, me.ID)
			passPriorityAroundTable(t, g)
		}
		c := layeredCard(t, g, bear)
		if c.Controller != me.ID {
			t.Errorf("promised=%v: the Bear is controlled by %s, want me", promised, c.Controller)
		}
		if !c.Tapped {
			t.Errorf("promised=%v: the Bear should be tapped", promised)
		}
		wantStun, wantHand := 3, oppHand
		if promised {
			wantStun, wantHand = 0, oppHand+1
		}
		if got := countersOn(g, bear, game.CounterStun); got != wantStun {
			t.Errorf("promised=%v: %d stun counters, want %d", promised, got, wantStun)
		}
		if got := opp.Hand.Size(); got != wantHand {
			t.Errorf("promised=%v: the opponent has %d cards, want %d", promised, got, wantHand)
		}
	}
}

// --- Kitsa, Otterball Elite ---------------------------------------------------

func TestBlbGLKitsaCopiesASpellOnlyAtPowerThree(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	kitsa := blbB1Push(g, me.ID, "Kitsa, Otterball Elite", "Legendary Creature — Otter Wizard", blbGLKitsaOracle, 1, 3)
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand := me.Hand.Size()
	spell := castCatalogSpell(t, g, "Divination", "Sorcery", divinationOracle, nil)
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, kitsa, 1, params); err == nil {
		t.Fatal("the copy activated at power 1")
	}
	pfAddCounters(g, kitsa, game.CounterPlusOne, 2)
	b16Activate(t, g, me.ID, kitsa, 1, params)
	for i := 0; i < 4; i++ {
		passPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d, want %d: Divination and its copy draw two each", got, hand+4)
	}
	for _, kw := range []string{"vigilance", "prowess"} {
		if !hasKeywordOnBattlefield(t, g, kitsa, kw) {
			t.Errorf("Kitsa lacks %s", kw)
		}
	}
}

func TestBlbGLKitsaLootsWithItsTapAbility(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	kitsa := blbB1Push(g, me.ID, "Kitsa, Otterball Elite", "Legendary Creature — Otter Wizard", blbGLKitsaOracle, 1, 3)
	top := seedLibrary(me, "Top")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, kitsa, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[0]) {
		t.Fatal("the loot drew nothing")
	}
	if discardChoiceFor(g, me.ID) == nil {
		t.Error("the loot asked for no discard")
	}
}

// --- Knightfisher -------------------------------------------------------------

func TestBlbGLKnightfisherMakesAFishForAnotherNontokenBird(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Knightfisher", "Creature — Bird Knight", blbGLKnightfisherOracle, 4, 5)
	castHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Fish"); n != 0 {
		t.Fatalf("a Bear made %d Fish, want 0", n)
	}
	castHoldCreature(t, g, "Bird", "Creature — Bird Scout", "", 1, 1)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Fish"); n != 1 {
		t.Errorf("a nontoken Bird made %d Fish, want 1", n)
	}
}

// --- Lightshell Duo -----------------------------------------------------------

func TestBlbGLLightshellDuoSurveilsTwoAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	duo := castHoldCreature(t, g, "Lightshell Duo", "Creature — Rat Otter", blbGLLightshellOracle, 3, 4)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looks at %d cards, want 2", len(c.ScryCards))
	}
	if !hasKeywordOnBattlefield(t, g, duo, "prowess") {
		t.Error("the Duo lacks prowess")
	}
}

// --- Lilysplash Mentor --------------------------------------------------------

func TestBlbGLLilysplashMentorFlickersAnotherCreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	mentor := blbB1Push(g, me.ID, "Lilysplash Mentor", "Creature — Frog Druid", blbGLMentorOracle, 4, 4)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "G", "U")
	if err := g.ActivateCatalogAbility(me.ID, mentor, 0, game.ActivateAbilityParams{Targets: cardRefs(mentor)}); err == nil {
		t.Fatal("the Mentor was accepted as another target creature")
	}
	b16Activate(t, g, me.ID, mentor, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the old Bear is still on the battlefield")
	}
	back := findBattlefieldByName(g, "Bear")
	if back == uuid.Nil {
		t.Fatal("the Bear did not return")
	}
	if got := plusOnes(g, back); got != 1 {
		t.Errorf("the Bear returned with %d +1/+1 counters, want 1", got)
	}
}

// --- Longstalk Brawl ----------------------------------------------------------

func TestBlbGLLongstalkBrawlCountersOnlyWithTheGift(t *testing.T) {
	for _, promised := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		mine := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 3, 3)
		theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
		to := uuid.Nil
		if promised {
			to = opp.ID
		}
		mustCastWithGift(t, g, "Longstalk Brawl", "Sorcery", blbGLBrawlOracle, []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
		}, to)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(theirs) {
			t.Errorf("promised=%v: their 3/3 survived the fight", promised)
		}
		if got := g.Battlefield.Contains(mine); got != promised {
			t.Errorf("promised=%v: my creature survived=%v", promised, got)
		}
		wantFish := 0
		if promised {
			wantFish = 1
		}
		if n, tapped := countTokensNamed(g, opp.ID, "Fish"); n != wantFish || tapped != wantFish {
			t.Errorf("promised=%v: %d Fish (%d tapped), want %d tapped", promised, n, tapped, wantFish)
		}
	}
}

// --- Lupinflower Village ------------------------------------------------------

func TestBlbGLLupinflowerVillageDigsForABird(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	village := blbB1Push(g, me.ID, "Lupinflower Village", "Land", blbGLVillageOracle, 0, 0)
	me.Library.Cards = nil
	var bird, bear uuid.UUID
	deep := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: deep, Name: "Deep Bird", TypeLine: "Creature — Bird", Owner: me.ID, Controller: me.ID})
	for i, typeLine := range []string{"Creature — Bear", "Creature — Bird Scout", "Land", "Sorcery", "Artifact", "Creature — Rat"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Card", TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		switch i {
		case 0:
			bear = id
		case 1:
			bird = id
		}
	}
	advanceToMain(t, g)
	b06AddMana(me, "C", "W")
	b16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(village) {
		t.Fatal("the Village was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look")
	}
	if len(c.ChooseCards) != 1 {
		t.Errorf("%d cards offered, want the Bird only (the Deep Bird is seventh)", len(c.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, bird)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bird) {
		t.Error("the Bird is not in hand")
	}
	if me.Hand.Contains(bear) || !me.Library.Contains(bear) {
		t.Error("the Bear should stay in the library")
	}
	if n := len(me.Library.Cards); n != 6 || me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the untouched seventh card should be on top over the five bottomed ones: %d cards", n)
	}
}
