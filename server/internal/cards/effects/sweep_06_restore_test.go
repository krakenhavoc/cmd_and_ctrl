package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_06_restore_test.go — #4253's closure sweep, batch 06 (selesnya_eulogist
// through the_gitrog_ravenous_ride): the prompts that carried a Go closure are
// keyed now, so a table waiting on any of them is a restore point. Each test
// opens the prompt, takes a STRICT restore through JSON (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) and answers on
// the restored table.

// sweep06OptionKey returns the open option pick of `chooser`, failing the test
// when it names a key other than `want`.
func sweep06OptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the option pick names key %q, want %q", c.PickThen, want)
	}
	return c
}

// sweep06PickKey returns the open choose_cards prompt of `chooser`, failing the
// test when it names a key other than `want`.
func sweep06PickKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
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

// Mind Bend (the key the other six text-changing cards share): both prompts
// are restore points, and the target class, the target and the word replaced
// ride the second prompt as data.
func TestSweep06ChangeTextWordIsARestorePointAtBothPrompts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wraith := twPut(g, me.ID, "Bog Wraith", "Creature — Wraith", []string{"B"}, "swampwalk")
	castCatalogSpell(t, g, "Mind Bend", "Instant", twMindBendOracle, twTargetRef(wraith))
	passPriorityAroundTable(t, g)
	sweep06OptionKey(t, g, me.ID, "option-pick/change-text-word-replace")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	first := sweep06OptionKey(t, g, me.ID, "option-pick/change-text-word-replace")
	if len(first.PickOptions) != 10 {
		t.Fatalf("the restored first prompt offers %d words, want the five colours and five land types", len(first.PickOptions))
	}
	answerOptionPick(t, g, me.ID, 7) // Swamp
	second := sweep06OptionKey(t, g, me.ID, "option-pick/change-text-word-with")
	var labels []string
	for _, o := range second.PickOptions {
		labels = append(labels, o.Label)
	}
	if !slices.Equal(labels, []string{"Plains", "Island", "Mountain", "Forest"}) {
		t.Fatalf("the second prompt offers %v, want the other land types", labels)
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	sweep06OptionKey(t, g, me.ID, "option-pick/change-text-word-with")
	answerOptionPick(t, g, me.ID, 3) // Forest
	if !hasEffectiveKeyword(t, g, wraith, "forestwalk") {
		t.Error("swampwalk did not become forestwalk on the restored table")
	}
}

// Shah of Naar Isle: the opponents still to ask and the answers so far ride the
// prompt as data, so a restore between two opponents' answers loses neither.
func TestSweep06ShahChainIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	shah := seatEchoCard(g, me.ID, "Shah of Naar Isle", "Creature — Efreet")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if err := g.ResolvePayUnless(echoPrompt(g, me.ID).ID, me.ID, true); err != nil {
		t.Fatalf("pay echo {0}: %v", err)
	}
	if triggerOnStack(g, shah) == nil {
		t.Fatal("paying the echo did not trigger Shah of Naar Isle")
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	passPriorityAroundTable(t, g)

	// Turn order after seat 1: seats 2, 3, 0. Each picks 3, 0, 1; the table is
	// restored before every answer.
	for i, seat := range []int{2, 3, 0} {
		g = restoreRoundTrip(t, g, true)
		opp := g.Seats[seat]
		c := sweep06OptionKey(t, g, opp.ID, "option-pick/shah-of-naar-isle-draw")
		if err := g.ResolveOptionPick(c.ID, opp.ID, []int{3, 0, 1}[i]); err != nil {
			t.Fatalf("seat %d answers on the restored table: %v", seat, err)
		}
	}
	for seat, want := range map[int]int{2: 3, 3: 0, 0: 1, 1: 0} {
		p := g.Seats[seat]
		if got := p.Hand.Size() - before[p.ID]; got != want {
			t.Errorf("seat %d drew %d, want %d", seat, got, want)
		}
	}
}

