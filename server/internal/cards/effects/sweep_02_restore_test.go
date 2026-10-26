package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_02_restore_test.go — #4253's closure sweep, batch 02 (endemic_plague
// through gluntch_the_bestower): the prompts that carried a Go closure are
// keyed now, so a table waiting on any of them is a restore point. Each test
// opens the prompt, takes a STRICT restore through JSON (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) and answers on
// the restored table.

// sweep02OptionKey returns the open option pick of `chooser`, failing the
// test when it names a key other than `want`.
func sweep02OptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// sweep02PickKey returns the open choose_cards prompt of `chooser`, failing
// the test when it names a key other than `want`.
func sweep02PickKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
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

// "Your choice of a +1/+1, first strike or trample counter" is one shared
// registration (Assaultron Dominator, T-45 Power Armor, Ichormoon Gauntlet).
// The permanent and the placer ride the prompt, and the kind is read off the
// option that was chosen, so the restored pick lands the counter chosen.
func TestSweep02ChosenCounterIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dom := pushDiesCreatureForTest(g, me.ID, "Assaultron Dominator", "2c7818f4-a818-4603-9af1-05aee2d56ec8", "Artifact Creature — Robot", 2, 2)
	setEnergy(t, g, me, 1)
	declareAttack(t, g, opp.ID, dom)
	for i := 0; i < 20 && latestOptionPickFor(g, me.ID) == nil; i++ {
		passPriorityAroundTable(t, g)
		if p := energyPrompt(g, me.ID); p != nil {
			if err := g.ResolvePayUnless(p.ID, me.ID, true); err != nil {
				t.Fatalf("pay the energy: %v", err)
			}
		}
	}
	sweep02OptionKey(t, g, me.ID, energyChosenCounter.Key())

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep02OptionKey(t, g, me.ID, energyChosenCounter.Key())
	if len(c.PickOptions) != 3 || c.PickOptions[2].Label != counterOptionLabel(game.CounterTrample) {
		t.Fatalf("the restored pick offers %+v, want the three counters with trample last", c.PickOptions)
	}
	answerOptionPick(t, g, me.ID, 2)
	if got := apaLive(g, dom).Counters[game.CounterTrample]; got != 1 {
		t.Errorf("trample counters = %d on the restored table, want 1", got)
	}
	if got := apaLive(g, dom).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("+1/+1 counters = %d, want none: the kind chosen is the one placed", got)
	}
}

// The kind is read back off the label, so the round trip has to be exact for
// every counter the cards offer.
func TestSweep02CounterKindRoundTripsThroughItsLabel(t *testing.T) {
	for _, kind := range []string{
		game.CounterPlusOne, game.CounterFirstStrike, game.CounterTrample,
		game.CounterMenace, game.CounterLifelink,
	} {
		if got := counterKindOfOptionLabel(counterOptionLabel(kind)); got != kind {
			t.Errorf("kind %q came back as %q", kind, got)
		}
	}
}

