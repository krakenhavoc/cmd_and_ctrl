package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// usg_batch3_cards_test.go — Urza's Saga, batch 3 (#5624). Each test
// plays the real catalog entry and asserts what its text says.

const (
	usgB3BarrinsCodexOracle  = "f7c237e6-17a4-4e72-8cc3-456aec738394"
	usgB3BereavementOracle   = "fae834ee-2b2e-4f68-b4a4-81592cbecefb"
	usgB3BrandOracle         = "eec7500b-418d-4387-beb2-d53e047c1d22"
	usgB3BulwarkOracle       = "259da825-02f7-4e72-8f44-01683330f457"
	usgB3CathodionOracle     = "fcd4f816-2de1-4b30-82fb-cb87f45747ea"
	usgB3CaveTigerOracle     = "d3bac8b8-0b2a-4c67-b087-7a0bf1359468"
	usgB3ChildOfGaeaOracle   = "d4e8aa93-c0d1-49b7-bbf4-35a155048774"
	usgB3ChimericStaffOracle = "771f5e29-b4d1-4466-8510-ce3f89ad7173"
	usgB3CitanulFluteOracle  = "6b8183bb-20c4-4d45-b944-508a6a874bb3"
	usgB3ContaminationOracle = "4013b4c2-c9ed-4d14-90f0-97214ec1fded"
	usgB3CopperGnomesOracle  = "4fd6d9af-4d53-4e10-bc23-672454803099"
	usgB3CorruptOracle       = "a88a1c53-7de3-423a-a95a-24ed5666ab96"
	usgB3DarkHatchlingOracle = "36633b43-855b-4620-8afb-70c39fc07280"
	usgB3DarkestHourOracle   = "5667376e-e59c-4b17-b096-5d92cdfe3db1"
)

func TestUsgB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		usgB3BarrinsCodexOracle:  "Barrin's Codex",
		usgB3BereavementOracle:   "Bereavement",
		usgB3BrandOracle:         "Brand",
		usgB3BulwarkOracle:       "Bulwark",
		usgB3CathodionOracle:     "Cathodion",
		usgB3CaveTigerOracle:     "Cave Tiger",
		usgB3ChildOfGaeaOracle:   "Child of Gaea",
		usgB3ChimericStaffOracle: "Chimeric Staff",
		usgB3CitanulFluteOracle:  "Citanul Flute",
		usgB3ContaminationOracle: "Contamination",
		usgB3CopperGnomesOracle:  "Copper Gnomes",
		usgB3CorruptOracle:       "Corrupt",
		usgB3DarkHatchlingOracle: "Dark Hatchling",
		usgB3DarkestHourOracle:   "Darkest Hour",
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

// usgB3Colored seeds a creature of the given colours, not summoning sick.
func usgB3Colored(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Beast",
		Colors: colors, Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

// usgB3Hand empties a player's hand and fills it with n plain cards.
func usgB3Hand(p *game.Player, n int) {
	p.Hand.Cards = nil
	for range n {
		pushCatalogHandCard(p, "Filler", "Sorcery", "")
	}
}

// usgB3Destroy destroys a permanent and lets what it triggered resolve.
func usgB3Destroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
}

// --- Barrin's Codex ------------------------------------------------------------------

func TestUsgB3BarrinsCodexMayAddAPageEachUpkeep(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		codex := pushCatalogPermanent(g, me.ID, "Barrin's Codex", "Artifact — Book", usgB3BarrinsCodexOracle, false)
		walkToUpkeepOfSeat0(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)
		want := 0
		if accept {
			want = 1
		}
		if got := pr8Counters(g, codex, "page"); got != want {
			t.Errorf("accept=%v: %d page counters, want %d", accept, got, want)
		}
	}
}

