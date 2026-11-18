package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB13SwiftspearOracle  = "dafd2713-d1bc-474b-b390-d2ff20b5375e"
	rprB13PikeOracle        = "2efea460-c3e1-4578-a587-134e13bdf264"
	rprB13VaultOracle       = "7acd7ef5-0e26-436d-962a-1a391f81c2e8"
	rprB13AssemblerOracle   = "8f36e058-e5fa-48f9-9996-09b77fc193b3"
	rprB13BaubleOracle      = "63afc3d1-7653-476e-838f-fc18d4a62a21"
	rprB13WellOracle        = "b0394cf2-12a0-4d4f-87e0-fe8937e6faff"
	rprB13SigilOracle       = "b9447f25-f285-4b62-8976-4a49363dfdd7"
	rprB13MeekOracle        = "215c287d-56a5-46da-b49e-8524b6d320a4"
	rprB13AlelaOracle       = "936ae5dc-9838-47f3-ba1f-66523a7f5b76"
	rprB13PrecursorOracle   = "668fc293-1b8f-4870-9784-4ac660925e30"
	rprB13SunderingOracle   = "d17d071d-dbee-4d65-8435-137f35cd89ca"
	rprB13GiantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
)

func rprB13OnBattlefield(g *game.Game, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = g.Battlefield.Contains(id) })
	return ok
}

func rprB13Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return b12Permanent(g, owner, name, typeLine)
}

func TestRprB13MonasterySwiftspearHasHasteAndProwess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Monastery Swiftspear", "Creature — Human Monk", rprB13SwiftspearOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, id, "haste") || !hasEffectiveKeyword(t, g, id, game.KeywordProwess) {
		t.Fatal("Swiftspear should have haste and prowess")
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, id); p != 2 {
		t.Errorf("a noncreature spell pumps it to 2/3; power = %d", p)
	}
}

func TestRprB13RunechantersPikeCountsInstantsAndSorceriesInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	graveyardCardOf(g, me, "Opt", "Instant", "", "")
	graveyardCardOf(g, me, "Divination", "Sorcery", "", "")
	graveyardCardOf(g, me, "Bear Cub", "Creature — Bear", "", "")
	graveyardCardOf(g, opp, "Shock", "Instant", "", "")
	pike := seedEquipment(g, me.ID, "Runechanter's Pike", rprB13PikeOracle)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Fatalf("power unequipped = %d, want 2", got)
	}
	equipTo(t, g, me.ID, pike, bear)
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("equipped creature should have first strike")
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 4 (two of your instants and sorceries; not the opponent's or the creature)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("toughness = %d, want 2 (+X/+0)", got)
	}
}

func TestRprB13PerilousVaultExilesEveryNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vault := pushCatalogPermanent(g, me.ID, "Perilous Vault", "Artifact", rprB13VaultOracle, false)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	art := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := rprB13Land(g, me.ID, "Forest", "Basic Land — Forest")

	// {5}, {T}: a pool that can't pay is refused.
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activating with an empty pool should be refused")
	}
	fillPool(me, 5)
	b16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{})
	for name, id := range map[string]uuid.UUID{"vault": vault, "my creature": mine, "their creature": theirs, "artifact": art} {
		if rprB13OnBattlefield(g, id) {
			t.Errorf("%s should have been exiled", name)
		}
	}
	if !rprB13OnBattlefield(g, land) {
		t.Error("lands stay")
	}
}

func TestRprB13SelfAssemblerSearchesForAnAssemblyWorker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Clue Assembler", TypeLine: "Artifact Creature — Assembly-Worker"},
		game.Card{Name: "Assembly Blueprint", TypeLine: "Artifact — Assembly-Worker"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	castCatalogSpell(t, g, "Self-Assembler", "Artifact Creature — Assembly-Worker", rprB13AssemblerOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the ETB search is optional, so it always asks")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil || searchOptionNamed(g, c, "Assembly Blueprint") != uuid.Nil {
		t.Error("only Assembly-Worker creature cards are offered")
	}
	answerSearchNamed(t, g, me.ID, "Clue Assembler")
	if !b02bHandHasNamed(me, "Clue Assembler") {
		t.Error("the Assembly-Worker reaches the hand")
	}
}

