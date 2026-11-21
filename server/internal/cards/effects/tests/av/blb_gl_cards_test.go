package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBlbGLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.BlbGLMooseOracle:        "Galewind Moose",
		effectstest.BlbGLGevOracle:          "Gev, Scaled Scorch",
		effectstest.BlbGLGlidediveOracle:    "Glidedive Duo",
		effectstest.BlbGLRabbitOracle:       "Intrepid Rabbit",
		effectstest.BlbGLVinelasherOracle:   "Iridescent Vinelasher",
		effectstest.BlbGLBruiserOracle:      "Junkblade Bruiser",
		effectstest.BlbGLKindlesparkOracle:  "Kindlespark Duo",
		effectstest.BlbGLKitnapOracle:       "Kitnap",
		effectstest.BlbGLKitsaOracle:        "Kitsa, Otterball Elite",
		effectstest.BlbGLKnightfisherOracle: "Knightfisher",
		effectstest.BlbGLLightshellOracle:   "Lightshell Duo",
		effectstest.BlbGLMentorOracle:       "Lilysplash Mentor",
		effectstest.BlbGLBrawlOracle:        "Longstalk Brawl",
		effectstest.BlbGLVillageOracle:      "Lupinflower Village",
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

// --- Galewind Moose -----------------------------------------------------------

func TestBlbGLGalewindMooseHasItsFourKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	moose := effectstest.BlbB1Push(g, me.ID, "Galewind Moose", "Creature — Elemental Elk", effectstest.BlbGLMooseOracle, 6, 6)
	for _, kw := range []string{"flash", "reach", "vigilance", "trample"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, moose, kw) {
			t.Errorf("the Moose lacks %s", kw)
		}
	}
}

func TestBlbGLGalewindMooseCastsDuringCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := effectstest.PushCatalogHandCard(me, "Galewind Moose", "Creature — Elemental Elk", effectstest.BlbGLMooseOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Errorf("flash: casting during combat was refused: %v", err)
	}
}

// --- Gev, Scaled Scorch -------------------------------------------------------

func TestBlbGLGevGivesOthersACounterPerOpponentWhoLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	gev := effectstest.BlbB1Push(g, me.ID, "Gev, Scaled Scorch", "Legendary Creature — Lizard Mercenary", effectstest.BlbGLGevOracle, 3, 2)
	first := effectstest.CastHoldCreature(t, g, "First Bear", "Creature — Bear", "", 2, 2)
	if got := effectstest.PlusOnes(g, first); got != 0 {
		t.Fatalf("no opponent lost life: the Bear has %d counters, want 0", got)
	}
	blbDFOpponentLosesLife(t, g, opp.ID)
	blbDFOpponentLosesLife(t, g, opp.ID)
	blbDFOpponentLosesLife(t, g, other.ID)
	blbDFOpponentLosesLife(t, g, me.ID)
	second := effectstest.CastHoldCreature(t, g, "Second Bear", "Creature — Bear", "", 2, 2)
	if got := effectstest.PlusOnes(g, second); got != 2 {
		t.Errorf("two opponents lost life: the Bear has %d counters, want 2", got)
	}
	if got := effectstest.PlusOnes(g, gev); got != 0 {
		t.Errorf("Gev has %d counters; it gives only other creatures counters", got)
	}
}

func TestBlbGLGevPingsAnOpponentForALizardSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Gev, Scaled Scorch", "Legendary Creature — Lizard Mercenary", effectstest.BlbGLGevOracle, 3, 2)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Bear", "Creature — Bear", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Bear: %v", err)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Lizard spell triggered Gev")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := effectstest.BlbHCastWithMana(t, g, "Lizard", "Creature — Lizard Warrior", "", "", 2, 2); err != nil {
		t.Fatalf("cast the Lizard: %v", err)
	}
	if err := blbHAnswerPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("Gev's controller was accepted as target opponent")
	}
	life := opp.Life
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestBlbGLGevHasWardPayTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gev := effectstest.PushCatalogPermanent(g, me.ID, "Gev, Scaled Scorch", "Legendary Creature — Lizard Mercenary", effectstest.BlbGLGevOracle, false)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, gev)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the opponent who targeted Gev")
	}
	if prompt.LifeCost != 2 {
		t.Errorf("the ward charges %d life, want 2", prompt.LifeCost)
	}
}

// --- Glidedive Duo ------------------------------------------------------------