// Fandaniel: each opponent's sacrifice-or-lose-life question is a keyed option
// pick that carries the opponents still to be asked, so the table is a restore
// point at it. The restored answer asks the next opponent after it.
func TestSweep02FandanielOpponentsAreARestorePoint(t *testing.T) {
	g := fandanielGame(t, 2)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	theirs := pushVanillaCreature(g, g.Seats[2].ID, "Third Bear", 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	const optionKey = "option-pick/fandaniel-telophoroi-ascian-sacrifice-or-lose-life"
	c := sweep02OptionKey(t, g, opp.ID, optionKey)
	// The asking controller rides first, then the opponents still to be asked.
	if len(c.OptionCarry) != 3 {
		t.Fatalf("the first question carries %d IDs, want the controller and the other two opponents", len(c.OptionCarry))
	}

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	sweep02OptionKey(t, g, opp.ID, optionKey)
	answerOptionPick(t, g, opp.ID, 0) // lose life instead

	// The next opponent is asked on the restored table, with its own question.
	third := g.Seats[2]
	sweep02OptionKey(t, g, third.ID, optionKey)
	g = restoreRoundTrip(t, g, true)
	third = g.Seats[2]
	answerOptionPick(t, g, third.ID, 0)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(theirs) {
		t.Error("declining kept both creatures on the restored table")
	}
	for _, i := range []int{1, 2} {
		if got := g.Seats[i].Life; got != 36 {
			t.Errorf("seat %d life = %d on the restored table, want 36 (2 x 2)", i, got)
		}
	}
	if got := g.Seats[3].Life; got != 36 {
		t.Errorf("seat 3 had nothing to sacrifice and loses 2 x 2: life = %d, want 36", got)
	}
}

// Finale of Revelation: "untap up to five lands" is a keyed pick that carries
// the spell as data, so the restored pick untaps the lands, grants no maximum
// hand size and exiles the spell.
func TestSweep02FinaleOfRevelationUntapIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	var lands []uuid.UUID
	for i := 0; i < 6; i++ {
		landID := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: landID, Name: "Island", TypeLine: "Basic Land — Island",
			Owner: caster.ID, Controller: caster.ID, Tapped: true,
		})
		lands = append(lands, landID)
	}
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Finale of Revelation", TypeLine: "Sorcery",
		OracleID: finaleOfRevelationOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{XValue: 10}); err != nil {
		t.Fatalf("CastSpell Finale of Revelation: %v", err)
	}
	passPriorityAroundTable(t, g)
	const key = "choose-cards/finale-of-revelation-untap-lands"
	sweep02PickKey(t, g, caster.ID, key)

	g = restoreRoundTrip(t, g, true)
	caster = g.Seats[0]
	c := sweep02PickKey(t, g, caster.ID, key)
	if c.ChooseMax != 5 || len(c.ChooseCards) != 6 {
		t.Errorf("the restored pick is %d..%d over %d lands, want up to 5 over the 6 tapped", c.ChooseMin, c.ChooseMax, len(c.ChooseCards))
	}
	answerChooseCards(t, g, caster.ID, lands[:5]...)
	untapped := 0
	for _, card := range g.Battlefield.Cards {
		for _, lid := range lands {
			if card.InstanceID == lid && !card.Tapped {
				untapped++
			}
		}
	}
	if untapped != 5 {
		t.Errorf("untapped %d of the lands on the restored table, want 5", untapped)
	}
	if caster.MaxHandSize != game.NoMaxHandSize {
		t.Error("X>=10 should grant no maximum hand size on the restored table")
	}
	if !g.Exile.Contains(id) {
		t.Error("Finale of Revelation should exile itself on the restored table")
	}
}

// Extract from Darkness (and Fortune, Loyal Steed, which shares the helper):
// the pick carries the asking spell as data, so the restored pick puts the
// chosen creature card onto the battlefield under the caster's control.
func TestSweep02ExtractFromDarknessPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b17GraveyardCard(opp, "Their Dead Golem", "Creature — Golem", "{4}")
	castCatalogSpell(t, g, "Extract from Darkness", "Sorcery", b30ExtractFromDarknessOracle, nil)
	passPriorityAroundTable(t, g)
	const key = "choose-cards/extract-from-darkness-reanimate"
	sweep02PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep02PickKey(t, g, me.ID, key)
	if !hasID(c.ChooseCards, dead) {
		t.Fatal("the restored pick does not offer the creature card in the graveyard")
	}
	answerChooseCards(t, g, me.ID, dead)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the Golem is not on the battlefield on the restored table")
	}
	if card := b12Card(t, g, dead); card.Controller != me.ID {
		t.Error("the Golem is not under the caster's control on the restored table")
	}
}