func TestUsgB3BarrinsCodexDrawsOneCardPerPage(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	codex := pushCatalogPermanent(g, me.ID, "Barrin's Codex", "Artifact — Book", usgB3BarrinsCodexOracle, false)
	addCountersForTest(g, codex, "page", 3)
	if err := usgB2Activate(t, g, me, codex, "CCC", nil); err == nil {
		t.Fatal("activated for {3}; it costs {4}")
	}
	hand := me.Hand.Size()
	if err := usgB2Activate(t, g, me, codex, "C", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(codex) || !me.Graveyard.Contains(codex) {
		t.Error("the Codex was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3 (the page counters on the sacrificed Codex)", got)
	}
}

// --- Bereavement ----------------------------------------------------------------------

func TestUsgB3BereavementMakesAGreenCreaturesControllerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedEnchantment(g, usgB3BereavementOracle, "Bereavement", "Enchantment", me.ID)
	usgB3Hand(opp, 2)
	usgB3Hand(me, 2)
	red := usgB3Colored(g, opp.ID, "Red Beast", "R")
	usgB3Destroy(t, g, red)
	if discardChoiceFor(g, opp.ID) != nil {
		t.Fatal("a red creature dying asked for a discard")
	}
	green := usgB3Colored(g, opp.ID, "Green Beast", "G")
	usgB3Destroy(t, g, green)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("Bereavement's controller was asked to discard")
	}
	pick := opp.Hand.Cards[0].InstanceID
	answerDiscard(t, g, opp.ID, pick)
	if opp.Hand.Size() != 1 || !opp.Graveyard.Contains(pick) {
		t.Errorf("the green creature's controller has %d cards; the picked card is not in their graveyard", opp.Hand.Size())
	}
}

func TestUsgB3BereavementReadsTheColourTheCreatureDiedWith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedEnchantment(g, usgB3BereavementOracle, "Bereavement", "Enchantment", me.ID)
	usgB3Hand(opp, 2)
	// A printed-green creature painted black is not green as it dies.
	painted := pushCardsColorByEffect(t, g, opp.ID)
	usgB3Destroy(t, g, painted)
	if discardChoiceFor(g, opp.ID) != nil {
		t.Error("a creature that died black was read as green")
	}
}

// --- Brand -------------------------------------------------------------------------------

func TestUsgB3BrandTakesBackEveryPermanentYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stolen := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, stolen, opp.ID, game.IndefiniteDuration(), "test theft")
	})
	if controllerOf(t, g, stolen) != opp.ID {
		t.Fatal("setup: the opponent does not control the Bear")
	}
	castCatalogSpell(t, g, "Brand", "Instant", usgB3BrandOracle, nil)
	passPriorityAroundTable(t, g)
	if controllerOf(t, g, stolen) != me.ID {
		t.Error("Brand did not return control of a permanent you own")
	}
	if controllerOf(t, g, theirs) != opp.ID {
		t.Error("Brand took a permanent you don't own")
	}
}

// --- Bulwark -------------------------------------------------------------------------------

func TestUsgB3BulwarkDealsTheHandDifference(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		want         int
	}{
		{"you hold three more", 4, 1, 3},
		{"they hold more", 1, 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[2]
			seedEnchantment(g, usgB3BulwarkOracle, "Bulwark", "Enchantment", me.ID)
			walkToUpkeepOfSeat0(t, g)
			pickPlayer(t, g, me.ID, opp.ID)
			usgB3Hand(me, tc.mine)
			usgB3Hand(opp, tc.theirs)
			life, myLife := opp.Life, me.Life
			passPriorityAroundTable(t, g)
			if got := life - opp.Life; got != tc.want {
				t.Errorf("the target opponent took %d, want %d", got, tc.want)
			}
			if me.Life != myLife {
				t.Errorf("Bulwark's controller lost life: %d → %d", myLife, me.Life)
			}
		})
	}
}

func TestUsgB3BulwarkCantTargetYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedEnchantment(g, usgB3BulwarkOracle, "Bulwark", "Enchantment", me.ID)
	walkToUpkeepOfSeat0(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for Bulwark's trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("Bulwark's trigger targeted its controller")
	}
}

// --- Cathodion -------------------------------------------------------------------------------

func TestUsgB3CathodionAddsThreeColorlessWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cathodion := b12Push(g, me.ID, "Cathodion", "Artifact Creature — Construct", usgB3CathodionOracle, 3, 3)
	usgB3Destroy(t, g, cathodion)
	if got := poolColors(me); !reflect.DeepEqual(got, []string{"C", "C", "C"}) {
		t.Errorf("pool %v, want C C C", got)
	}
}

// --- Cave Tiger ---------------------------------------------------------------------------------

func TestUsgB3CaveTigerGrowsOncePerBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tiger := b12Push(g, me.ID, "Cave Tiger", "Creature — Cat", usgB3CaveTigerOracle, 2, 2)
	a := b12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
	b := b12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
	woeS2Block(g, opp.ID, a, tiger)
	woeS2Block(g, opp.ID, b, tiger)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, tiger); p != 4 || tough != 4 {
		t.Errorf("blocked by two creatures the Tiger is %d/%d, want 4/4", p, tough)
	}
}

