package testss82v

import (
	"slices"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// secret_choice_cards_test.go — #5948 Q3: the five cards built on the
// secret choice round (game/secret_choice_round.go, ADR 0180 amendment
// 2026-11-24). Every answer is made through the engine's own prompts, and
// the hiding is read off the per-seat wire view.

const (
	secretItazura   = "55b45ac9-9ce7-4d17-9efb-3c0c51482950"
	secretDCI       = "27a7380d-4c18-4754-b5fb-1302d939c9d9"
	secretToymaker  = "8e93ea21-5e3f-448e-a175-96c8438c328e"
	secretGollum    = "3fb5c9a6-3fab-4109-875b-83c92518c5eb"
	secretCallVoid  = "c2db2c28-3ddc-41ce-a0c2-ddc555dea133"
	secretBearTL    = "Creature — Bear"
	secretSorceryTL = "Sorcery"
)

func TestSecretChoiceCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		secretItazura:  "Itazura, Lingering Wick",
		secretDCI:      "Duelists' Convocation International",
		secretToymaker: "The Toymaker's Trap",
		secretGollum:   "Gollum, Scheming Guide",
		secretCallVoid: "Call to the Void",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// secretWireBallots is the ballots the open round shows `viewer` on the
// wire: nil when no round is open.
func secretWireBallots(t *testing.T, g *game.Game, viewer string) []protocol.CouncilBallotView {
	t.Helper()
	var out []protocol.CouncilBallotView
	g.ReadSnapshot(func() {
		for _, c := range protocol.FilterViewFor(protocol.ViewOfGame(g), viewer).PendingChoices {
			if c.CouncilVote != nil {
				if !c.CouncilVote.Secret || !c.CouncilVote.Choice {
					t.Errorf("an open round is not a secret choice on the wire: %+v", c.CouncilVote)
				}
				out = append(out, c.CouncilVote.Ballots...)
			}
		}
	})
	return out
}

func secretRevealed(g *game.Game) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventVotesRevealed {
			n++
		}
	}
	return n
}

func secretOnBattlefield(g *game.Game, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() { ok = effects.OnBattlefield(g, id) })
	return ok
}

// --- Call to the Void ------------------------------------------------------

func TestCallToTheVoidDestroysEachCreatureChosenAfterTheReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	mineA := effectstest.B12Creature(g, me.ID, "Mine A", secretBearTL, 2, 2)
	mineB := effectstest.B12Creature(g, me.ID, "Mine B", secretBearTL, 2, 2)
	bobs := effectstest.B12Creature(g, b.ID, "Bob Bear", secretBearTL, 2, 2)
	cys := effectstest.B12Creature(g, c.ID, "Cy Bear", secretBearTL, 2, 2)
	// d controls no creature: their "a creature you control" is skipped.
	effectstest.CastCatalogSpell(t, g, "Call to the Void", secretSorceryTL, secretCallVoid, nil)
	effectstest.PassPriorityAroundTable(t, g)

	// The caster first: one of their own, then one of someone else's.
	own := effectstest.BallotFor(g, me.ID)
	if own == nil {
		t.Fatal("no ask for the caster")
	}
	var labels []string
	for _, o := range own.PickOptions {
		labels = append(labels, o.Label)
	}
	if !slices.Equal(labels, []string{"Mine A", "Mine B"}) {
		t.Fatalf("a creature you control offers %v", labels)
	}
	effectstest.VoteFor(t, g, me.ID, "Mine B")
	// Nobody else sees it; the caster does.
	if got := secretWireBallots(t, g, b.ID.String()); len(got) != 0 {
		t.Errorf("Bob sees %v before the reveal", got)
	}
	if got := secretWireBallots(t, g, protocol.SpectatorViewerID); len(got) != 0 {
		t.Errorf("a spectator sees %v before the reveal", got)
	}
	if got := secretWireBallots(t, g, me.ID.String()); len(got) != 1 {
		t.Errorf("the caster sees %d of their own answers, want 1", len(got))
	}
	labels = nil
	for _, o := range effectstest.BallotFor(g, me.ID).PickOptions {
		labels = append(labels, o.Label)
	}
	if !slices.Equal(labels, []string{"Bob Bear", "Cy Bear"}) {
		t.Fatalf("a creature you don't control offers %v", labels)
	}
	effectstest.VoteFor(t, g, me.ID, "Cy Bear")
	effectstest.VoteFor(t, g, b.ID, "Bob Bear")
	effectstest.VoteFor(t, g, b.ID, "Cy Bear")
	effectstest.VoteFor(t, g, c.ID, "Cy Bear")
	if !secretOnBattlefield(g, cys) || secretRevealed(g) != 0 {
		t.Fatal("something happened before every choice was in")
	}
	effectstest.VoteFor(t, g, c.ID, "Mine B")
	// d has no creature of their own: their only ask is someone else's.
	if effectstest.BallotFor(g, d.ID) == nil {
		t.Fatal("d is not asked for a creature they don't control")
	}
	effectstest.VoteFor(t, g, d.ID, "Bob Bear")
	effectstest.PassPriorityAroundTable(t, g)

	if secretRevealed(g) != 1 {
		t.Fatalf("revealed %d times, want once", secretRevealed(g))
	}
	for id, want := range map[uuid.UUID]bool{mineA: true, mineB: false, bobs: false, cys: false} {
		if secretOnBattlefield(g, id) != want {
			t.Errorf("creature %s on the battlefield = %v, want %v", id, !want, want)
		}
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventVotesFinished {
			t.Error("a secret choice is not a vote: nobody finishes voting")
		}
	}
}

