package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entry_controller_test.go — ADR 0102, #1759: "this enters under the
// control of an opponent of your choice". Every test goes through the
// real entry door (a cast, a token copy, a Clone) rather than a
// hand-built event.

const (
	entryHostageCreatureOracle = "test-adr0102-hostage-creature"
	entryHostageEnchantOracle  = "test-adr0102-hostage-enchantment"
	authorityOfConsulsOracle   = "55f3c721-e13a-406e-bc8e-d6cdc91ac477"
	cloneOracle                = "42226b87-0746-4ebf-9fd0-108d508462af"
)

// registerHostages registers two test permanents carrying the clause:
// a creature (so Kismet, Authority of the Consuls, summoning sickness
// and a Clone have something to act on) and an enchantment.
func registerHostages(t *testing.T) {
	t.Helper()
	registerForTest(t, Spec{
		OracleID:     entryHostageCreatureOracle,
		Name:         "Test Hostage",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersUnderTheControlOfAnOpponentOfYourChoice("Test Hostage", game.ControlForHarm),
		},
	})
	registerForTest(t, Spec{
		OracleID:     entryHostageEnchantOracle,
		Name:         "Test Gift",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersUnderTheControlOfAnOpponentOfYourChoice("Test Gift", game.ControlForBenefit),
		},
	})
}

func entryControllerPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryController {
			return c
		}
	}
	return nil
}

// castHostage casts a hostage from the active seat and passes priority
// until the entry asks. Returns the card's ID.
func castHostage(t *testing.T, g *game.Game, oracle, typeLine string) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Hostage", typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

// answerEntryController answers the open prompt with the seat `want`.
func answerEntryController(t *testing.T, g *game.Game, want uuid.UUID) {
	t.Helper()
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("no entry_controller prompt is open")
	}
	for i, opt := range c.PickOptions {
		if opt.Player == want {
			if err := g.ResolveEntryController(c.ID, c.Chooser, i); err != nil {
				t.Fatalf("ResolveEntryController: %v", err)
			}
			passPriorityAroundTable(t, g)
			return
		}
	}
	t.Fatalf("seat %s is not offered", want)
}

// TestEntryControllerAsksTheCasterAndLandsUnderTheChosenOpponent is
// the whole clause: the caster is asked, the options are the three
// opponents in turn order starting after the caster, and the permanent
// lands under the chosen one — owned by the caster, with no control
// CHANGE (it entered under that player, CR 110.2).
func TestEntryControllerAsksTheCasterAndLandsUnderTheChosenOpponent(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	id := castHostage(t, g, entryHostageCreatureOracle, "Creature — Human")

	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("no entry_controller prompt after the spell resolved")
	}
	if c.Chooser != caster.ID {
		t.Fatalf("chooser = %s, want the caster", c.Chooser)
	}
	if c.ControlPurpose != game.ControlForHarm {
		t.Fatalf("purpose = %q, want harm", c.ControlPurpose)
	}
	if _, onBF := battlefieldCardByID(g, id); onBF {
		t.Fatal("the permanent is on the battlefield before the opponent was chosen")
	}
	start := g.Turn.ActiveSeat
	if len(c.PickOptions) != 3 {
		t.Fatalf("options = %d, want 3", len(c.PickOptions))
	}
	for k, opt := range c.PickOptions {
		want := g.Seats[(start+1+k)%4].ID
		if opt.Player != want {
			t.Fatalf("option %d = %s, want seat %d (turn order after the caster)", k, opt.Player, (start+1+k)%4)
		}
	}

	chosen := g.Seats[(start+2)%4]
	answerEntryController(t, g, chosen.ID)

	card, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the permanent never landed")
	}
	if card.Controller != chosen.ID {
		t.Fatalf("controller = %s, want the chosen opponent %s", card.Controller, chosen.ID)
	}
	if card.Owner != caster.ID {
		t.Fatalf("owner = %s, want the caster (CR 108.3)", card.Owner)
	}
	if n := countEvents(g, game.EventControlChanged); n != 0 {
		t.Fatalf("%d EventControlChanged fired; entering under a player is not a change of control", n)
	}
	if n := countEvents(g, game.EventEntryControllerChosen); n != 1 {
		t.Fatalf("EventEntryControllerChosen fired %d times, want 1", n)
	}
	// CR 110.2: the default controller is the player it entered under,
	// so a later control effect ending would give it back to them.
	g.ReadSnapshot(func() {})
	card, _ = battlefieldCardByID(g, id)
	if card.BaseController != uuid.Nil && card.BaseController != chosen.ID {
		t.Fatalf("BaseController = %s, want the chosen opponent", card.BaseController)
	}
}