func TestRprB13MishrasBaubleLooksThenDrawsNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bauble := pushCatalogPermanent(g, me.ID, "Mishra's Bauble", "Artifact", rprB13BaubleOracle, false)
	b16Activate(t, g, me.ID, bauble, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	passPriorityAroundTable(t, g)
	if rprB13OnBattlefield(g, bauble) {
		t.Fatal("the Bauble is sacrificed as a cost")
	}
	if pick := chooseCardsChoiceFor(g, me.ID); pick != nil {
		answerChooseCards(t, g, me.ID, pick.ChooseCards...)
	}
	handBefore := me.Hand.Size()
	startSeat := g.Turn.ActiveSeat
	// Nothing is drawn now; the draw is the next turn's upkeep.
	for i := 0; i < 40 && !(g.Turn.ActiveSeat != startSeat && g.Turn.Step == game.StepUpkeep); i++ {
		if got := me.Hand.Size(); got != handBefore {
			t.Fatalf("hand changed to %d before the next upkeep", got)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == startSeat || g.Turn.Step != game.StepUpkeep {
		t.Fatal("never reached the next turn's upkeep")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards at the next upkeep, want 1", got)
	}
}

func TestRprB13WellOfLostDreamsDrawsXUpToTheLifeGained(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Well of Lost Dreams", "Artifact", rprB13WellOracle, false)
	fillPool(me, 5)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Max != 3 {
		t.Errorf("X prompt max = %d, want the 3 life gained", pa.Max)
	}
	pay := payUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{2}" {
		t.Fatalf("payment prompt = %+v, want {2}", pay)
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestRprB13SigilOfValorPumpsOnlyAnAloneAttacker(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		other1 := b12Creature(g, me.ID, "Ally 1", "Creature — Bear", 1, 1)
		b12Creature(g, me.ID, "Ally 2", "Creature — Bear", 1, 1)
		sigil := seedEquipment(g, me.ID, "Sigil of Valor", rprB13SigilOracle)
		advanceToMain(t, g)
		equipTo(t, g, me.ID, sigil, bear)
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		if !alone {
			if err := g.DeclareAttacker(other1, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		lockInAttacks(t, g)
		passPriorityAroundTable(t, g)
		want := 2
		if alone {
			want = 4 // base 2, +1/+1 for each of the two other creatures
		}
		if got := effectivePower(t, g, bear); got != want {
			t.Errorf("alone=%v: power = %d, want %d", alone, got, want)
		}
	}
}

func TestRprB13SwordOfTheMeekReturnsFromGraveyardForAOneOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sword := graveyardCardOf(g, me, "Sword of the Meek", "Artifact — Equipment", "{2}", rprB13MeekOracle)
	big := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: big, Actor: me.ID}) })
	passPriorityAroundTable(t, g)
	if rprB13OnBattlefield(g, sword) {
		t.Fatal("a 2/2 entering must not bring the Sword back")
	}
	mouse := b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: mouse, Actor: me.ID}) })
	g.RunStateChecksForTest()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !rprB13OnBattlefield(g, sword) {
		t.Fatal("a 1/1 entering brings the Sword back")
	}
	if got := effectivePower(t, g, mouse); got != 2 {
		t.Errorf("the Mouse is %d power, want 2 (attached, +1/+2)", got)
	}
	if got := effectiveToughness(t, g, mouse); got != 3 {
		t.Errorf("the Mouse is %d toughness, want 3", got)
	}
}

func TestRprB13AlelaMakesAFaerieAndPumpsFliers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	alela := b12Push(g, me.ID, "Alela, Artful Provocateur", "Legendary Creature — Faerie Warlock", rprB13AlelaOracle, 2, 3)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	walker := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, flier); got != 2 {
		t.Errorf("flier power = %d, want 2", got)
	}
	if got := effectivePower(t, g, walker); got != 2 {
		t.Errorf("a non-flier is not pumped; power = %d", got)
	}
	if got := effectivePower(t, g, alela); got != 2 {
		t.Errorf("Alela does not pump herself; power = %d", got)
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Faerie"); n != 1 {
		t.Fatalf("faeries = %d, want 1", n)
	}
	castCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Faerie"); n != 1 {
		t.Errorf("a creature spell must not make a Faerie; faeries = %d", n)
	}
}