// Spinal Parasite: the permanent is the carry and the kind removed is read off
// the option chosen.
func TestSweep06SpinalParasiteCounterChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	rock := b12Push(g, opp.ID, "Charged Rock", "Artifact", "", 0, 0)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(rock, game.CounterCharge, 3)
		_ = g.AddCounterForEffect(rock, game.CounterPlusOne, 1)
	})
	floatForTest(g, me, "WUBRG")
	parasite := castFromHandForTest(t, g, me, "Spinal Parasite", "Artifact Creature — Insect", "{5}", gmrSpinalParasiteOracle, strict)
	passPriorityAroundTable(t, g)
	b16Activate(t, g, me.ID, parasite, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})
	passPriorityAroundTable(t, g)
	c := sweep06OptionKey(t, g, me.ID, "option-pick/remove-a-counter-of-your-choice")
	if len(c.PickOptions) != 2 {
		t.Fatalf("the prompt offers %d kinds, want 2", len(c.PickOptions))
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	c = sweep06OptionKey(t, g, me.ID, "option-pick/remove-a-counter-of-your-choice")
	index := -1
	for i, o := range c.PickOptions {
		if o.Label == "Remove a charge counter" {
			index = i
		}
	}
	if index < 0 {
		t.Fatalf("no charge counter option on the restored prompt: %+v", c.PickOptions)
	}
	answerOptionPick(t, g, me.ID, index)
	if got := counterCount(g, rock, game.CounterCharge); got != 2 {
		t.Errorf("charge counters = %d on the restored table, want 2", got)
	}
	if got := counterCount(g, rock, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d on the restored table, want the 1 left alone", got)
	}
}

