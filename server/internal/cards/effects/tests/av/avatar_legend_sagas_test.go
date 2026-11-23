package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func kurukRow() cards.Card {
	return effectstest.AvatarSagaRow(effects.TheLegendOfKurukOracleID, "The Legend of Kuruk", "{2}{U}{U}", "Avatar Kuruk", "4", "3", "U")
}

func rokuRow() cards.Card {
	return effectstest.AvatarSagaRow(effects.TheLegendOfRokuOracleID, "The Legend of Roku", "{2}{R}{R}", "Avatar Roku", "4", "4", "R")
}

// assertReturnedAsAvatar is chapter III's card-level proof, common to
// all four Sagas: the original object is gone (not merely
// transformed in place, and not sacrificed to CR 714.4 either — a
// Saga that has already left when the SBA next runs), a NEW object
// with the Avatar's name/type/P-T has arrived under the controller's
// control, with no lore counters, and it is summoning sick because it
// just entered (CR 400.7). Mirrors
// TestFableChapterThreeExilesAndReturnsItTransformed's assertion set.
func assertReturnedAsAvatar(t *testing.T, g *game.Game, saga, controller uuid.UUID, avatarName string, power, toughness int) *game.Card {
	t.Helper()
	if effectstest.BattlefieldCardFor(g, saga) != nil {
		t.Fatal("the original Saga object is still on the battlefield — chapter III is an exile and a return, not an in-place transform")
	}
	if effectstest.InGraveyardOf(g, controller, saga) {
		t.Fatal("the Saga was sacrificed (CR 714.4) instead of coming back")
	}
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == avatarName {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatalf("%s never reached the battlefield", avatarName)
	}
	if back.InstanceID == saga {
		t.Error("it kept its InstanceID — CR 400.7 makes the returning permanent a new object")
	}
	if back.Controller != controller {
		t.Errorf("controller = %s, want %s (under your control)", back.Controller, controller)
	}
	if back.ActiveFace != 1 {
		t.Errorf("ActiveFace = %d, want the back face", back.ActiveFace)
	}
	if back.Counters[game.CounterLore] != 0 {
		t.Errorf("it came back with %d lore counters", back.Counters[game.CounterLore])
	}
	if !back.IsCreature() || back.Power != power || back.Toughness != toughness {
		t.Errorf("came back as %q %d/%d, want a %d/%d creature",
			back.TypeLine, back.Power, back.Toughness, power, toughness)
	}
	if !back.SummonedThisTurn {
		t.Error("it is not summoning sick — the returning permanent entered this turn")
	}
	return back
}

// --- The Legend of Kuruk // Avatar Kuruk ---------------------------

// TestKurukEntersAndChapterOneScriesThenDraws is the entry + chapter
// I proof: one lore counter from CR 714.3, then "scry 2, then draw a
// card" (Preordain's own body, chaptered).
func TestKurukEntersAndChapterOneScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	handBefore := me.Hand.Size()

	saga := effectstest.ImportAndCast(t, g, kurukRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Fatalf("hand delta after chapter I = %d, want 1 (scry 2, then draw a card)", got)
	}
}

// TestKurukChapterThreeReturnsAsAvatarKuruk walks all three chapters
// (answering the identical chapter II scry along the way) and checks
// the transform.
func TestKurukChapterThreeReturnsAsAvatarKuruk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := effectstest.ImportAndCast(t, g, kurukRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Kuruk", 4, 3)
	effectstest.AssertFaceInvariant(t, g)
}

// TestKurukChapterThreeCopiedReturnsItOnce is #2497, found by the
// catalog soak: Gogo, Master of Mimicry copied chapter III. A copy
// refers to the same Saga as the original (CR 707.10b), so whichever
// resolves first exiles it and returns it as Avatar Kuruk, a new
// object (CR 400.7), and the other finds nothing to exile and does
// nothing. One Avatar, and no effect error.
func TestKurukChapterThreeCopiedReturnsItOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := effectstest.ImportAndCast(t, g, kurukRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)

	chapter := effectstest.AcItemOnStack(g, "The Legend of Kuruk — exile it")
	if chapter == nil {
		t.Fatal("chapter III is not on the stack")
	}
	var err error
	g.WithWriteLock(func() { err = g.CopyAbilityForEffect(chapter.ID, me.ID, false) })
	if err != nil {
		t.Fatalf("copy chapter III: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if errs := effectstest.EffectErrors(g); len(errs) != 0 {
		t.Fatalf("effect errors: %v — the copy that resolved second must do nothing", errs)
	}
	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Kuruk", 4, 3)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Avatar Kuruk" {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("%d Avatar Kuruks on the battlefield, want 1", n)
	}
	effectstest.AssertFaceInvariant(t, g)
}