// --- The Toymaker's Trap -----------------------------------------------

func TestToymakersTrapWrongGuessCostsTheGuessAndDrawsThenTheNumberIsUsed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trap := effectstest.B12Permanent(g, me.ID, "The Toymaker's Trap", "Enchantment")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, trap).OracleID = secretToymaker })
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// The controller names the guesser.
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil || pick.CouncilVote != nil {
		t.Fatalf("no choose-an-opponent prompt: %+v", g.PendingChoices)
	}
	for i, o := range pick.PickOptions {
		if o.Player == opp.ID {
			effectstest.AnswerOptionPick(t, g, me.ID, i)
		}
	}
	effectstest.VoteFor(t, g, me.ID, "4")
	if got := secretWireBallots(t, g, opp.ID.String()); len(got) != 0 {
		t.Fatalf("the guesser sees %v before guessing", got)
	}
	hand, life := me.Hand.Size(), opp.Life
	effectstest.VoteFor(t, g, opp.ID, "2")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("a wrong guess of 2: life %d → %d", life, opp.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("a wrong guess draws the controller a card: hand %d → %d", hand, me.Hand.Size())
	}
	if !secretOnBattlefield(g, trap) {
		t.Fatal("a wrong guess sacrificed the Trap")
	}
	var notes []string
	g.ReadSnapshot(func() { notes = g.NotesForEffect(trap) })
	if !slices.Equal(notes, []string{"4"}) {
		t.Fatalf("noted %v, want [4]", notes)
	}

	// Next upkeep: 4 is no longer on offer, and a right guess sacrifices it.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	var offered []string
	for _, o := range effectstest.BallotFor(g, me.ID).PickOptions {
		offered = append(offered, o.Label)
	}
	if !slices.Equal(offered, []string{"1", "2", "3", "5"}) {
		t.Fatalf("offered %v, want 1, 2, 3 and 5", offered)
	}
	effectstest.VoteFor(t, g, me.ID, "5")
	var guesser uuid.UUID
	for _, s := range g.Seats[1:] {
		if effectstest.BallotFor(g, s.ID) != nil {
			guesser = s.ID
		}
	}
	if guesser == uuid.Nil {
		t.Fatal("nobody is asked to guess")
	}
	if n := len(effectstest.BallotFor(g, guesser).PickOptions); n != 5 {
		t.Errorf("the guess offers %d numbers, want 1 to 5", n)
	}
	effectstest.VoteFor(t, g, guesser, "5")
	effectstest.PassPriorityAroundTable(t, g)
	if secretOnBattlefield(g, trap) {
		t.Fatal("a right guess did not sacrifice the Trap")
	}
}