// TestEntryControllerForcedWithOneOpponent is owner decision 3: at a
// two-seat table the choice is forced, so no prompt is shown.
func TestEntryControllerForcedWithOneOpponent(t *testing.T) {
	registerHostages(t)
	g := newTwoSeatCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1-g.Turn.ActiveSeat]
	id := castHostage(t, g, entryHostageEnchantOracle, "Enchantment")
	if c := entryControllerPrompt(g); c != nil {
		t.Fatal("a prompt was shown with only one opponent to choose")
	}
	card, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the permanent never landed")
	}
	if card.Controller != opp.ID || card.Owner != caster.ID {
		t.Fatalf("controller/owner = %s/%s, want %s/%s", card.Controller, card.Owner, opp.ID, caster.ID)
	}
	if n := countEvents(g, game.EventEntryControllerChosen); n != 1 {
		t.Fatalf("EventEntryControllerChosen fired %d times, want 1 (the forced choice is still logged)", n)
	}
}

// TestEntryControllerIsOrderedBeforeKismet is CR 616.1b, the ADR's
// worked example: with the control change applied first, a creature
// given to Kismet's controller is not an opponent's creature and
// enters untapped; one given to anybody else enters tapped. And the
// affected player is never asked to order the two.
func TestEntryControllerIsOrderedBeforeKismet(t *testing.T) {
	for _, tc := range []struct {
		name       string
		toKismet   bool
		wantTapped bool
	}{
		{"given to Kismet's controller", true, false},
		{"given to another opponent", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registerHostages(t)
			g := newCatalogGame(t)
			start := g.Turn.ActiveSeat
			kismetSeat := g.Seats[(start+2)%4]
			other := g.Seats[(start+3)%4]
			pushPermanentForTest(g, kismetSeat.ID, "Kismet", kismetOracle, "Enchantment")

			id := castHostage(t, g, entryHostageCreatureOracle, "Creature — Human")
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
					t.Fatal("the affected player was asked to order the control change against Kismet (CR 616.1b)")
				}
			}
			target := other
			if tc.toKismet {
				target = kismetSeat
			}
			answerEntryController(t, g, target.ID)
			card, ok := battlefieldCardByID(g, id)
			if !ok {
				t.Fatal("the permanent never landed")
			}
			if card.Controller != target.ID {
				t.Fatalf("controller = %s, want %s", card.Controller, target.ID)
			}
			if card.Tapped != tc.wantTapped {
				t.Fatalf("tapped = %v, want %v", card.Tapped, tc.wantTapped)
			}
		})
	}
}

// TestEntryControllerLetsAuthorityOfTheConsulsSeeTheNewController is
// CR 616.1f's other half: Authority of the Consuls did not apply to
// the caster's own creature, and does once the creature is entering
// under an opponent — it enters tapped, and Authority's "whenever a
// creature an opponent controls enters" gains its controller a life.
// The caster's own "creature you control enters" reading does not see
// it.
func TestEntryControllerLetsAuthorityOfTheConsulsSeeTheNewController(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	pushPermanentForTest(g, caster.ID, "Authority of the Consuls", authorityOfConsulsOracle, "Enchantment")
	lifeBefore := caster.Life

	id := castHostage(t, g, entryHostageCreatureOracle, "Creature — Human")
	target := g.Seats[(g.Turn.ActiveSeat+1)%4]
	answerEntryController(t, g, target.ID)
	passPriorityAroundTable(t, g)

	card, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the permanent never landed")
	}
	if !card.Tapped {
		t.Fatal("Authority of the Consuls did not tap an opponent's entering creature")
	}
	if caster.Life != lifeBefore+1 {
		t.Fatalf("caster life = %d, want %d (Authority's opponent-creature trigger)", caster.Life, lifeBefore+1)
	}
}