func TestUsgB3CaveTigerBlockingIsNotBecomingBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tiger := b12Push(g, opp.ID, "Cave Tiger", "Creature — Cat", usgB3CaveTigerOracle, 2, 2)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	woeS2Block(g, opp.ID, tiger, attacker)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, tiger); p != 2 || tough != 2 {
		t.Errorf("a blocking Tiger is %d/%d, want 2/2", p, tough)
	}
}

// --- Child of Gaea ---------------------------------------------------------------------------------

func TestUsgB3ChildOfGaeaHasTrampleAndAnUpkeep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		child := b12Push(g, me.ID, "Child of Gaea", "Creature — Elemental", usgB3ChildOfGaeaOracle, 7, 7)
		if !hasKeywordOnBattlefield(t, g, child, "trample") {
			t.Error("Child of Gaea lacks trample")
		}
		walkToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		if pay {
			dcMana(t, g, me.ID, "{G}{G}")
		}
		answerPayUnless(t, g, me.ID, pay)
		if g.Battlefield.Contains(child) != pay {
			t.Errorf("pay=%v: on the battlefield = %v", pay, g.Battlefield.Contains(child))
		}
	}
}

func TestUsgB3ChildOfGaeaRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	child := b12Push(g, me.ID, "Child of Gaea", "Creature — Elemental", usgB3ChildOfGaeaOracle, 7, 7)
	if err := usgB2Activate(t, g, me, child, "G", nil); err == nil {
		t.Fatal("regenerated for {G}; it costs {1}{G}")
	}
	if err := usgB2Activate(t, g, me, child, "C", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	usgB3Destroy(t, g, child)
	if !g.Battlefield.Contains(child) {
		t.Fatal("the regenerated Child of Gaea was destroyed")
	}
	if !b13Tapped(t, g, child) {
		t.Error("a regenerated creature is tapped")
	}
}

// --- Chimeric Staff ---------------------------------------------------------------------------------

func TestUsgB3ChimericStaffBecomesAnXByXConstruct(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	staff := pushCatalogPermanent(g, me.ID, "Chimeric Staff", "Artifact", usgB3ChimericStaffOracle, false)
	floatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("activate X=3: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, staff)
	if !ok || !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Construct") {
		t.Fatalf("the Staff is not a Construct artifact creature: %+v", c.TypeLine)
	}
	if len(c.Colors) != 0 {
		t.Errorf("the Staff gained colours %v", c.Colors)
	}
	if p, tough := effectivePTOf(t, g, staff); p != 3 || tough != 3 {
		t.Errorf("the Staff is %d/%d, want 3/3", p, tough)
	}
}

// --- Citanul Flute ---------------------------------------------------------------------------------

func TestUsgB3CitanulFluteFindsACreatureUpToX(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{4}{G}{G}"},
		game.Card{Name: "Cheap Spell", TypeLine: "Sorcery", ManaCost: "{G}"},
	)
	flute := pushCatalogPermanent(g, me.ID, "Citanul Flute", "Artifact", usgB3CitanulFluteOracle, false)
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, flute, 0, game.ActivateAbilityParams{XValue: 2, Strict: true}); err != nil {
		t.Fatalf("activate X=2: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Big Wurm") != uuid.Nil || searchOptionNamed(g, c, "Cheap Spell") != uuid.Nil {
		t.Error("offered a card that isn't a creature card with mana value 2 or less")
	}
	answerSearchNamed(t, g, me.ID, "Small Bear")
	if !b02bHandHasNamed(me, "Small Bear") {
		t.Error("the found creature card is not in hand")
	}
	if !b13Tapped(t, g, flute) {
		t.Error("the Flute did not tap")
	}
}

// --- Contamination ---------------------------------------------------------------------------------

func TestUsgB3ContaminationMakesEveryLandTapForOneBlack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedEnchantment(g, usgB3ContaminationOracle, "Contamination", "Enchantment", me.ID)
	forest := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	tapForMana(t, g, opp.ID, forest)
	if got := sortedColors(opp); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("an opponent's Forest made %v, want [B]", got)
	}
	tomb := seedPermanentWithOracle(g, me.ID, "Ancient Tomb", "Land", ancientTombOracle)
	tapForMana(t, g, me.ID, tomb)
	if got := sortedColors(me); !reflect.DeepEqual(got, []string{"B"}) {
		t.Errorf("Ancient Tomb made %v, want one [B]", got)
	}
}

