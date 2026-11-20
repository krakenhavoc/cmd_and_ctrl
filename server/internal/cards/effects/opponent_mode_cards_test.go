package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_mode_cards_test.go — #5771, CR 700.2e: "An opponent chooses
// one —". Each card is cast for real at a four-seat table: the caster's
// payload names no mode, the caster names which opponent chooses, that
// opponent answers a mode_pick, the caster then picks any targets, and
// only then is anything paid or put on the stack.

const (
	omFatalLore       = "aeb85786-16fc-45d8-9383-60054a878c40"
	omLibraryOfLatNam = "67f68a0f-bd3d-47ba-ad99-95c1a2acd170"
	omMisfortune      = "e282eb5b-f9e6-4fb3-9edf-44e23359357d"
)

// omCast puts `name` in the active seat's hand at its main phase and
// casts it with `params`, returning the card and the error.
func omCast(t *testing.T, g *game.Game, name, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return id, g.CastSpell(me.ID, id, params)
}

// omModePick is the open mode_pick addressed to `chooser`, or nil.
func omModePick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceModePick && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// omChoose answers chooser's open mode_pick with `mode`.
func omChoose(t *testing.T, g *game.Game, chooser uuid.UUID, mode int) {
	t.Helper()
	c := omModePick(g, chooser)
	if c == nil {
		t.Fatalf("no mode_pick open for %s", chooser)
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

// omStartFatalLore casts Fatal Lore at a four-seat table and names seat 1
// to choose the mode. The table has a creature for every opponent.
func omStartFatalLore(t *testing.T) (g *game.Game, spell uuid.UUID, mine, theirs, other []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine = []uuid.UUID{pr7Creature(g, me.ID, "My Bear", 2)}
	theirs = []uuid.UUID{pr7Creature(g, a.ID, "Their Bear", 2), pr7Creature(g, a.ID, "Their Wolf", 3), pr7Creature(g, a.ID, "Their Elk", 4)}
	other = []uuid.UUID{pr7Creature(g, b.ID, "Other Bear", 2)}
	spell, err := omCast(t, g, "Fatal Lore", omFatalLore, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if len(g.StackMeta) != 0 || !me.Hand.Contains(spell) {
		t.Fatal("Fatal Lore was put on the stack before its mode was chosen")
	}
	// Every opponent could choose (the draw bullet always can), so the
	// caster says which one (CR 700.2e).
	if got := otOfferedSeats(otOptionPick(t, g, me.ID)); len(got) != 3 {
		t.Fatalf("offered %d opponents, want 3", len(got))
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	return g, spell, mine, theirs, other
}

// The caster cannot name the mode; the opponent the caster named gets
// the prompt, and nobody else does; it names the spell and its caster.
func TestFatalLoreTheOpponentChoosesTheMode(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := omCast(t, g, "Fatal Lore", omFatalLore, game.CastSpellParams{Modes: []int{0}}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("a caster naming the mode: %v, want ErrInvalidParam", err)
	}
	if len(g.PendingChoices) != 0 || len(g.StackMeta) != 0 {
		t.Fatal("a refused cast left a prompt or a stack item")
	}

	g, spell, _, _, _ := omStartFatalLore(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	pick := omModePick(g, a.ID)
	if pick == nil {
		t.Fatal("the named opponent was not asked for the mode")
	}
	for _, p := range []*game.Player{me, b, c} {
		if omModePick(g, p.ID) != nil {
			t.Fatalf("seat %s was asked for the mode too", p.Name)
		}
	}
	if pick.ModePickFor != me.ID || pick.OpponentModeSpellName() != "Fatal Lore" {
		t.Fatalf("the prompt is for %s / %q, want the caster and Fatal Lore", pick.ModePickFor, pick.OpponentModeSpellName())
	}
	if !containsText(pick.Reason, "Fatal Lore") || !containsText(pick.Reason, me.Name) {
		t.Fatalf("the prompt %q does not name the spell and its caster", pick.Reason)
	}
	if !slices.Equal(pick.ModeOptionIndex, []int{0, 1}) {
		t.Fatalf("offered bullets %v, want both", pick.ModeOptionIndex)
	}
	// The caster cannot answer it.
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); !errors.Is(err, game.ErrNotTheChooser) {
		t.Fatalf("the caster answering the opponent's mode: %v, want ErrNotTheChooser", err)
	}
	// A second cast of the same card while this one waits is refused.
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("a second cast while parked: %v, want ErrChoicePending", err)
	}
	omChoose(t, g, a.ID, 0)
	item := otItemFrom(g, spell)
	if item == nil || !slices.Equal(item.Modes, []int{0}) || item.ModeChooser != a.ID {
		t.Fatalf("the spell's item = %+v, want mode 0 chosen by %s", item, a.Name)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Fatalf("caster's hand %d, want %d: \"you draw three cards\"", got, hand+3)
	}
}

// The destroy bullet: the caster chooses up to two of THAT opponent's
// creatures after the mode (the 2004 ruling), they are destroyed, and
// that opponent draws up to three.
func TestFatalLoreDestroyBulletTargetsThatPlayersCreatures(t *testing.T) {
	g, spell, mine, theirs, other := omStartFatalLore(t)
	me, a := g.Seats[0], g.Seats[1]
	omChoose(t, g, a.ID, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the caster was not asked for the destroy bullet's targets")
	}
	if p.PickTargetFor != uuid.Nil || p.PickTargetMin != 0 || p.PickTargetMax != 2 {
		t.Fatalf("the caster's pick: for %s, %d..%d, want their own, 0..2", p.PickTargetFor, p.PickTargetMin, p.PickTargetMax)
	}
	for _, id := range theirs {
		if !slices.Contains(p.PickTargetCards, id) {
			t.Fatalf("that player's creature %s is not on offer", id)
		}
	}
	for _, id := range append(append([]uuid.UUID(nil), mine...), other...) {
		if slices.Contains(p.PickTargetCards, id) {
			t.Fatalf("%s is on offer, but only that player's creatures are", id)
		}
	}
	if len(g.StackMeta) != 0 {
		t.Fatal("the spell reached the stack before its targets were chosen")
	}
	otAnswer(t, g, me.ID, otCard(theirs[0]), otCard(theirs[1]))
	item := otItemFrom(g, spell)
	if item == nil || item.ModeChooser != a.ID || len(item.Targets) != 2 {
		t.Fatalf("the spell's item = %+v", item)
	}
	passPriorityAroundTable(t, g)
	for _, id := range theirs[:2] {
		if g.Battlefield.Contains(id) {
			t.Fatalf("target %s was not destroyed", id)
		}
	}
	if !g.Battlefield.Contains(theirs[2]) || !g.Battlefield.Contains(other[0]) {
		t.Fatal("a creature that was not a target was destroyed")
	}
	hand := a.Hand.Size()
	answerNumber(t, g, a.ID, 2)
	if got := a.Hand.Size(); got != hand+2 {
		t.Fatalf("that player's hand %d, want %d: they chose to draw two", got, hand+2)
	}
}

// With no creature that player controls, the destroy bullet is still a
// legal choice ("up to two"), and the caster is asked nothing more.
func TestFatalLoreDestroyBulletWithNothingToTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	pr7Creature(g, g.Seats[2].ID, "Other Bear", 2)
	spell, err := omCast(t, g, "Fatal Lore", omFatalLore, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	omChoose(t, g, a.ID, 1)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the caster was asked for targets that player does not have")
	}
	item := otItemFrom(g, spell)
	if item == nil || len(item.Targets) != 0 || item.ModeChooser != a.ID {
		t.Fatalf("the spell's item = %+v, want no target, chosen by that player", item)
	}
	passPriorityAroundTable(t, g)
	hand := a.Hand.Size()
	answerNumber(t, g, a.ID, 3)
	if got := a.Hand.Size(); got != hand+3 {
		t.Fatalf("that player's hand %d, want %d", got, hand+3)
	}
}

// A restore while the opponent is choosing resumes the cast: the
// restored table still waits on the same seat, and its answer, then the
// caster's targets, put the spell on the stack there.
func TestFatalLoreRestoredMidChoiceResumes(t *testing.T) {
	g, spell, _, theirs, _ := omStartFatalLore(t)
	g = s13StrictRestore(t, g, game.OpponentModeAnnounceKey)
	me, a := g.Seats[0], g.Seats[1]
	pick := omModePick(g, a.ID)
	if pick == nil || pick.ModePickFor != me.ID || pick.OpponentModeSpellName() != "Fatal Lore" {
		t.Fatalf("restored mode_pick = %+v", pick)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); !errors.Is(err, game.ErrNotTheChooser) {
		t.Fatalf("the caster answering after a restore: %v, want ErrNotTheChooser", err)
	}
	omChoose(t, g, a.ID, 1)
	// The caster's pick is a restore point too.
	g = s13StrictRestore(t, g, game.OpponentTargetAnnounceKey)
	me = g.Seats[0]
	if p := latestPickTarget(g, me.ID); p == nil || p.PickTargetFor != uuid.Nil {
		t.Fatalf("restored caster's pick = %+v", p)
	}
	otAnswer(t, g, me.ID, otCard(theirs[2]))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs[2]) {
		t.Fatal("the target was not destroyed on the restored table")
	}
	answerNumber(t, g, g.Seats[1].ID, 0)
	passPriorityAroundTable(t, g)
	if !g.Seats[0].Graveyard.Contains(spell) {
		t.Fatal("Fatal Lore did not resolve on the restored table")
	}
}

// The caster's "which opponent" question is a restore point as well.
func TestOpponentModeChooserQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if _, err := omCast(t, g, "Misfortune", omMisfortune, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	otOptionPick(t, g, me.ID)
	g = restoreRoundTrip(t, g, true)
	me, b := g.Seats[0], g.Seats[2]
	otChooseOpponent(t, g, me.ID, b.ID)
	if omModePick(g, b.ID) == nil {
		t.Fatal("the opponent named on the restored table was not asked")
	}
}

// An opponent who leaves while choosing passes the choice to another
// opponent (CR 800.4g), never to the caster.
func TestOpponentModeChoicePassesToAnotherOpponent(t *testing.T) {
	g, spell, _, _, _ := omStartFatalLore(t)
	me, a := g.Seats[0], g.Seats[1]
	if err := g.Concede(a.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if omModePick(g, me.ID) != nil {
		t.Fatal("the caster was handed the mode choice")
	}
	var heir *game.PendingChoice
	for _, p := range g.Seats[2:] {
		if c := omModePick(g, p.ID); c != nil {
			heir = c
		}
	}
	if heir == nil {
		t.Fatal("no other opponent inherited the mode choice")
	}
	omChoose(t, g, heir.Chooser, 0)
	if item := otItemFrom(g, spell); item == nil || item.ModeChooser != heir.Chooser {
		t.Fatalf("the spell's item = %+v, want the mode chosen by the heir", item)
	}
}

// A caster who leaves takes the parked cast with them (CR 800.4a).
func TestOpponentModeChoiceIsWithdrawnWhenTheCasterLeaves(t *testing.T) {
	g, spell, _, _, _ := omStartFatalLore(t)
	me, a := g.Seats[0], g.Seats[1]
	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if omModePick(g, a.ID) != nil {
		t.Fatal("the opponent still owes a mode for a player who has left")
	}
	if otItemFrom(g, spell) != nil {
		t.Fatal("the departed caster's spell reached the stack")
	}
}

// Library of Lat-Nam: the draw bullet draws three at the next upkeep;
// the search bullet tutors. With one opponent left to choose, the engine
// names them and the caster is not asked.
func TestLibraryOfLatNamBothBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	if _, err := omCast(t, g, "Library of Lat-Nam", omLibraryOfLatNam, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	omChoose(t, g, a.ID, 0)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	for i := 0; g.Turn.Step != game.StepUpkeep || g.Turn.ActiveSeat == 0; i++ {
		if i > 40 {
			t.Fatal("never reached the next upkeep")
		}
		if len(g.StackMeta) > 0 {
			passPriorityAroundTable(t, g)
			continue
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Fatalf("caster's hand %d at the next upkeep, want %d", got, hand+3)
	}

	g = newCatalogGame(t)
	me, a = g.Seats[0], g.Seats[1]
	for _, p := range g.Seats[2:] {
		if err := g.Concede(p.ID); err != nil {
			t.Fatalf("Concede: %v", err)
		}
	}
	needle := pushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Instant"})
	if _, err := omCast(t, g, "Library of Lat-Nam", omLibraryOfLatNam, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			t.Fatal("the caster was asked which opponent with only one left")
		}
	}
	omChoose(t, g, a.ID, 1)
	passPriorityAroundTable(t, g)
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatal("the search bullet asked nothing")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{needle}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(needle) {
		t.Fatal("the searched card is not in the caster's hand")
	}
}

// Misfortune: the first bullet is the caster's counters and life; the
// second puts -1/-1 counters on THAT player's creatures and deals that
// player 4.
func TestMisfortuneBothBullets(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	mine := pr7Creature(g, me.ID, "Mine", 2)
	theirs := pr7Creature(g, a.ID, "Theirs", 2)
	if _, err := omCast(t, g, "Misfortune", omMisfortune, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	omChoose(t, g, a.ID, 0)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatalf("caster's life %d, want %d", me.Life, life+4)
	}
	if c := findBattlefieldCardForTest(g, mine); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Fatal("the caster's creature has no +1/+1 counter")
	}
	if c := findBattlefieldCardForTest(g, theirs); c == nil || c.Counters[game.CounterPlusOne] != 0 {
		t.Fatal("an opponent's creature got a +1/+1 counter")
	}

	g = newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine = pr7Creature(g, me.ID, "Mine", 2)
	theirs = pr7Creature(g, a.ID, "Theirs", 2)
	other := pr7Creature(g, b.ID, "Other", 2)
	if _, err := omCast(t, g, "Misfortune", omMisfortune, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	otChooseOpponent(t, g, me.ID, a.ID)
	omChoose(t, g, a.ID, 1)
	life, bLife := a.Life, b.Life
	passPriorityAroundTable(t, g)
	if a.Life != life-4 || b.Life != bLife {
		t.Fatalf("that player's life %d (want %d), another's %d (want %d)", a.Life, life-4, b.Life, bLife)
	}
	if c := findBattlefieldCardForTest(g, theirs); c == nil || c.Counters[game.CounterMinusOne] != 1 {
		t.Fatal("that player's creature has no -1/-1 counter")
	}
	for _, id := range []uuid.UUID{mine, other} {
		if c := findBattlefieldCardForTest(g, id); c == nil || c.Counters[game.CounterMinusOne] != 0 {
			t.Fatal("a creature that player does not control got a -1/-1 counter")
		}
	}
}

// The three cards declare the seam as data.
func TestOpponentChoosesOneIsDeclared(t *testing.T) {
	for _, oracle := range []string{omFatalLore, omLibraryOfLatNam, omMisfortune} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Modes == nil || !spec.Modes.ChosenByOpponent || spec.Completeness != CompletenessFull {
			t.Errorf("%s: modes %+v, completeness %v", spec.Name, spec.Modes, spec.Completeness)
		}
	}
}