// Teferi Akosa, +1: "two cards, or a creature card" is a registered set rule,
// so the restored prompt still refuses a lone land and accepts a lone creature.
func TestSweep06TeferiAkosaDiscardKeepsItsSetRule(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	pw := pushTeferiAkosa(g, owner.ID, 4)
	advanceToMainOf(t, g, seat)
	for _, c := range append([]uuid.UUID(nil), handIDs(g, owner.ID)...) {
		if err := g.MoveCardByID(
			game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
			game.ZoneRef{Kind: game.ZoneLibrary, Owner: owner.ID}, c,
		); err != nil {
			t.Fatalf("clear the hand: %v", err)
		}
	}
	land := pushToHand(g, owner.ID, "Plains", "Basic Land — Plains")
	beast := pushToHand(g, owner.ID, "Bear", "Creature — Bear")
	if err := g.ActivateCatalogAbility(owner.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	sweep06PickKey(t, g, owner.ID, "choose-cards/teferi-akosa-discard-two-or-creature")

	g = restoreRoundTrip(t, g, true)
	owner = g.Seats[seat]
	prompt := sweep06PickKey(t, g, owner.ID, "choose-cards/teferi-akosa-discard-two-or-creature")
	if prompt.ChooseMin != 1 || prompt.ChooseMax != 2 {
		t.Errorf("restored discard bounds = [%d, %d], want [1, 2]", prompt.ChooseMin, prompt.ChooseMax)
	}
	if err := g.ResolveChooseCards(prompt.ID, owner.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("the restored prompt accepted a lone land: the set rule did not survive")
	}
	if err := g.ResolveChooseCards(prompt.ID, owner.ID, []uuid.UUID{beast}); err != nil {
		t.Fatalf("a lone creature card was refused on the restored table: %v", err)
	}
	if !inGraveyardOf(g, owner.ID, beast) {
		t.Error("the creature card was not discarded on the restored table")
	}
}

// Teferi Akosa, -3: the asking item rides the tap pick, so the reflexive trigger
// after a restore still offers the permanents that fit X.
func TestSweep06TeferiAkosaTapPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pw := pushTeferiAkosa(g, owner.ID, 4)
	a := pushCreatureToBattlefieldForTest(g, owner.ID, "Tapper A")
	b := pushCreatureToBattlefieldForTest(g, owner.ID, "Tapper B")
	cheap := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cheap, Name: "Two Drop", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(owner.ID, pw, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	sweep06PickKey(t, g, owner.ID, "choose-cards/teferi-akosa-tap-any-number")

	g = restoreRoundTrip(t, g, true)
	owner = g.Seats[seat]
	tap := sweep06PickKey(t, g, owner.ID, "choose-cards/teferi-akosa-tap-any-number")
	if err := g.ResolveChooseCards(tap.ID, owner.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("tap two creatures on the restored table: %v", err)
	}
	pick := latestPickTarget(g, owner.ID)
	if pick == nil {
		t.Fatal("no reflexive trigger asked for a target on the restored table")
	}
	if !slices.Contains(pick.PickTargetCards, cheap) {
		t.Error("the mana value 2 permanent was not offered to a two-creature tap")
	}
}

// Summon: Esper Valigarmanda, chapter I: the seats still to ask and the cards
// picked so far ride the prompt, so a restore between the two real choices
// loses neither the first pick nor the walk.
func TestSweep06ValigarmandaWalkIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp1, opp3 := g.Seats[(seat+1)%4], g.Seats[(seat+3)%4]
	valigarmandaGraveCard(me, "My Bolt", "Instant", "{R}")
	mySorcery := valigarmandaGraveCard(me, "My Divination", "Sorcery", "{2}{U}")
	valigarmandaGraveCard(opp1, "Opp1 Only Spell", "Instant", "{U}")
	valigarmandaGraveCard(opp3, "Opp3 First", "Instant", "{B}")
	opp3Second := valigarmandaGraveCard(opp3, "Opp3 Second", "Sorcery", "{1}{B}")
	saga := valigarmandaCast(t, g)
	passPriorityAroundTable(t, g)

	const key = "choose-cards/summon-esper-valigarmanda-exile-each"
	first := chooseCardsChoiceFor(g, me.ID)
	if first == nil || first.FromPlayer != me.ID {
		t.Fatalf("the first choice should be about my own graveyard: %+v", first)
	}
	answerChooseCards(t, g, me.ID, mySorcery)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[seat]
	second := sweep06PickKey(t, g, me.ID, key)
	if second.FromPlayer != opp3.ID {
		t.Fatalf("the second choice is about %v, want the third opponent's graveyard", second.FromPlayer)
	}
	answerChooseCards(t, g, me.ID, opp3Second)
	passPriorityAroundTable(t, g)
	got := exiledWithNamed(g, saga)
	for _, name := range []string{"My Divination", "Opp1 Only Spell", "Opp3 Second"} {
		if !got[name] {
			t.Errorf("%s is not exiled with the Saga on the restored table: %v", name, got)
		}
	}
}

// Summon: Kujata, chapter III: the discard is a restore point, and the draws
// and the damage trigger follow the answer on the restored table.
func TestSweep06KujataDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kujata := cloudDeckKujata(t, g)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	cloudDeckPickTargets(t, g, me.ID, kujata)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat) // chapter II
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	cloudDeckPickTargets(t, g, me.ID, kujata)
	passPriorityAroundTable(t, g)

	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fodder", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	before := me.Hand.Size()

	advanceToPrecombatMainOf(t, g, seat) // chapter III
	passPriorityAroundTable(t, g)
	sweep06PickKey(t, g, me.ID, "choose-cards/summon-kujata-fire-discard")

	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	answerDiscard(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand grew by %d on the restored table, want 2 (the turn's draw, discard one, draw two)", got)
	}
	if got := lives[opp.ID] - opp.Life; got != 3 {
		t.Errorf("opponent lost %d on the restored table, want 3", got)
	}
}