func TestUsgB3ContaminationUpkeepSacrificesACreatureOrItself(t *testing.T) {
	t.Run("a creature", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		cont := seedEnchantment(g, usgB3ContaminationOracle, "Contamination", "Enchantment", me.ID)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		walkToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		c := latestOptionPickFor(g, me.ID)
		if c == nil || len(c.PickOptions) != 2 {
			t.Fatal("no two-way choice between a creature and Contamination")
		}
		sweep05OptionKey(t, g, me.ID, "option-pick/sacrifice-this-unless-you-sacrifice-a-creature")
		if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
			t.Fatalf("pick the creature: %v", err)
		}
		answerChooseCards(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) {
			t.Error("the creature was not sacrificed")
		}
		if !g.Battlefield.Contains(cont) {
			t.Error("Contamination was sacrificed after a creature was")
		}
	})
	t.Run("no creature", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		cont := seedEnchantment(g, usgB3ContaminationOracle, "Contamination", "Enchantment", me.ID)
		b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		walkToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(cont) {
			t.Error("with no creature to sacrifice, Contamination stayed")
		}
	})
}

// --- Copper Gnomes ---------------------------------------------------------------------------------

func TestUsgB3CopperGnomesPutsAnArtifactFromHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gnomes := b12Push(g, me.ID, "Copper Gnomes", "Artifact Creature — Gnome", usgB3CopperGnomesOracle, 1, 1)
	rock := handCardForTest(me, "Big Rock", "Artifact", "")
	bear := handCardForTest(me, "Bear", "Creature — Bear", "")
	if err := usgB2Activate(t, g, me, gnomes, "CCCC", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(gnomes) {
		t.Error("the Gnomes were not sacrificed")
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-from-hand prompt")
	}
	if pick.ChooseMin != 0 || hasID(pick.ChooseCards, bear) || !hasID(pick.ChooseCards, rock) {
		t.Errorf("offered %v (min %d), want only the artifact card, optionally", pick.ChooseCards, pick.ChooseMin)
	}
	answerChooseCards(t, g, me.ID, rock)
	if !g.Battlefield.Contains(rock) || me.Hand.Contains(rock) {
		t.Error("the artifact card did not enter the battlefield")
	}
}

// --- Corrupt ---------------------------------------------------------------------------------------

func TestUsgB3CorruptDrainsBySwamps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for range 3 {
		b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	b12Permanent(g, opp.ID, "Their Swamp", "Basic Land — Swamp")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	life, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Corrupt", "Sorcery", usgB3CorruptOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Errorf("the target lost %d, want 3 (your three Swamps)", oppLife-opp.Life)
	}
	if me.Life != life+3 {
		t.Errorf("you gained %d, want 3", me.Life-life)
	}
}

// --- Dark Hatchling --------------------------------------------------------------------------------

func TestUsgB3DarkHatchlingDestroysANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	black := usgB3Colored(g, opp.ID, "Black Beast", "B")
	green := usgB3Colored(g, opp.ID, "Green Beast", "G")
	hatchling := castAndResolveCreature(t, g, "Dark Hatchling", "Creature — Horror", usgB3DarkHatchlingOracle)
	if !hasKeywordOnBattlefield(t, g, hatchling, "flying") {
		t.Error("Dark Hatchling lacks flying")
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: black}); err == nil {
		t.Fatal("the trigger targeted a black creature")
	}
	pickTriggerTarget(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature was not destroyed")
	}
	if !g.Battlefield.Contains(black) {
		t.Error("the black creature was destroyed")
	}
}

// --- Darkest Hour ----------------------------------------------------------------------------------

func TestUsgB3DarkestHourMakesEveryCreatureBlackOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := usgB3Colored(g, opp.ID, "Green Beast", "G")
	gold := usgB3Colored(g, me.ID, "Gold Beast", "W", "U")
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Rock", TypeLine: "Artifact", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	seedEnchantment(g, usgB3DarkestHourOracle, "Darkest Hour", "Enchantment", me.ID)
	g.ReadSnapshot(func() {})
	for _, id := range []uuid.UUID{green, gold} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.HasColor("B") {
			t.Errorf("%s is not black", c.Name)
		}
		for _, other := range []string{"W", "U", "R", "G"} {
			if c.HasColor(other) {
				t.Errorf("%s kept colour %s", c.Name, other)
			}
		}
	}
	if c, _ := battlefieldCard(g, rock); c.HasColor("B") || !c.HasColor("R") {
		t.Error("a noncreature artifact changed colour")
	}
}