func TestBlbGLGlidediveDuoDrainsEachOpponentForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	myLife, oppLife, otherLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, other.ID)
	duo := effectstest.CastHoldCreature(t, g, "Glidedive Duo", "Creature — Bat Lizard", effectstest.BlbGLGlidediveOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-2 {
		t.Errorf("an opponent is at %d, want %d", got, oppLife-2)
	}
	if got := effectstest.LifeOf(g, other.ID); got != otherLife-2 {
		t.Errorf("another opponent is at %d, want %d", got, otherLife-2)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife+2 {
		t.Errorf("I am at %d, want %d", got, myLife+2)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, duo, "flying") {
		t.Error("the Duo lacks flying")
	}
}

// --- Intrepid Rabbit ----------------------------------------------------------

func TestBlbGLIntrepidRabbitPumpsACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Intrepid Rabbit", effectstest.BlbGLRabbitTypeLine, effectstest.BlbGLRabbitOracle, 3, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the pump")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("an opponent's creature was offered")
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.PtOf(t, g, bear); pw != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", pw, tough)
	}
}

func TestBlbGLIntrepidRabbitOffspringMakesAOneOneCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	rabbit := effectstest.BlbGLCastCreature(t, g, "Intrepid Rabbit", effectstest.BlbGLRabbitTypeLine, effectstest.BlbGLRabbitOracle, 3, 2, []int{0})
	for i := 0; i < 6; i++ {
		effectstest.BlbGLSettleTriggerOrder(g, me.ID)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.AnswerPickTarget(t, g, rabbit)
		}
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Iridescent Vinelasher", effectstest.BlbGLVinelasherType, effectstest.BlbGLVinelasherOracle, 1, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	if err := blbHAnswerPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("the Vinelasher's controller was accepted as target opponent")
	}
	life := opp.Life
	if err := blbHAnswerPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestBlbGLIridescentVinelasherHasOffspringTwo(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.BlbGLVinelasherOracle)
	if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != effects.OffspringKey || spec.OptionalCosts[0].ManaCost != "{2}" {
		t.Fatalf("optional costs %+v, want Offspring {2}", spec.OptionalCosts)
	}
}

// --- Junkblade Bruiser --------------------------------------------------------

func TestBlbGLJunkbladeBruiserGrowsOnExpendFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	bruiser := effectstest.BlbB1Push(g, me.ID, "Junkblade Bruiser", "Creature — Raccoon Berserker", effectstest.BlbGLBruiserOracle, 4, 5)
	effectstest.ToMainForCost(t, g)
	effectstest.CastPaid(t, g, me, "Three", "{2}{G}", "CCG")
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.PtOf(t, g, bruiser); pw != 4 || tough != 5 {
		t.Fatalf("three mana spent: the Bruiser is %d/%d, want 4/5", pw, tough)
	}
	effectstest.CastPaid(t, g, me, "One", "{G}", "G")
	effectstest.PassPriorityAroundTable(t, g)
	if pw, tough := effectstest.PtOf(t, g, bruiser); pw != 6 || tough != 6 {
		t.Errorf("the fourth mana was spent: the Bruiser is %d/%d, want 6/6", pw, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bruiser, "trample") {
		t.Error("the Bruiser lacks trample")
	}
}

// --- Kindlespark Duo ----------------------------------------------------------

func TestBlbGLKindlesparkDuoPingsAndUntapsOnANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	duo := effectstest.BlbB1Push(g, me.ID, "Kindlespark Duo", "Creature — Lizard Otter", effectstest.BlbGLKindlesparkOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, duo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Fatal("the Duo's controller was accepted as target opponent")
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, duo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, life-1)
	}
	if c, _ := effectstest.BattlefieldCard(g, duo); !c.Tapped {
		t.Fatal("the Duo should be tapped")
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, duo); c.Tapped {
		t.Error("a noncreature spell did not untap the Duo")
	}
}

// --- Kitnap -------------------------------------------------------------------

func TestBlbGLKitnapStealsTapsAndStunsWithoutTheGift(t *testing.T) {
	for _, promised := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		bear := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
		oppHand := opp.Hand.Size()
		to := uuid.Nil
		if promised {
			to = opp.ID
		}
		effectstest.MustCastWithGift(t, g, "Kitnap", "Enchantment — Aura", effectstest.BlbGLKitnapOracle, effectstest.CardTarget(bear), to)
		for i := 0; i < 3; i++ {
			effectstest.BlbGLSettleTriggerOrder(g, me.ID)
			effectstest.PassPriorityAroundTable(t, g)
		}
		c := effectstest.LayeredCard(t, g, bear)
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
		if got := effectstest.CountersOn(g, bear, game.CounterStun); got != wantStun {
			t.Errorf("promised=%v: %d stun counters, want %d", promised, got, wantStun)
		}
		if got := opp.Hand.Size(); got != wantHand {
			t.Errorf("promised=%v: the opponent has %d cards, want %d", promised, got, wantHand)
		}
	}
}

// --- Kitsa, Otterball Elite ---------------------------------------------------