func TestToymakersTrapDoesNothingOnceAllFiveAreChosen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trap := effectstest.B12Permanent(g, me.ID, "The Toymaker's Trap", "Enchantment")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, trap)
		c.OracleID = secretToymaker
		c.Notes = []string{"1", "2", "3", "4", "5"}
	})
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("with every number chosen something was asked: %+v", g.PendingChoices[0])
	}
}

// --- Gollum, Scheming Guide --------------------------------------------

func gollumBoard(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gollum := effectstest.B12Push(g, me.ID, "Gollum, Scheming Guide", "Legendary Creature — Halfling Horror", secretGollum, 2, 1)
	land := secretTop(me, game.Card{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	secretTop(me, game.Card{InstanceID: uuid.New(), Name: "Top Spell", TypeLine: "Instant", ManaCost: "{B}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(gollum, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// Look at the top two and put the land on top.
	var order *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePutInLibrary {
			order = c
		}
	}
	if order == nil {
		t.Fatalf("no look-and-reorder prompt: %+v", g.PendingChoices)
	}
	rest := []uuid.UUID{land}
	for _, id := range order.ScryCards {
		if id != land {
			rest = append(rest, id)
		}
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, rest); err != nil {
		t.Fatalf("reorder %v (land %v) placement %v top %d: %v", order.ScryCards, land, order.LibraryPlacement, order.LibraryTopCount, err)
	}
	for i, o := range effectstest.LatestOptionPickFor(g, me.ID).PickOptions {
		if o.Player == opp.ID {
			effectstest.AnswerOptionPick(t, g, me.ID, i)
		}
	}
	return g, me, opp, gollum, land
}

func TestGollumRightGuessRemovesItFromCombat(t *testing.T) {
	g, me, opp, gollum, _ := gollumBoard(t)
	effectstest.VoteFor(t, g, me.ID, "Land")
	if got := secretWireBallots(t, g, opp.ID.String()); len(got) != 0 {
		t.Fatalf("the guesser sees the chosen kind: %v", got)
	}
	effectstest.VoteFor(t, g, opp.ID, "It is the chosen kind")
	effectstest.PassPriorityAroundTable(t, g)
	var attacking bool
	g.ReadSnapshot(func() { attacking = effectstest.FindBattlefieldCardForTest(g, gollum).AttackingTarget != uuid.Nil })
	if attacking {
		t.Error("a right guess left Gollum attacking")
	}
}

func TestGollumWrongGuessDrawsAndMakesItUnblockable(t *testing.T) {
	g, me, opp, gollum, land := gollumBoard(t)
	effectstest.VoteFor(t, g, me.ID, "Nonland")
	hand := me.Hand.Size()
	effectstest.VoteFor(t, g, opp.ID, "It is the chosen kind")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(land) {
		t.Errorf("a wrong guess draws the revealed card: hand %d → %d", hand, me.Hand.Size())
	}
	var attacking bool
	g.ReadSnapshot(func() {
		attacking = effectstest.FindBattlefieldCardForTest(g, gollum).AttackingTarget != uuid.Nil
	})
	restricted := effectstest.RestrictionsOf(t, g, gollum).Has(game.CantBeBlocked)
	if !attacking {
		t.Error("a wrong guess removed Gollum from combat")
	}
	if !restricted {
		t.Error("a wrong guess left Gollum blockable")
	}
}

// --- Itazura, Lingering Wick ------------------------------------------

func TestItazuraHighestNumberTakesTheDamageAndMayCastThenYouTakeACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.B12Push(g, me.ID, "Itazura, Lingering Wick", "Legendary Creature — Spirit", secretItazura, 1, 5)
	top := make([]uuid.UUID, 3)
	for i := range top {
		top[i] = secretTop(me, game.Card{InstanceID: uuid.New(), Name: "Exiled Land", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range top {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Fatalf("card %s was not exiled (zone %+v); pending %+v; stack %d", id, z, g.PendingChoices, len(g.Stack.Cards))
		}
	}
	// Each opponent in turn order names a number, hidden from the rest.
	pa := effectstest.AnswerNumber(t, g, b.ID, 3)
	if !pa.NoMax || pa.Min != 0 {
		t.Errorf("the number is 0 or greater with no ceiling: %+v", pa)
	}
	for _, viewer := range []string{c.ID.String(), d.ID.String(), me.ID.String(), protocol.SpectatorViewerID, ""} {
		if got := secretWireBallots(t, g, viewer); len(got) != 0 {
			t.Errorf("viewer %q sees %v before the reveal", viewer, got)
		}
	}
	if got := secretWireBallots(t, g, b.ID.String()); len(got) != 1 || got[0].Number != 3 {
		t.Errorf("b sees their own number as %v", got)
	}
	effectstest.AnswerNumber(t, g, c.ID, 5)
	lifeC, lifeD := c.Life, d.Life
	effectstest.AnswerNumber(t, g, d.ID, 0)
	if c.Life != lifeC-5 || d.Life != lifeD {
		t.Errorf("the highest number deals that much: c %d → %d, d %d → %d", lifeC, c.Life, lifeD, d.Life)
	}
	// Three lands: nothing to cast, so no offer; the controller takes one.
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no card for the hand: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top[1]) {
		t.Error("the chosen card is not in the controller's hand")
	}
	for _, id := range []uuid.UUID{top[0], top[2]} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("an unchosen card left exile")
		}
	}
}

func TestItazuraTiedOpponentIsChosenAndMayCastASpellFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, b, c, d := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.B12Push(g, me.ID, "Itazura, Lingering Wick", "Legendary Creature — Spirit", secretItazura, 1, 5)
	spell := secretTop(me, game.Card{InstanceID: uuid.New(), Name: "Exiled Bear", TypeLine: secretBearTL, ManaCost: "{5}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	landA := secretTop(me, game.Card{InstanceID: uuid.New(), Name: "Exiled Land", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	landB := secretTop(me, game.Card{InstanceID: uuid.New(), Name: "Exiled Land", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, b.ID, 2)
	effectstest.AnswerNumber(t, g, c.ID, 2)
	lifeB, lifeC := b.Life, c.Life
	effectstest.AnswerNumber(t, g, d.ID, 1)
	// b and c tie: the controller chooses c.
	tie := effectstest.LatestOptionPickFor(g, me.ID)
	if tie == nil || len(tie.PickOptions) != 2 {
		t.Fatalf("no choice between the tied opponents: %+v", g.PendingChoices)
	}
	for i, o := range tie.PickOptions {
		if o.Player == c.ID {
			effectstest.AnswerOptionPick(t, g, me.ID, i)
		}
	}
	if c.Life != lifeC-2 || b.Life != lifeB {
		t.Errorf("damage: b %d → %d, c %d → %d", lifeB, b.Life, lifeC, c.Life)
	}
	offer := effectstest.ResolutionCastFor(g, c.ID)
	if offer == nil {
		t.Fatalf("the chosen opponent is not offered the cast: %+v", g.PendingChoices)
	}
	if err := g.CastForResolutionChoice(c.ID, offer.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("c casts the exiled spell free: %v", err)
	}
	if effectstest.ResolutionCastFor(g, c.ID) != nil {
		effectstest.DeclineResolutionCasts(t, g, c.ID)
	}
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatalf("no card for the hand: %+v", g.PendingChoices)
	}
	if slices.Contains(pick.ChooseCards, spell) {
		t.Error("the cast card is offered for the hand")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{landB}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(landB) {
		t.Error("the chosen land is not in hand")
	}
	if z := g.FindCardZoneForEffect(landA); z == nil || z.Kind != game.ZoneExile {
		t.Error("the other land left exile")
	}
	var bear game.Card
	var onField bool
	g.ReadSnapshot(func() {
		if c2 := effectstest.FindBattlefieldCardForTest(g, spell); c2 != nil {
			bear, onField = *c2, true
		}
	})
	if !onField || bear.Controller != c.ID {
		t.Errorf("the free spell did not resolve under c's control (on field %v)", onField)
	}
}

// --- Duelists' Convocation International -------------------------------

func TestDCIWritesDownATenDigitNumberAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastCatalogSpell(t, g, "Duelists' Convocation International", "Enchantment", secretDCI, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var notes []string
	g.ReadSnapshot(func() { notes = g.NotesForEffect(id) })
	if len(notes) != 1 || len(notes[0]) != len("Number ")+10 || notes[0][:7] != "Number " || notes[0][7] == '0' {
		t.Fatalf("wrote down %v, want one 10-digit number", notes)
	}
	var wire []string
	g.ReadSnapshot(func() {
		for _, z := range protocol.ViewOfGame(g).Battlefield.Cards {
			if z.InstanceID == id.String() {
				wire = z.Notes
			}
		}
	})
	if !slices.Equal(wire, notes) {
		t.Errorf("the table sees %v on the card, want %v", wire, notes)
	}
}

func TestDCICrossesOutAMatchingDigitDrawsAndWinsOnTheTenth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dci := effectstest.B12Permanent(g, me.ID, "Duelists' Convocation International", "Enchantment")
	notes := []string{"Number 3141592653"}
	for place := 1; place <= 9; place++ {
		notes = append(notes, "Crossed out the "+string("314159265"[place-1])+" in place "+itoa(place))
	}
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, dci)
		c.OracleID = secretDCI
		c.Notes = notes
	})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	// A spell of mana value 7 shares no uncrossed digit: no trigger.
	seven := effectstest.ImportToHand(secretRow("Seven", "{7}"), me)
	if err := g.CastSpell(me.ID, seven, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	if len(g.Stack.Cards) != 1 {
		t.Fatalf("a mana value 7 spell triggered: stack %d", len(g.Stack.Cards))
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A mana value 3 spell crosses out the last 3, draws, and wins.
	hand := me.Hand.Size()
	three := effectstest.ImportToHand(secretRow("Three", "{3}"), me)
	if err := g.CastSpell(me.ID, three, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8 && g.State == game.StateActive && len(g.Stack.Cards) > 0; i++ {
		if err := g.PassPriority(); err != nil && g.State == game.StateActive {
			t.Fatal(err)
		}
	}
	if g.State == game.StateActive {
		t.Fatal("all ten digits are crossed out but the game goes on")
	}
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("the outcome is %+v, want the controller winning", g.Outcome)
	}
	if me.Hand.Size() < hand {
		t.Errorf("no card was drawn: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestDCILandPlayCrossesOutAZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dci := effectstest.B12Permanent(g, me.ID, "Duelists' Convocation International", "Enchantment")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, dci)
		c.OracleID = secretDCI
		c.Notes = []string{"Number 1000000002"}
	})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	effectstest.PlayLandFromHand(t, g, "Plains", "")
	effectstest.PassPriorityAroundTable(t, g)
	var got []string
	g.ReadSnapshot(func() { got = g.NotesForEffect(dci) })
	if !slices.Equal(got, []string{"Number 1000000002", "Crossed out the 0 in place 2"}) {
		t.Errorf("notes %v", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the land play drew a card: hand %d → %d", hand, me.Hand.Size())
	}
	if g.State != game.StateActive {
		t.Error("one digit crossed out won the game")
	}
}

func itoa(n int) string { return strconv.Itoa(n) }

// secretRow is a vanilla artifact card with the given mana cost.
func secretRow(name, cost string) cards.Card {
	return cards.Card{ID: uuid.New(), OracleID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: cost}
}

// secretTop puts c on top of p's library and returns its ID.
func secretTop(p *game.Player, c game.Card) uuid.UUID {
	p.Library.PushTop(c)
	return c.InstanceID
}
