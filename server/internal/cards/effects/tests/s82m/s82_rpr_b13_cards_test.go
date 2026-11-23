package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rprB13OnBattlefield(g *game.Game, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = g.Battlefield.Contains(id) })
	return ok
}

func rprB13Land(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return effectstest.B12Permanent(g, owner, name, typeLine)
}

func TestRprB13MonasterySwiftspearHasHasteAndProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Monastery Swiftspear", "Creature — Human Monk", effectstest.RprB13SwiftspearOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") || !effectstest.HasEffectiveKeyword(t, g, id, game.KeywordProwess) {
		t.Fatal("Swiftspear should have haste and prowess")
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("a noncreature spell pumps it to 2/3; power = %d", p)
	}
}

func TestRprB13RunechantersPikeCountsInstantsAndSorceriesInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.GraveyardCardOf(g, me, "Opt", "Instant", "", "")
	effectstest.GraveyardCardOf(g, me, "Divination", "Sorcery", "", "")
	effectstest.GraveyardCardOf(g, me, "Bear Cub", "Creature — Bear", "", "")
	effectstest.GraveyardCardOf(g, opp, "Shock", "Instant", "", "")
	pike := effectstest.SeedEquipment(g, me.ID, "Runechanter's Pike", effectstest.RprB13PikeOracle)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Fatalf("power unequipped = %d, want 2", got)
	}
	effectstest.EquipTo(t, g, me.ID, pike, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("equipped creature should have first strike")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("power = %d, want 4 (two of your instants and sorceries; not the opponent's or the creature)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("toughness = %d, want 2 (+X/+0)", got)
	}
}

func TestRprB13PerilousVaultExilesEveryNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Perilous Vault", "Artifact", effectstest.RprB13VaultOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	art := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	land := rprB13Land(g, me.ID, "Forest", "Basic Land — Forest")

	// {5}, {T}: a pool that can't pay is refused.
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activating with an empty pool should be refused")
	}
	effectstest.FillPool(me, 5)
	effectstest.B16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{})
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Clue Assembler", TypeLine: "Artifact Creature — Assembly-Worker"},
		game.Card{Name: "Assembly Blueprint", TypeLine: "Artifact — Assembly-Worker"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Self-Assembler", "Artifact Creature — Assembly-Worker", effectstest.RprB13AssemblerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the ETB search is optional, so it always asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Assembly Blueprint") != uuid.Nil {
		t.Error("only Assembly-Worker creature cards are offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Clue Assembler")
	if !effectstest.B02bHandHasNamed(me, "Clue Assembler") {
		t.Error("the Assembly-Worker reaches the hand")
	}
}

func TestRprB13MishrasBaubleLooksThenDrawsNextUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bauble := effectstest.PushCatalogPermanent(g, me.ID, "Mishra's Bauble", "Artifact", effectstest.RprB13BaubleOracle, false)
	effectstest.B16Activate(t, g, me.ID, bauble, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if rprB13OnBattlefield(g, bauble) {
		t.Fatal("the Bauble is sacrificed as a cost")
	}
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		effectstest.AnswerChooseCards(t, g, me.ID, pick.ChooseCards...)
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("drew %d cards at the next upkeep, want 1", got)
	}
}

func TestRprB13WellOfLostDreamsDrawsXUpToTheLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Well of Lost Dreams", "Artifact", effectstest.RprB13WellOracle, false)
	effectstest.FillPool(me, 5)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Max != 3 {
		t.Errorf("X prompt max = %d, want the 3 life gained", pa.Max)
	}
	pay := effectstest.PayUnlessFor(g, me.ID)
	if pay == nil || pay.PayCost != "{2}" {
		t.Fatalf("payment prompt = %+v, want {2}", pay)
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestRprB13SigilOfValorPumpsOnlyAnAloneAttacker(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		other1 := effectstest.B12Creature(g, me.ID, "Ally 1", "Creature — Bear", 1, 1)
		effectstest.B12Creature(g, me.ID, "Ally 2", "Creature — Bear", 1, 1)
		sigil := effectstest.SeedEquipment(g, me.ID, "Sigil of Valor", effectstest.RprB13SigilOracle)
		effectstest.AdvanceToMain(t, g)
		effectstest.EquipTo(t, g, me.ID, sigil, bear)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		if !alone {
			if err := g.DeclareAttacker(other1, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if alone {
			want = 4 // base 2, +1/+1 for each of the two other creatures
		}
		if got := effectstest.EffectivePower(t, g, bear); got != want {
			t.Errorf("alone=%v: power = %d, want %d", alone, got, want)
		}
	}
}

func TestRprB13SwordOfTheMeekReturnsFromGraveyardForAOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sword := effectstest.GraveyardCardOf(g, me, "Sword of the Meek", "Artifact — Equipment", "{2}", effectstest.RprB13MeekOracle)
	big := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: big, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if rprB13OnBattlefield(g, sword) {
		t.Fatal("a 2/2 entering must not bring the Sword back")
	}
	mouse := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: mouse, Actor: me.ID}) })
	g.RunStateChecksForTest()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !rprB13OnBattlefield(g, sword) {
		t.Fatal("a 1/1 entering brings the Sword back")
	}
	if got := effectstest.EffectivePower(t, g, mouse); got != 2 {
		t.Errorf("the Mouse is %d power, want 2 (attached, +1/+2)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, mouse); got != 3 {
		t.Errorf("the Mouse is %d toughness, want 3", got)
	}
}