func TestBlbGLKitsaCopiesASpellOnlyAtPowerThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	kitsa := effectstest.BlbB1Push(g, me.ID, "Kitsa, Otterball Elite", "Legendary Creature — Otter Wizard", effectstest.BlbGLKitsaOracle, 1, 3)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	hand := me.Hand.Size()
	spell := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationOracle, nil)
	params := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}}
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, kitsa, 1, params); err == nil {
		t.Fatal("the copy activated at power 1")
	}
	effectstest.PfAddCounters(g, kitsa, game.CounterPlusOne, 2)
	effectstest.B16Activate(t, g, me.ID, kitsa, 1, params)
	for i := 0; i < 4; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size(); got != hand+4 {
		t.Errorf("hand %d, want %d: Divination and its copy draw two each", got, hand+4)
	}
	for _, kw := range []string{"vigilance", "prowess"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, kitsa, kw) {
			t.Errorf("Kitsa lacks %s", kw)
		}
	}
}

func TestBlbGLKitsaLootsWithItsTapAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	kitsa := effectstest.BlbB1Push(g, me.ID, "Kitsa, Otterball Elite", "Legendary Creature — Otter Wizard", effectstest.BlbGLKitsaOracle, 1, 3)
	top := effectstest.SeedLibrary(me, "Top")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, kitsa, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[0]) {
		t.Fatal("the loot drew nothing")
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Error("the loot asked for no discard")
	}
}

// --- Knightfisher -------------------------------------------------------------

func TestBlbGLKnightfisherMakesAFishForAnotherNontokenBird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Knightfisher", "Creature — Bird Knight", effectstest.BlbGLKnightfisherOracle, 4, 5)
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Fish"); n != 0 {
		t.Fatalf("a Bear made %d Fish, want 0", n)
	}
	effectstest.CastHoldCreature(t, g, "Bird", "Creature — Bird Scout", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Fish"); n != 1 {
		t.Errorf("a nontoken Bird made %d Fish, want 1", n)
	}
}

// --- Lightshell Duo -----------------------------------------------------------

func TestBlbGLLightshellDuoSurveilsTwoAsItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	duo := effectstest.CastHoldCreature(t, g, "Lightshell Duo", "Creature — Rat Otter", effectstest.BlbGLLightshellOracle, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil looks at %d cards, want 2", len(c.ScryCards))
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, duo, "prowess") {
		t.Error("the Duo lacks prowess")
	}
}

// --- Lilysplash Mentor --------------------------------------------------------

func TestBlbGLLilysplashMentorFlickersAnotherCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mentor := effectstest.BlbB1Push(g, me.ID, "Lilysplash Mentor", "Creature — Frog Druid", effectstest.BlbGLMentorOracle, 4, 4)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "G", "U")
	if err := g.ActivateCatalogAbility(me.ID, mentor, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mentor)}); err == nil {
		t.Fatal("the Mentor was accepted as another target creature")
	}
	effectstest.B16Activate(t, g, me.ID, mentor, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the old Bear is still on the battlefield")
	}
	back := effectstest.FindBattlefieldByName(g, "Bear")
	if back == uuid.Nil {
		t.Fatal("the Bear did not return")
	}
	if got := effectstest.PlusOnes(g, back); got != 1 {
		t.Errorf("the Bear returned with %d +1/+1 counters, want 1", got)
	}
}

// --- Longstalk Brawl ----------------------------------------------------------

func TestBlbGLLongstalkBrawlCountersOnlyWithTheGift(t *testing.T) {
	for _, promised := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		mine := effectstest.BlbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 3, 3)
		theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
		to := uuid.Nil
		if promised {
			to = opp.ID
		}
		effectstest.MustCastWithGift(t, g, "Longstalk Brawl", "Sorcery", effectstest.BlbGLBrawlOracle, []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
		}, to)
		effectstest.PassPriorityAroundTable(t, g)
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
		if n, tapped := effectstest.CountTokensNamed(g, opp.ID, "Fish"); n != wantFish || tapped != wantFish {
			t.Errorf("promised=%v: %d Fish (%d tapped), want %d tapped", promised, n, tapped, wantFish)
		}
	}
}

// --- Lupinflower Village ------------------------------------------------------

func TestBlbGLLupinflowerVillageDigsForABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	village := effectstest.BlbB1Push(g, me.ID, "Lupinflower Village", "Land", effectstest.BlbGLVillageOracle, 0, 0)
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
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "W")
	effectstest.B16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(village) {
		t.Fatal("the Village was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look")
	}
	if len(c.ChooseCards) != 1 {
		t.Errorf("%d cards offered, want the Bird only (the Deep Bird is seventh)", len(c.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bird)
	effectstest.PassPriorityAroundTable(t, g)
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