// TestEntryControllerSummoningSicknessFollowsTheNewController is
// CR 302.6: the creature is sick for the player it entered under, and
// stops being sick when that player's turn begins.
func TestEntryControllerSummoningSicknessFollowsTheNewController(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	start := g.Turn.ActiveSeat
	id := castHostage(t, g, entryHostageCreatureOracle, "Creature — Human")
	next := (start + 1) % 4
	answerEntryController(t, g, g.Seats[next].ID)

	card, _ := battlefieldCardByID(g, id)
	if !game.HasSummoningSickness(&card) {
		t.Fatal("the creature is not summoning sick the turn it entered")
	}
	advanceToUpkeepOf(t, g, next)
	card, _ = battlefieldCardByID(g, id)
	if game.HasSummoningSickness(&card) {
		t.Fatal("the creature is still sick on its controller's own turn")
	}
}

// TestEntryControllerControllerLeavesExilesIt and the owner case below
// are the two rulings (Captive Audience, Pendant of Prosperity,
// 2019): the controller leaving exiles it, the owner leaving takes it
// out of the game.
func TestEntryControllerControllerLeavesExilesIt(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	start := g.Turn.ActiveSeat
	id := castHostage(t, g, entryHostageEnchantOracle, "Enchantment")
	holder := g.Seats[(start+1)%4]
	answerEntryController(t, g, holder.ID)
	if err := g.Concede(holder.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if _, ok := battlefieldCardByID(g, id); ok {
		t.Fatal("still on the battlefield after its controller left")
	}
	found := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			found = true
		}
	}
	if !found {
		t.Fatal("not exiled after its (non-owner) controller left (CR 800.4a)")
	}
}

func TestEntryControllerOwnerLeavesTakesItOutOfTheGame(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	start := g.Turn.ActiveSeat
	owner := g.Seats[start]
	id := castHostage(t, g, entryHostageEnchantOracle, "Enchantment")
	answerEntryController(t, g, g.Seats[(start+1)%4].ID)
	if err := g.Concede(owner.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if _, ok := battlefieldCardByID(g, id); ok {
		t.Fatal("still on the battlefield after its owner left")
	}
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			t.Fatal("exiled; the owner leaving takes it out of the game instead")
		}
	}
}

// TestEntryControllerOfferedSeatLeavingPrunesTheOption: an opponent
// who concedes while the question is open comes off it, and the answer
// still lands.
func TestEntryControllerOfferedSeatLeavingPrunesTheOption(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	start := g.Turn.ActiveSeat
	id := castHostage(t, g, entryHostageEnchantOracle, "Enchantment")
	leaver := g.Seats[(start+1)%4]
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("the prompt was dropped when one offered seat left")
	}
	for _, opt := range c.PickOptions {
		if opt.Player == leaver.ID {
			t.Fatal("a seat that left the game is still offered (CR 800.4a)")
		}
	}
	target := g.Seats[(start+3)%4]
	answerEntryController(t, g, target.ID)
	if card, ok := battlefieldCardByID(g, id); !ok || card.Controller != target.ID {
		t.Fatal("the permanent did not land under the chosen opponent after the prune")
	}
}

// TestEntryControllerChooserLeavingDropsThePrompt: the caster
// conceding with the question open takes their card out of the game,
// and leaves nothing waiting on an answer nobody can give.
func TestEntryControllerChooserLeavingDropsThePrompt(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	id := castHostage(t, g, entryHostageEnchantOracle, "Enchantment")
	if entryControllerPrompt(g) == nil {
		t.Fatal("no prompt")
	}
	if err := g.Concede(caster.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if entryControllerPrompt(g) != nil {
		t.Fatal("the prompt survived its chooser leaving")
	}
	if _, ok := battlefieldCardByID(g, id); ok {
		t.Fatal("the departed caster's card entered the battlefield")
	}
}

// TestEntryControllerTokenCopyIsOwnedByItsCreator is Xantcha's 2018
// ruling: a token copy's creator owns it, and it enters under the
// opponent the creator chooses.
func TestEntryControllerTokenCopyIsOwnedByItsCreator(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	start := g.Turn.ActiveSeat
	me := g.Seats[start]
	src := pushPermanentForTest(g, g.Seats[(start+2)%4].ID, "Test Hostage", entryHostageCreatureOracle, "Creature — Human")
	before := len(g.Battlefield.Cards)
	g.WithWriteLock(func() {
		if err := (CreateTokenCopy{Controller: me.ID, Copy: src, N: 1}).Apply(NewContext(g, nil)); err != nil {
			t.Errorf("CreateTokenCopy: %v", err)
		}
	})
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("the token copy's entry did not ask")
	}
	if c.Chooser != me.ID {
		t.Fatalf("chooser = %s, want the token's creator", c.Chooser)
	}
	target := g.Seats[(start+1)%4]
	if err := g.ResolveEntryController(c.ID, c.Chooser, 0); err != nil {
		t.Fatalf("ResolveEntryController: %v", err)
	}
	if len(g.Battlefield.Cards) != before+1 {
		t.Fatalf("battlefield has %d cards, want %d", len(g.Battlefield.Cards), before+1)
	}
	tok := g.Battlefield.Cards[len(g.Battlefield.Cards)-1]
	for _, bc := range g.Battlefield.Cards {
		if bc.InstanceID != src && IsToken(bc) {
			tok = bc
		}
	}
	if tok.Owner != me.ID {
		t.Fatalf("token owner = %s, want its creator", tok.Owner)
	}
	if tok.Controller != target.ID {
		t.Fatalf("token controller = %s, want the chosen opponent", tok.Controller)
	}
}