// Smuggler's Surprise: both picks are restore points. The mill ahead of the first
// still hands over its own closure, so the table is restored once the milled
// cards are on offer.
func TestSweep06SmugglersSurprisePicksAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := ssCard(me.Library, me.ID, "Milled Elf", "Creature — Elf", 1, 1)
	beast := ssCard(me.Library, me.ID, "Milled Beast", "Creature — Beast", 5, 5)
	me.Hand.Cards = nil
	giant := ssCard(me.Hand, me.ID, "Hand Giant", "Creature — Giant", 4, 4)

	castModal(t, g, "Smuggler's Surprise", "Instant", smugglersSurpriseOracle, []int{0, 1, 2}, nil)
	passPriorityAroundTable(t, g)
	sweep06PickKey(t, g, me.ID, "choose-cards/smugglers-surprise-take-milled")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := sweep06PickKey(t, g, me.ID, "choose-cards/smugglers-surprise-take-milled")
	if !hasID(pick.ChooseCards, elf) || !hasID(pick.ChooseCards, beast) {
		t.Fatalf("the restored pick does not offer the milled creatures: %v", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, beast)
	sweep06PickKey(t, g, me.ID, "choose-cards/smugglers-surprise-put-creatures")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick = sweep06PickKey(t, g, me.ID, "choose-cards/smugglers-surprise-put-creatures")
	if !hasID(pick.ChooseCards, giant) || !hasID(pick.ChooseCards, beast) {
		t.Fatalf("the restored second pick does not offer both creature cards in hand: %v", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, giant, beast)
	if !g.Battlefield.Contains(giant) || !g.Battlefield.Contains(beast) {
		t.Fatal("both creatures were not put onto the battlefield on the restored table")
	}
	for _, id := range []uuid.UUID{giant, beast} {
		ab := effectiveAbilities(t, g, id)
		if !hasAbility(ab, "hexproof") || !hasAbility(ab, "indestructible") {
			t.Errorf("a power-4+ creature gains hexproof and indestructible on the restored table: %v", ab)
		}
	}
}

// Sphinx Approach: "exactly four, or none" is a registered set rule, so the
// restored prompt still refuses two copies, and four are exiled with the spell
// before the Sphinx is fetched.
func TestSweep06SphinxsApproachKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sphinx := pushLibraryCardForTest(me, game.Card{Name: "Sphinx of Testing", TypeLine: "Creature — Sphinx", Power: 3, Toughness: 3})
	var copies []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sphinx's Approach", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		copies = append(copies, id)
	}
	spell := castCatalogSpell(t, g, "Sphinx's Approach", "Instant", rfbSphinxsApproachOracle, nil)
	passPriorityAroundTable(t, g)
	sweep06PickKey(t, g, me.ID, "choose-cards/sphinxs-approach-exile-four")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := sweep06PickKey(t, g, me.ID, "choose-cards/sphinxs-approach-exile-four")
	if err := g.ResolveChooseCards(pick.ID, me.ID, copies[:2]); err == nil {
		t.Fatal("the restored prompt accepted two copies: the set rule did not survive")
	}
	answerChooseCards(t, g, me.ID, copies...)
	if search := searchChoiceFor(g, me.ID); search != nil {
		answerSearchByID(t, g, me.ID, sphinx)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sphinx) {
		t.Error("the Sphinx is not on the battlefield on the restored table")
	}
	for _, id := range append(copies, spell) {
		if !g.Exile.Contains(id) {
			t.Errorf("card %s is not exiled on the restored table", id)
		}
	}
}

// Tether Technician: the optional discard is a restore point, and the
// reflexive 2 damage follows the answer on the restored table.
func TestSweep06TetherTechnicianDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, foe := g.Seats[seat], rfCrEFoe(g)
	rfCrACast(t, g, "Tether Technician", "Creature — Minotaur Artificer", oracleTetherTech, 4, 5)
	passPriorityAroundTable(t, g)
	sweep06PickKey(t, g, me.ID, "choose-cards/tether-technician-strike")
	life := foe.Life

	g = restoreRoundTrip(t, g, true)
	me, foe = g.Seats[seat], rfCrEFoe(g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	pickPlayerTarget(t, g, me.ID, foe.ID)
	passPriorityAroundTable(t, g)
	if foe.Life != life-2 {
		t.Fatalf("life %d on the restored table, want %d", foe.Life, life-2)
	}
}
