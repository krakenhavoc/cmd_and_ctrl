package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b30_a_cards_test.go — Avatar: The Last Airbender Eternal batch 30,
// group a (#5075).

const (
	avB30aLoAndLiOracle  = "b5018037-cf4e-46d2-86c3-fb8eec6e44d1"
	avB30aRuthlessOracle = "100771fd-07db-47be-ada3-935dc918fa33"
	avB30aScarringOracle = "63957ec0-89e6-4be2-8740-846ca56ee48c"
	avB30aDeerDogOracle  = "76c7a339-015b-4074-8e10-14d52b784f9b"
	avB30aOmashuOracle   = "ea58fb10-3d70-4c25-b317-615dfb20a48d"
	avB30aFranticOracle  = "a866cc05-69b0-45f2-8829-b6ef08193ed1"
	avB30aRecruitOracle  = "33e66662-aea6-480c-b69b-1907d987b69e"
	avB30aLongshotOracle = "f93c3468-5b1d-4fa7-a281-1009ea0f2062"
	avB30aLostOracle     = "69ac3757-66f8-47d4-af7b-dda66c96c254"
	avB30aMokuOracle     = "d6e1b1ff-e4e5-4814-9a20-2d527014463f"
)

func TestAvatarB30aCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB30aLoAndLiOracle:  "Lo and Li, Royal Advisors",
		avB30aRuthlessOracle: "Ruthless Waterbender",
		avB30aScarringOracle: "Scarring Memories",
		avB30aDeerDogOracle:  "Deer-Dog",
		avB30aOmashuOracle:   "Founding of Omashu",
		avB30aFranticOracle:  "Frantic Confrontation",
		avB30aRecruitOracle:  "Freedom Fighter Recruit",
		avB30aLongshotOracle: "Longshot, Rebel Bowman",
		avB30aLostOracle:     "Lost in Memories",
		avB30aMokuOracle:     "Moku, Meandering Drummer",
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

func avB30aOpp(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }

// --- Lo and Li, Royal Advisors ---------------------------------------------------

func TestAvatarB30LoAndLiCountsEveryCardAnOpponentDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	loli := effectstest.B12Push(g, me.ID, "Lo and Li, Royal Advisors", "Legendary Creature — Human Advisor", avB30aLoAndLiOracle, 3, 3)
	advisor := effectstest.B12Creature(g, me.ID, "Court Advisor", "Creature — Human Advisor", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Advisor", "Creature — Human Advisor", 1, 1)
	var err error
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(opp.ID, 2) })
	if err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{loli: 2, advisor: 2, bear: 0, theirs: 0} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s: %d +1/+1 counters, want %d", id, got, want)
		}
	}
	// Your own discard is not an opponent's.
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(me.ID, 1) })
	if err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, loli, game.CounterPlusOne); got != 2 {
		t.Errorf("your own discard added counters: %d, want 2", got)
	}
}

func TestAvatarB30LoAndLiMillsFourAndAMillIsOneTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	loli := effectstest.B12Push(g, me.ID, "Lo and Li, Royal Advisors", "Legendary Creature — Human Advisor", avB30aLoAndLiOracle, 3, 3)
	theirs := opp.Library.Size()
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, loli, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := theirs - opp.Library.Size(); got != 4 {
		t.Fatalf("the opponent milled %d, want 4", got)
	}
	if got := effectstest.CountersOn(g, loli, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after one mill of four cards, want 1", got)
	}
	// Milling yourself is not an opponent milling.
	mine := me.Library.Size()
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, loli, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := mine - me.Library.Size(); got != 4 {
		t.Fatalf("you milled %d, want 4", got)
	}
	if got := effectstest.CountersOn(g, loli, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after milling yourself, want still 1", got)
	}
}

func TestAvatarB30LoAndLiAbilityRefusesAShortPool(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	loli := effectstest.B12Push(g, me.ID, "Lo and Li, Royal Advisors", "Legendary Creature — Human Advisor", avB30aLoAndLiOracle, 3, 3)
	effectstest.FillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, loli, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("{2}{U/B} was paid with three red mana")
	}
}

// --- Ruthless Waterbender --------------------------------------------------------

func TestAvatarB30RuthlessWaterbenderWaterbendsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wb := effectstest.B12Push(g, me.ID, "Ruthless Waterbender", "Creature — Human Soldier Ally", avB30aRuthlessOracle, 1, 3)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 2)
	if err := avB4Activate(g, me.ID, wb, 0, tappers); err != nil {
		t.Fatalf("waterbend {2}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, wb); p != 2 || th != 4 {
		t.Errorf("Ruthless Waterbender is %d/%d, want 2/4", p, th)
	}
}

