package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_07_restore_test.go — #4253's closure sweep, batch 07 (The Gold
// Saucer through Zukos Exile): the prompts that carried a Go closure are
// keyed now, so a table waiting on any of them is a restore point. Each test
// opens the prompt, takes a STRICT restore through JSON (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) and answers on
// the restored table.

// sweep07PickKey returns the open choose_cards prompt of `chooser`, failing
// the test when it names a key other than `want`.
func sweep07PickKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestChooseCardsFor(g, chooser)
	if c == nil {
		t.Fatalf("no choose-cards prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// Tireless Provisioner: the Food-or-Treasure question is a keyed option
// pick, so the restored table makes the Food.
func TestSweep07TirelessProvisionerChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Tireless Provisioner", "Creature — Elf Scout", tirelessProvisionerOracle, false)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if c := latestOptionPickFor(g, me.ID); c == nil || c.PickThen != "option-pick/tireless-provisioner-token" {
		t.Fatalf("the landfall choice is %+v, want the keyed option pick", c)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerOptionPick(t, g, me.ID, 1)
	if countBattlefieldNamed(g, me.ID, "Food") != 1 || countBattlefieldNamed(g, me.ID, "Treasure") != 0 {
		t.Error("choosing Food on the restored table did not make a Food")
	}
}

// Transpose: the discard is a keyed discard prompt, and whether the spell was
// cast from the hand rides it as data, so the Wizard comes only from the hand
// cast, on the restored table.
func TestSweep07TransposeDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, id := castReboundSpellAtMain(t, g, "Transpose", "Instant", transposeOracle)
	meID := me.ID
	life := me.Life
	passPriorityAroundTable(t, g)
	sweep07PickKey(t, g, meID, "choose-cards/transpose-after-discard")

	g = restoreRoundTrip(t, g, true)
	sweep07PickKey(t, g, meID, "choose-cards/transpose-after-discard")
	discardFromHand(t, g, meID)
	me = g.PlayerByIDForEffect(meID)
	if me.Life != life-1 {
		t.Errorf("life %d on the restored table, want %d", me.Life, life-1)
	}
	if n := onBattlefieldNamed(g, "Wizard"); n != 1 {
		t.Errorf("Wizards after the restored hand cast %d, want 1", n)
	}
	if !g.Exile.Contains(id) {
		t.Error("Transpose was not exiled by rebound on the restored table")
	}
}

// Undercity Plague: the discard is a keyed discard prompt and the sacrifice
// follows it on the restored table.
func TestSweep07UndercityPlagueDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	oppID := opp.ID
	ogre := pushVanillaCreature(g, oppID, "Their Ogre", 3, 3)
	castCatalogSpell(t, g, "Undercity Plague", "Sorcery", cibUndercityPlagueOracle, cibPlayer(oppID))
	passPriorityAroundTable(t, g)
	sweep07PickKey(t, g, oppID, "choose-cards/undercity-plague-sacrifice")

	g = restoreRoundTrip(t, g, true)
	sweep07PickKey(t, g, oppID, "choose-cards/undercity-plague-sacrifice")
	discardFromHand(t, g, oppID)
	answerSacrifice(t, g, oppID, ogre)
	g.SettleResolution()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the player did not sacrifice a permanent on the restored table")
	}
}

// Ultimate Magic: Meteor, foretold: the walk over the opponents is data. The
// table is restored after each pick, with the picks so far and the opponents
// still to ask on the prompt, and both chosen permanents are destroyed
// together after the last pick.
func TestSweep07UltimateMagicMeteorWalkIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	meID := me.ID
	opp1, opp2 := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4]
	rock1 := b12Permanent(g, opp1.ID, "Rock One", "Artifact")
	spare1 := b12Permanent(g, opp1.ID, "Spare One", "Artifact")
	rock2 := b12Permanent(g, opp2.ID, "Rock Two", "Artifact")

	pay := append(cloudDeckMana("C", 5), game.ManaToken{Color: "R"})
	cloudDeckForetellAndCast(t, g, "Ultimate Magic: Meteor", "Sorcery", ultimateMeteorOracle, "{5}{R}", pay, nil)
	passPriorityAroundTable(t, g)
	first := sweep07PickKey(t, g, meID, "choose-cards/ultimate-magic-meteor-walk")
	if !hasID(first.ChooseCards, rock1) || !hasID(first.ChooseCards, spare1) {
		t.Fatalf("the first prompt offers %v, want the first opponent's artifacts", first.ChooseCards)
	}

	g = restoreRoundTrip(t, g, true)
	first = sweep07PickKey(t, g, meID, "choose-cards/ultimate-magic-meteor-walk")
	if err := g.ResolveChooseCards(first.ID, meID, []uuid.UUID{rock1}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !g.Battlefield.Contains(rock1) {
		t.Fatal("nothing is destroyed until the last opponent's pick is in")
	}

	g = restoreRoundTrip(t, g, true)
	second := sweep07PickKey(t, g, meID, "choose-cards/ultimate-magic-meteor-walk")
	if !hasID(second.ChooseCards, rock2) || hasID(second.ChooseCards, spare1) {
		t.Fatalf("the second prompt offers %v, want only the second opponent's artifact", second.ChooseCards)
	}
	if err := g.ResolveChooseCards(second.ID, meID, []uuid.UUID{rock2}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock1) || g.Battlefield.Contains(rock2) {
		t.Error("the chosen artifacts were not both destroyed on the restored table")
	}
	if !g.Battlefield.Contains(spare1) {
		t.Error("an unchosen artifact was destroyed")
	}
}