// TestAvatarKurukCreatesASpiritWhenYouCastASpell is the back-face
// ability: "Whenever you cast a spell, create a 1/1 colorless
// Spirit." The Spirit's own "can't block or be blocked by non-Spirit
// creatures" clause is the declared caveat (#521) — not asserted
// here, because it does not exist to assert.
func TestAvatarKurukCreatesASpiritWhenYouCastASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PutBackFaceOntoBattlefield(t, g, kurukRow(), me)

	before := effectstest.CountBattlefieldByName(g, "Spirit")
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldByName(g, "Spirit") - before; got != 1 {
		t.Fatalf("Spirits created = %d, want 1", got)
	}
}

// --- The Legend of Kyoshi // Avatar Kyoshi --------------------------

// TestKyoshiChapterOneDrawsEqualToGreatestPower is the entry +
// chapter I proof.
func TestKyoshiChapterOneDrawsEqualToGreatestPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
		Owner: me.ID, Controller: me.ID, Power: 5, Toughness: 5,
	})
	handBefore := me.Hand.Size()

	saga := effectstest.ImportAndCast(t, g, effectstest.KyoshiRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if got := me.Hand.Size() - handBefore; got != 5 {
		t.Fatalf("hand delta after chapter I = %d, want 5 (the greatest power on the board)", got)
	}
}

// TestKyoshiChapterThreeReturnsAsAvatarKyoshi walks all three
// chapters. Chapter II fizzles (no land in play to target — CR
// 603.3d, zero legal targets), which is fine: this test is chapter
// III's, not chapter II's.
func TestKyoshiChapterThreeReturnsAsAvatarKyoshi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := effectstest.ImportAndCast(t, g, effectstest.KyoshiRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Kyoshi", 5, 4)
	effectstest.AssertFaceInvariant(t, g)
}

// TestAvatarKyoshiManaAbilityAddsGreatestPowerInOneColor is the
// back-face ability: "{T}: Add X mana of any one color, X = the
// greatest power among creatures you control" — Selvala's shape with
// one colour instead of a combination.
func TestAvatarKyoshiManaAbilityAddsGreatestPowerInOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	entered := effectstest.PutBackFaceOntoBattlefield(t, g, effectstest.KyoshiRow(), me)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast",
		Owner: me.ID, Controller: me.ID, Power: 5, Toughness: 5,
	})

	if err := g.ActivateManaAbility(me.ID, entered, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	color := effectstest.ResolveManaPickAnyColor(t, g, me.ID)

	got := 0
	for _, tok := range me.ManaPool.Units() {
		if tok.Color == color {
			got++
		}
	}
	if got != 5 {
		t.Fatalf("pool has %d mana of %s, want 5 (the greatest power)", got, color)
	}
}

// --- The Legend of Roku // Avatar Roku ------------------------------

// TestRokuChapterOneExilesTopThreeWithPlayPermission is the entry +
// chapter I proof: Reckless Impulse's primitive with three cards
// instead of two.
func TestRokuChapterOneExilesTopThreeWithPlayPermission(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	top := make([]uuid.UUID, 3)
	for i := range top {
		top[i] = me.Library.Cards[me.Library.Size()-1-i].InstanceID
	}

	saga := effectstest.ImportAndCast(t, g, rokuRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	for _, id := range top {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s never reached exile", id)
			continue
		}
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm == nil {
			t.Errorf("exiled card %s has no play permission", id)
		}
	}
}

// TestRokuChapterThreeReturnsAsAvatarRoku walks all three chapters,
// answering chapter II's colour pick along the way.
func TestRokuChapterThreeReturnsAsAvatarRoku(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]

	saga := effectstest.ImportAndCast(t, g, rokuRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Roku", 4, 4)
	effectstest.AssertFaceInvariant(t, g)
}

// TestAvatarRokuCreatesADragonToken is the back-face ability:
// "{8}: Create a 4/4 red Dragon creature token with flying and
// firebending 4." Permissive mode waives the unfunded mana cost
// (docs/adding-cards.md's mana section — the human default), so the test does
// not need to float eight mana first.
func TestAvatarRokuCreatesADragonToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	entered := effectstest.PutBackFaceOntoBattlefield(t, g, rokuRow(), me)

	before := effectstest.CountBattlefieldByName(g, "Dragon")
	if err := g.ActivateCatalogAbility(me.ID, entered, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldByName(g, "Dragon") - before; got != 1 {
		t.Fatalf("Dragon tokens created = %d, want 1", got)
	}
}