// Gilgamesh, Master-at-Arms: the Equipment pick and the Samurai it goes on are
// keyed picks, the creature type riding the prompt as its word, so a table
// waiting on either is a restore point and the Equipment attaches.
func TestSweep02GilgameshAttachIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sword := plTop(me, "Sword", "Artifact — Equipment", "{3}")
	gil := castCatalogSpell(t, g, "Gilgamesh, Master-at-Arms", "Legendary Creature — Human Samurai", plGilgameshOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	const equipmentKey = "choose-cards/gilgamesh-master-at-arms-equipment"
	sweep02PickKey(t, g, me.ID, equipmentKey)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep02PickKey(t, g, me.ID, equipmentKey)
	if c.ChooseMin != 0 || !hasID(c.ChooseCards, sword) {
		t.Fatalf("the restored pick is %d.. over %v, want an optional pick over the Sword", c.ChooseMin, c.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, sword)
	const hostKey = "choose-cards/gilgamesh-master-at-arms-samurai"
	sweep02PickKey(t, g, me.ID, hostKey)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c = sweep02PickKey(t, g, me.ID, hostKey)
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != gil {
		t.Fatalf("the restored host pick offers %v, want Gilgamesh alone", c.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, gil)
	card, _ := g.LookupCardForEffect(sword)
	if !card.IsAttachedTo(gil) {
		t.Error("the Equipment was not attached to Gilgamesh on the restored table")
	}
}

// Sol'Kanar the Tainted (and Zuko, Conflicted): naming the opponent it comes
// back under is a keyed choose-player prompt that carries the permanent, so
// the restored answer returns it under the chosen opponent as a new object.
func TestSweep02ChooseAnOpponentIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	b12Push(g, me.ID, "Sol'Kanar the Tainted", "Legendary Creature — Elemental Demon",
		"4efcdefc-e49d-4bce-8581-69037bb48c0a", 5, 5)

	advanceToEndStepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 3)
	passPriorityAroundTable(t, g)
	const key = "option-pick/flicker-return-under-an-opponent"
	sweep02OptionKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	sweep02OptionKey(t, g, me.ID, key)
	answerChoosePlayer(t, g, me.ID, opp)
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, findBattlefieldByName(g, "Sol'Kanar the Tainted"))
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Sol'Kanar does not return under the chosen opponent's control on the restored table: %+v", c)
	}
}

// Riddles in the Dark: "choose an opponent to choose one of the piles" is a
// keyed choose-player prompt that carries the cards looked at, so the restored
// answer opens the split over the same four cards.
func TestSweep02RiddlesInTheDarkOpponentIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedFactOrFictionLibrary(me, "Rd One", "Rd Two", "Rd Three", "Rd Four")
	castCatalogSpell(t, g, "Riddles in the Dark", "Instant", riddlesOracle, nil)
	passPriorityAroundTable(t, g)
	const key = "option-pick/face-down-piles-riddles-in-the-dark"
	sweep02OptionKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sweep02OptionKey(t, g, me.ID, key)
	answerChoosePlayer(t, g, me.ID, opp)
	split := latestRevealPickFor(g, me.ID)
	if split == nil || len(split.ChooseCards) != 4 {
		t.Fatalf("the caster separates the four they looked at on the restored table: %+v", split)
	}
}

// Ethereal Forager: the pick of the exiled instant or sorcery to return is a
// keyed pick, so the restored answer returns it to its owner's hand.
func TestSweep02EtherealForagerPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	instant := game.Card{InstanceID: uuid.New(), Name: "Opt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID}
	forager := pushLinkedForager(t, g, me.ID, instant)
	declareAttack(t, g, opp.ID, forager)
	passPriorityAroundTable(t, g)
	const key = "choose-cards/ethereal-forager-return"
	sweep02PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep02PickKey(t, g, me.ID, key)
	if c.ChooseMin != 0 || c.ChooseMax != 1 || !hasID(c.ChooseCards, instant.InstanceID) {
		t.Fatalf("the restored pick is %d..%d over %v, want an optional pick of the linked instant", c.ChooseMin, c.ChooseMax, c.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, instant.InstanceID)
	if z := g.FindCardZoneForEffect(instant.InstanceID); z == nil || z.Kind != game.ZoneHand {
		t.Error("the chosen instant is not in its owner's hand on the restored table")
	}
}