func TestAvatarB30RuthlessWaterbenderRefusesOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := avB30aOpp(g)
	wb := effectstest.B12Push(g, opp.ID, "Ruthless Waterbender", "Creature — Human Soldier Ally", avB30aRuthlessOracle, 1, 3)
	tappers := effectstest.PushTapCostSoldiers(g, opp.ID, 2)
	if err := avB4Activate(g, opp.ID, wb, 0, tappers); err == nil {
		t.Fatal("the ability was activated during another player's turn")
	}
}

// --- Scarring Memories -----------------------------------------------------------

func TestAvatarB30ScarringMemoriesSacrificesThenDiscardsThenLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := avB30aOpp(g)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	life, hand := opp.Life, opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Scarring Memories", "Sorcery — Lesson", avB30aScarringOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to sacrifice a creature")
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil || opp.Life != life {
		t.Fatal("the discard or the life loss came before the sacrifice was chosen")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the chosen creature was not the one sacrificed")
	}
	if opp.Life != life {
		t.Fatal("the life loss came before the discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-1)
	}
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestAvatarB30ScarringMemoriesRefusesToTargetYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Scarring Memories", "Sorcery — Lesson", avB30aScarringOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Scarring Memories targeted its caster")
	}
}

func TestAvatarB30ScarringMemoriesHasFlashOnlyWithAnAttackingLegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	// A legendary creature you control that is not attacking grants no flash.
	effectstest.B12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.FillPoolColored(me, "B", 4)
	if _, err := effectstest.CastFromHandErr(g, me, "Scarring Memories", "Sorcery — Lesson", "{3}{B}", avB30aScarringOracle,
		game.CastSpellParams{Strict: true, Targets: targets}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("cast with flash while only a nonlegendary creature attacks: err = %v, want ErrSorcerySpeedRequired", err)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	legend := effectstest.B12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(legend, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.FillPoolColored(me, "B", 4)
	// #5207: the flash is the card's own timing permission, so the
	// printed cost is what the instant-speed cast pays.
	if _, err := effectstest.CastFromHandErr(g, me, "Scarring Memories", "Sorcery — Lesson", "{3}{B}", avB30aScarringOracle,
		game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast with flash while a legendary creature attacks: %v", err)
	}
}

// --- Deer-Dog --------------------------------------------------------------------

func TestAvatarB30DeerDogHasFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dog := effectstest.B12Push(g, me.ID, "Deer-Dog", "Creature — Elk Dog", avB30aDeerDogOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, dog, "first strike") {
		t.Error("Deer-Dog has no first strike")
	}
}

// --- Founding of Omashu ----------------------------------------------------------

func TestAvatarB30FoundingOfOmashuChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	g.WithWriteLock(func() { avB10Fillers(me, 4) })
	saga := effectstest.CastCatalogSpell(t, g, "Founding of Omashu", "Enchantment — Saga", avB30aOmashuOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Ally", me.ID); got != 2 {
		t.Fatalf("%d Allies after chapter I, want 2", got)
	}
	ally := effectstest.FindBattlefieldByName(g, "Ally")
	if p, th := effectstest.EffectivePTOf(t, g, ally); p != 1 || th != 1 {
		t.Errorf("Ally is %d/%d, want 1/1", p, th)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II offered no discard")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a discard and a draw", hand, me.Hand.Size())
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, ally); p != 2 || th != 1 {
		t.Errorf("Ally is %d/%d after chapter III, want 2/1", p, th)
	}
	g.RunStateChecksForTest()
	if effects.OnBattlefield(g, saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestAvatarB30FoundingOfOmashuChapterTwoDeclinedDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	g.WithWriteLock(func() { avB10Fillers(me, 4) })
	effectstest.CastCatalogSpell(t, g, "Founding of Omashu", "Enchantment — Saga", avB30aOmashuOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("chapter II offered no discard")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw without a discard", hand, me.Hand.Size())
	}
}

// --- Frantic Confrontation -------------------------------------------------------

func TestAvatarB30FranticConfrontationPumpsByXWithFirstStrikeAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "R", 4)
	effectstest.CastFromHandForTest(t, g, me, "Frantic Confrontation", "Instant", "{X}{R}", avB30aFranticOracle,
		game.CastSpellParams{Strict: true, XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 5 || th != 2 {
		t.Errorf("Bear is %d/%d, want 5/2", p, th)
	}
	for _, kw := range []string{"first strike", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("Bear has no %s", kw)
		}
	}
}