func TestRprB13PrecursorGolemMakesTwoGolems(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Precursor Golem", "Artifact Creature — Golem", rprB13PrecursorOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Golem"); n != 2 {
		t.Errorf("Golem tokens = %d, want 2", n)
	}
}

func TestRprB13PrecursorGolemCopiesASpellForEveryOtherGolem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pre := b12Push(g, opp.ID, "Precursor Golem", "Artifact Creature — Golem", rprB13PrecursorOracle, 3, 3)
	a := b12Creature(g, me.ID, "Golem A", "Artifact Creature — Golem", 3, 3)
	b := b12Creature(g, opp.ID, "Golem B", "Artifact Creature — Golem", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", rprB13GiantGrowthOracle, cardTarget(a))
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no order prompt for the caster")
	}
	if len(pick.ChooseCards) != 2 || !hasID(pick.ChooseCards, pre) || !hasID(pick.ChooseCards, b) {
		t.Fatalf("order prompt over %v, want the two other Golems", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, b, pre)
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	for name, id := range map[string]uuid.UUID{"A": a, "B": b, "Precursor": pre} {
		if p := effectivePower(t, g, id); p != 6 {
			t.Errorf("Golem %s is %d power, want 6", name, p)
		}
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Golem is %d power, want 2", p)
	}
}

func TestRprB13PrecursorGolemIgnoresASpellOnANonGolem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pre := b12Push(g, me.ID, "Precursor Golem", "Artifact Creature — Golem", rprB13PrecursorOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", rprB13GiantGrowthOracle, cardTarget(bear))
	if n := triggersOnStackFrom(g, pre); n != 0 {
		t.Errorf("a spell targeting a non-Golem triggered it %d times", n)
	}
}

func TestRprB13SunderingTitanDestroysALandOfEachBasicType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	plains1 := rprB13Land(g, me.ID, "Plains", "Basic Land — Plains")
	plains2 := rprB13Land(g, opp.ID, "Plains", "Basic Land — Plains")
	island := rprB13Land(g, opp.ID, "Island", "Basic Land — Island")
	swampMountain := rprB13Land(g, me.ID, "Badlands", "Land — Swamp Mountain")
	forest := rprB13Land(g, me.ID, "Forest", "Basic Land — Forest")
	forest2 := rprB13Land(g, opp.ID, "Forest", "Basic Land — Forest")
	island2 := rprB13Land(g, me.ID, "Island", "Basic Land — Island")
	waste := rprB13Land(g, me.ID, "Wastes", "Basic Land")
	titan := b12Push(g, me.ID, "Sundering Titan", "Artifact Creature — Golem", rprB13SunderingOracle, 7, 10)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: titan, Actor: me.ID}) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	// Plains, Island and Forest have more than one land and ask; the
	// Badlands is the only Swamp and the only Mountain, so it is chosen
	// for both without a prompt.
	answers := []uuid.UUID{plains2, island, forest}
	for i, want := range answers {
		pick := chooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatalf("no choice for leg %d", i)
		}
		if !hasID(pick.ChooseCards, want) {
			t.Fatalf("leg %d does not offer the expected land: %v", i, pick.ChooseCards)
		}
		answerChooseCards(t, g, me.ID, want)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"chosen Plains": plains2, "Island": island, "dual": swampMountain, "Forest": forest} {
		if rprB13OnBattlefield(g, id) {
			t.Errorf("%s should be destroyed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"other Plains": plains1, "other Island": island2, "other Forest": forest2, "nonbasic-typed land": waste, "Titan": titan} {
		if !rprB13OnBattlefield(g, id) {
			t.Errorf("%s should remain", name)
		}
	}
}