// TestEntryControllerCloneCopyingItAsksToo is ADR 0102 decision 6 and
// CR 614.12: once a Clone has chosen to copy the hostage, it would
// exist on the battlefield as the hostage, so the hostage's clause
// applies to its entry.
func TestEntryControllerCloneCopyingItAsksToo(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	start := g.Turn.ActiveSeat
	me := g.Seats[start]
	src := pushPermanentForTest(g, g.Seats[(start+2)%4].ID, "Test Hostage", entryHostageCreatureOracle, "Creature — Human")
	cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", cloneOracle, nil)
	passPriorityAroundTable(t, g)
	cp := copyPrompt(g)
	if cp == nil {
		t.Fatal("Clone did not ask what to copy")
	}
	if err := g.ResolveCopyTarget(cp.ID, cp.Chooser, src); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("a Clone copying the hostage did not ask whose control it enters under")
	}
	if c.Chooser != me.ID {
		t.Fatalf("chooser = %s, want the Clone's caster", c.Chooser)
	}
	target := g.Seats[(start+3)%4]
	answerEntryController(t, g, target.ID)
	card, ok := battlefieldCardByID(g, cloneID)
	if !ok {
		t.Fatal("the Clone never landed")
	}
	if card.Controller != target.ID || card.Owner != me.ID {
		t.Fatalf("Clone controller/owner = %s/%s, want %s/%s", card.Controller, card.Owner, target.ID, me.ID)
	}
}

// TestEntryControllerLandDropTallyIsThePlayers pins the landPlayer
// split: the land-drop tally is keyed on the player who played the
// land, never on ev.Actor, which an entry-controller effect rewrites.
// No printed land carries the clause, so this is the plain land play
// still counting for the player who made it.
func TestEntryControllerLandDropTallyIsThePlayers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play land: %v", err)
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 1 {
		t.Fatalf("lands played = %d, want 1", n)
	}
}

// TestEntryControllerEnumeratorOffersEachOpponentAndDispatches is the
// bot's half (#544's rule): the enumerator offers one answer per
// opponent to the chooser, nothing else while the prompt blocks the
// table, and every offered answer is one the dispatcher accepts.
func TestEntryControllerEnumeratorOffersEachOpponentAndDispatches(t *testing.T) {
	registerHostages(t)
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	id := castHostage(t, g, entryHostageEnchantOracle, "Enchantment")
	c := entryControllerPrompt(g)
	if c == nil {
		t.Fatal("no prompt")
	}
	var answers []legal.Move
	for _, m := range legal.EnumerateFor(g, caster.ID) {
		if m.Type != legal.TypeResolveChoice {
			t.Fatalf("the chooser was offered %q while the entry prompt blocks the table", m.Type)
		}
		answers = append(answers, m)
	}
	if len(answers) != 3 {
		t.Fatalf("enumerated %d answers, want one per opponent (3)", len(answers))
	}
	last := answers[2]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(last.Type), Player: last.Player, Caller: last.Player, Params: last.Params,
	}); err != nil {
		t.Fatalf("dispatch the enumerated answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the permanent never landed")
	}
	if card.Controller != c.PickOptions[2].Player {
		t.Fatalf("controller = %s, want the third offered seat", card.Controller)
	}
}