// UntapUpToLands with nothing printed after it (Snap, Frantic Search,
// Palinchron, Teferi, Hero of Dominaria, and the rest) shares one keyed pick.
func TestSweep07UntapUpToLandsIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meID := me.ID
	tappedForest := func(name string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — Forest",
			Owner: meID, Controller: meID, Tapped: true,
		})
	}
	a, b, c := tappedForest("Tapped Forest A"), tappedForest("Tapped Forest B"), tappedForest("Tapped Forest C")
	g.WithWriteLock(func() {
		item := &game.StackItem{Kind: game.StackItemSpell, Controller: meID, Owner: meID}
		if err := (UntapUpToLands{N: 2, Question: "Test — untap up to two lands"}).Apply(NewContext(g, item)); err != nil {
			t.Fatal(err)
		}
	})
	sweep07PickKey(t, g, meID, "choose-cards/untap-up-to-lands")

	g = restoreRoundTrip(t, g, true)
	pick := sweep07PickKey(t, g, meID, "choose-cards/untap-up-to-lands")
	if len(pick.ChooseCards) != 3 || pick.ChooseMax != 2 {
		t.Errorf("the restored prompt offers %d lands up to %d, want 3 up to 2", len(pick.ChooseCards), pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, meID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if b16Tapped(t, g, a) || b16Tapped(t, g, b) || !b16Tapped(t, g, c) {
		t.Error("exactly the two chosen lands should be untapped on the restored table")
	}
}

// Zack Fair: which Equipment moves is a keyed pick carrying the target
// creature, and it attaches on the restored table.
func TestSweep07ZackFairEquipmentPickIsARestorePoint(t *testing.T) {
	g, zack, bear, me := zackFairBoard(t)
	first := seedEquipment(g, me, "First Blade", "")
	second := seedEquipment(g, me, "Second Blade", "")
	for _, eq := range []uuid.UUID{first, second} {
		if err := cloudDeckAttach(g, eq, game.TargetRef{Kind: game.TargetCard, ID: zack}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	}
	zackFairActivate(t, g, zack, bear)
	passPriorityAroundTable(t, g)
	sweep07PickKey(t, g, me, "choose-cards/zack-fair-move-equipment")

	g = restoreRoundTrip(t, g, true)
	pick := sweep07PickKey(t, g, me, "choose-cards/zack-fair-move-equipment")
	if len(pick.ChooseCards) != 2 {
		t.Errorf("the restored prompt offers %d Equipment, want 2", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if host := attachmentHostOf(t, g, second); host.ID != bear {
		t.Errorf("the chosen Equipment is attached to %+v, want the Bear", host)
	}
}

// Web of Inertia: the opponent's exile-or-restrict question is a keyed pick
// carrying the opponent and the trigger's item; declining on the restored
// table still shuts the attack.
func TestSweep07WebOfInertiaPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppID := opp.ID
	pushPermanentForTest(g, me.ID, "Web of Inertia", webOfInertiaOracle, "Enchantment")
	raider := b12Creature(g, oppID, "Raider", "Creature — Test", 2, 2)
	pushGraveyardCardTyped(opp, "Dead Card", "Sorcery")

	advanceToStepOf(t, g, 1, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	sweep07PickKey(t, g, oppID, "choose-cards/web-of-inertia-exile-or-restrict")

	g = restoreRoundTrip(t, g, true)
	pick := sweep07PickKey(t, g, oppID, "choose-cards/web-of-inertia-exile-or-restrict")
	if err := g.ResolveChooseCards(pick.ID, oppID, nil); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	var pe *game.PlayerCantAttackError
	if err := g.Clone().DeclareAttacker(raider, g.Seats[0].ID); !errors.As(err, &pe) {
		t.Errorf("attack at the Web's controller after declining: err = %v, want *PlayerCantAttackError", err)
	}
}

// Valakut Awakening: the any-number put-back is a keyed pick with nothing
// carried, and nothing put back still draws one card on the restored table.
func TestSweep07ValakutAwakeningPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meID := me.ID
	seedLibrary(me, "Library Card")
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Valakut Awakening", "Instant", valakutAwakeningOracle, nil)
	passPriorityAroundTable(t, g)
	sweep07PickKey(t, g, meID, "choose-cards/valakut-awakening-put-back")

	g = restoreRoundTrip(t, g, true)
	pick := sweep07PickKey(t, g, meID, "choose-cards/valakut-awakening-put-back")
	if err := g.ResolveChooseCards(pick.ID, meID, nil); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := g.PlayerByIDForEffect(meID).Hand.Size(); got != handBefore+1 {
		t.Errorf("hand size %d on the restored table, want %d (draw 1, nothing put back)", got, handBefore+1)
	}
}