// --- The Legend of Yangchen // Avatar Yangchen ----------------------

// TestYangchenChapterOneEveryPlayerChoosesFromTheSharedPool is the entry +
// chapter I proof: starting with the controller, each player chooses up to
// one permanent with mana value 3 or greater from among permanents the
// controller's opponents control, each answer narrows the next player's
// pool, and the picks are exiled together after the last answer.
func TestYangchenChapterOneEveryPlayerChoosesFromTheSharedPool(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	opp2 := g.Seats[(seat+2)%len(g.Seats)]
	push := func(owner uuid.UUID, name, cost string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Artifact",
			ManaCost: cost, Owner: owner, Controller: owner,
		})
	}
	rock := push(opp.ID, "Their Big Rock", "{4}")
	statue := push(opp2.ID, "Their Statue", "{5}")
	cheap := push(opp.ID, "Their Cheap Rock", "{2}")
	mine := push(me.ID, "My Own Rock", "{4}")

	saga := effectstest.ImportAndCast(t, g, effectstest.YangchenRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	ask := func(who uuid.UUID) *game.PendingChoice {
		t.Helper()
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Chooser == who && (pc.Kind == game.PendingChoiceOwnPermanents || pc.Kind == game.PendingChoiceTheirPermanents) {
				c = pc
			}
		}
		if c == nil {
			t.Fatalf("no permanent prompt for %s: %+v", who, g.PendingChoices)
		}
		return c
	}
	answer := func(c *game.PendingChoice, picks ...uuid.UUID) {
		t.Helper()
		var err error
		if c.Kind == game.PendingChoiceOwnPermanents {
			err = g.ResolveOwnPermanents(c.ID, c.Chooser, picks)
		} else {
			err = g.ResolveTheirPermanents(c.ID, c.Chooser, picks)
		}
		if err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	offers := func(c *game.PendingChoice, id uuid.UUID) bool {
		for _, cand := range c.ChooseCards {
			if cand == id {
				return true
			}
		}
		return false
	}

	first := ask(me.ID)
	if !offers(first, rock) || !offers(first, statue) || offers(first, cheap) || offers(first, mine) {
		t.Fatalf("the first pool is the opponents' permanents with mana value 3 or more: %v", first.ChooseCards)
	}
	if first.ChooseMin != 0 || first.ChooseMax != 1 {
		t.Errorf("up to one: %d..%d", first.ChooseMin, first.ChooseMax)
	}
	answer(first, rock)

	second := ask(opp.ID)
	if offers(second, rock) || !offers(second, statue) {
		t.Fatalf("the second pool lost what the first player took: %v", second.ChooseCards)
	}
	answer(second)

	// Players with nothing to take still get their say; nobody else gets
	// the prompt twice.
	for len(g.PendingChoices) > 0 {
		c := g.PendingChoices[0]
		if c == nil || (c.Kind != game.PendingChoiceOwnPermanents && c.Kind != game.PendingChoiceTheirPermanents) {
			break
		}
		if c.Chooser == opp2.ID {
			answer(c, statue)
			continue
		}
		answer(c)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(statue) {
		t.Error("chapter I did not exile both chosen permanents")
	}
	if !g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(mine) {
		t.Error("an ineligible permanent was exiled")
	}
}

// TestYangchenChapterThreeReturnsAsAvatarYangchen walks all three
// chapters, picking chapter II's target opponent and declining the
// "you may" so the assertion stays about chapter III.
func TestYangchenChapterThreeReturnsAsAvatarYangchen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]

	saga := effectstest.ImportAndCast(t, g, effectstest.YangchenRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if got := effectstest.LoreCountersOn(g, saga); got != 2 {
		t.Fatalf("lore counters after chapter II = %d, want 2", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)

	assertReturnedAsAvatar(t, g, saga, me.ID, "Avatar Yangchen", 4, 5)
	effectstest.AssertFaceInvariant(t, g)
}

// TestAvatarYangchenAirbendsOnSecondSpellEachTurn is the back-face
// ability: "Whenever you cast your second spell each turn, airbend up
// to one other target nonland permanent."
func TestAvatarYangchenAirbendsOnSecondSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")
	effectstest.PutBackFaceOntoBattlefield(t, g, effectstest.YangchenRow(), me)

	effectstest.CastCatalogSpell(t, g, "First Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the first spell each turn does not trigger: %+v", g.PendingChoices)
	}

	effectstest.CastCatalogSpell(t, g, "Second Spell", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, victim, opp.ID)
}