func TestAvatarB30FranticConfrontationRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "R", 2)
	if _, err := effectstest.CastFromHandErr(g, me, "Frantic Confrontation", "Instant", "{X}{R}", avB30aFranticOracle,
		game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("Frantic Confrontation targeted an opponent's creature")
	}
}

// --- Freedom Fighter Recruit -----------------------------------------------------

func TestAvatarB30FreedomFighterRecruitCountsYourCreaturesInEveryZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	recruit := effectstest.B12Push(g, me.ID, "Freedom Fighter Recruit", "Creature — Human Rebel Ally", avB30aRecruitOracle, 0, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, recruit); p != 1 {
		t.Fatalf("alone: power %d, want 1", p)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, th := effectstest.EffectivePTOf(t, g, recruit); p != 3 || th != 2 {
		t.Errorf("Recruit is %d/%d, want 3/2", p, th)
	}
	// The ability functions in a hand too (the ruling).
	inHand := game.Card{InstanceID: uuid.New(), Name: "Freedom Fighter Recruit", TypeLine: "Creature — Human Rebel Ally",
		OracleID: avB30aRecruitOracle, Toughness: 2, Owner: me.ID, Controller: me.ID}
	var p int
	g.WithWriteLock(func() { p = g.PowerAnywhereForEffect(inHand) })
	if p != 3 {
		t.Errorf("in hand: power %d, want 3", p)
	}
}

// --- Longshot, Rebel Bowman ------------------------------------------------------

func TestAvatarB30LongshotDiscountsNoncreatureSpellsAndPingsEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	longshot := effectstest.B12Push(g, me.ID, "Longshot, Rebel Bowman", "Legendary Creature — Human Rebel Ally", avB30aLongshotOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, longshot, "reach") {
		t.Error("Longshot has no reach")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.CastFromHandForTest(t, g, me, "Two Drop", "Sorcery", "{1}{R}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 2
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s: life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestAvatarB30LongshotIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Longshot, Rebel Bowman", "Legendary Creature — Human Rebel Ally", avB30aLongshotOracle, 3, 3)
	effectstest.FillPoolColored(me, "R", 1)
	if _, err := effectstest.CastFromHandErr(g, me, "Two Drop Bear", "Creature — Bear", "{1}{R}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was discounted")
	}
	life := opp.Life
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.CastFromHandForTest(t, g, me, "Two Drop Bear", "Creature — Bear", "{1}{R}", "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a creature spell dealt damage: life %d, want %d", opp.Life, life)
	}
}

// --- Lost in Memories ------------------------------------------------------------

func TestAvatarB30LostInMemoriesPumpsAndGrantsFlashbackOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	bolt := effectstest.SeedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-lost-in-memories-bolt")
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Lost in Memories", effectstest.AuraTypeLine, avB30aLostOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 3 || th != 3 {
		t.Fatalf("the enchanted creature is %d/%d, want 3/3", p, th)
	}
	if effectstest.GrantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Fatal("flashback before any combat damage")
	}
	effectstest.TrekB18AAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	perm := effectstest.GrantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || perm.Cost != "" || !perm.ExileOnResolution {
		t.Fatalf("granted permission %+v, want flashback at its mana cost", perm)
	}
}

func TestAvatarB30LostInMemoriesRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := avB30aOpp(g)
	theirs := effectstest.AuraBear(g, opp.ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Lost in Memories", effectstest.AuraTypeLine, avB30aLostOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("Lost in Memories enchanted an opponent's creature")
	}
}

// --- Moku, Meandering Drummer ----------------------------------------------------

func TestAvatarB30MokuPaidPumpsItselfAndHastesYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	effectstest.AdvanceToMain(t, g)
	moku := effectstest.B12Push(g, me.ID, "Moku, Meandering Drummer", "Legendary Creature — Human Bard Ally", avB30aMokuOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, moku); p != 4 || th != 3 {
		t.Errorf("Moku is %d/%d, want 4/3", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("Bear is %d/%d, want 2/2 (only Moku is pumped)", p, th)
	}
	for _, id := range []uuid.UUID{moku, bear} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", id)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature gained haste")
	}
}

func TestAvatarB30MokuDeclinedDoesNothingAndCreatureSpellsDontTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	moku := effectstest.B12Push(g, me.ID, "Moku, Meandering Drummer", "Legendary Creature — Human Bard Ally", avB30aMokuOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature spell triggered Moku")
	}
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, moku); p != 2 || th != 2 {
		t.Errorf("Moku is %d/%d after declining, want 2/2", p, th)
	}
	if effectstest.HasEffectiveKeyword(t, g, moku, "haste") {
		t.Error("Moku gained haste without the payment")
	}
}