func TestRprB13AlelaMakesAFaerieAndPumpsFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alela := effectstest.B12Push(g, me.ID, "Alela, Artful Provocateur", "Legendary Creature — Faerie Warlock", effectstest.RprB13AlelaOracle, 2, 3)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	walker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, flier); got != 2 {
		t.Errorf("flier power = %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, walker); got != 2 {
		t.Errorf("a non-flier is not pumped; power = %d", got)
	}
	if got := effectstest.EffectivePower(t, g, alela); got != 2 {
		t.Errorf("Alela does not pump herself; power = %d", got)
	}
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Faerie"); n != 1 {
		t.Fatalf("faeries = %d, want 1", n)
	}
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Faerie"); n != 1 {
		t.Errorf("a creature spell must not make a Faerie; faeries = %d", n)
	}
}

func TestRprB13PrecursorGolemMakesTwoGolems(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Precursor Golem", "Artifact Creature — Golem", effectstest.RprB13PrecursorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountNamed(g, "Golem"); n != 2 {
		t.Errorf("Golem tokens = %d, want 2", n)
	}
}

func TestRprB13PrecursorGolemCopiesASpellForEveryOtherGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pre := effectstest.B12Push(g, opp.ID, "Precursor Golem", "Artifact Creature — Golem", effectstest.RprB13PrecursorOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Golem A", "Artifact Creature — Golem", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Golem B", "Artifact Creature — Golem", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.RprB13GiantGrowthOracle, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no order prompt for the caster")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, pre) || !effectstest.HasID(pick.ChooseCards, b) {
		t.Fatalf("order prompt over %v, want the two other Golems", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b, pre)
	for i := 0; i < 4 && len(g.Stack.Cards) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	for name, id := range map[string]uuid.UUID{"A": a, "B": b, "Precursor": pre} {
		if p := effectstest.EffectivePower(t, g, id); p != 6 {
			t.Errorf("Golem %s is %d power, want 6", name, p)
		}
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("a non-Golem is %d power, want 2", p)
	}
}

func TestRprB13PrecursorGolemIgnoresASpellOnANonGolem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pre := effectstest.B12Push(g, me.ID, "Precursor Golem", "Artifact Creature — Golem", effectstest.RprB13PrecursorOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.RprB13GiantGrowthOracle, effectstest.CardTarget(bear))
	if n := effectstest.TriggersOnStackFrom(g, pre); n != 0 {
		t.Errorf("a spell targeting a non-Golem triggered it %d times", n)
	}
}

func TestRprB13SunderingTitanDestroysALandOfEachBasicType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	plains1 := rprB13Land(g, me.ID, "Plains", "Basic Land — Plains")
	plains2 := rprB13Land(g, opp.ID, "Plains", "Basic Land — Plains")
	island := rprB13Land(g, opp.ID, "Island", "Basic Land — Island")
	swampMountain := rprB13Land(g, me.ID, "Badlands", "Land — Swamp Mountain")
	forest := rprB13Land(g, me.ID, "Forest", "Basic Land — Forest")
	forest2 := rprB13Land(g, opp.ID, "Forest", "Basic Land — Forest")
	island2 := rprB13Land(g, me.ID, "Island", "Basic Land — Island")
	waste := rprB13Land(g, me.ID, "Wastes", "Basic Land")
	titan := effectstest.B12Push(g, me.ID, "Sundering Titan", "Artifact Creature — Golem", effectstest.RprB13SunderingOracle, 7, 10)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: titan, Actor: me.ID}) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// Plains, Island and Forest have more than one land and ask; the
	// Badlands is the only Swamp and the only Mountain, so it is chosen
	// for both without a prompt.
	answers := []uuid.UUID{plains2, island, forest}
	for i, want := range answers {
		pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
		if pick == nil {
			t.Fatalf("no choice for leg %d", i)
		}
		if !effectstest.HasID(pick.ChooseCards, want) {
			t.Fatalf("leg %d does not offer the expected land: %v", i, pick.ChooseCards)
		}
		effectstest.AnswerChooseCards(t, g, me.ID, want)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
