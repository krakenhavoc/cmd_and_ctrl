package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// curse_of_leeches_test.go — Curse of Leeches // Leeching Lurker
// (#2586): an Aura Curse that enchants a player, turns into a creature
// at night, and is attached to a player again, by its controller's
// choice, when it turns back over (ADR 0079's amendment of 2026-10-09).

const curseOfLeechesOracle = "44eb0caa-ba16-49e4-915c-bd5e1ce770e6"

func curseOfLeechesRow() cards.Card {
	return werewolfRow(curseOfLeechesOracle,
		"Curse of Leeches", "Enchantment — Aura Curse", "{2}{B}",
		"Enchant player\nAs this permanent transforms into Curse of Leeches, attach it to a player.\nAt the beginning of enchanted player's upkeep, they lose 1 life and you gain 1 life.\n"+dnDay,
		"Leeching Lurker", "Creature — Leech Horror",
		"Lifelink\n"+dnNight,
		"", "", "4", "4", []string{"B"}, "Lifelink", "Enchant", "Transform")
}

// leechesCast casts the Curse at `target` as an Aura spell and lets it
// resolve. It is day (the table's designation starts as neither and
// the daybound entry makes it day), so the front face stays up.
func leechesCast(t *testing.T, g *game.Game, caster *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := importToHand(curseOfLeechesRow(), caster)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
	if err != nil {
		t.Fatalf("CastSpell Curse of Leeches: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the Curse did not resolve onto the battlefield")
	}
	return id
}

func leechesHost(t *testing.T, g *game.Game, id uuid.UUID) game.TargetRef {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return c.AttachedTo
}

func leechesInGraveyard(g *game.Game, owner *game.Player, id uuid.UUID) bool {
	_, ok := cardInZone(owner.Graveyard, id)
	return ok
}

// answerAttachPrompt answers the attach question by picking `pick`.
func answerAttachPrompt(t *testing.T, g *game.Game, chooser *game.Player, pick *game.Player) {
	t.Helper()
	answerChoosePlayer(t, g, chooser.ID, pick)
}

func TestCurseOfLeechesCastByDayEnchantsItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := leechesCast(t, g, me, opp.ID)
	if host := leechesHost(t, g, id); host.Kind != game.TargetPlayer || host.ID != opp.ID {
		t.Fatalf("attached to %+v, want player %s", host, opp.ID)
	}
	if c, _ := battlefieldCard(g, id); c.ActiveFace != 0 || c.Name != "Curse of Leeches" {
		t.Errorf("by day it is %q on face %d, want the Curse", c.Name, c.ActiveFace)
	}
	if !slices.Contains(effectiveAbilities(t, g, id), "daybound") {
		t.Errorf("the Curse lacks daybound: %v", effectiveAbilities(t, g, id))
	}
}

// leechesCastASpell has `seat` cast one spell in its own main phase, so
// the CR 502.2 check at the next untap step sees a spell and the table
// stays at day.
func leechesCastASpell(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	advanceToMainOf(t, g, seat)
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
}

func TestCurseOfLeechesDrainsOnlyTheEnchantedPlayersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	leechesCast(t, g, me, opp.ID)

	myLife, oppLife, otherLife := me.Life, opp.Life, other.Life
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("enchanted player at %d, want %d", opp.Life, oppLife-1)
	}
	if me.Life != myLife+1 {
		t.Errorf("controller at %d, want %d", me.Life, myLife+1)
	}
	leechesCastASpell(t, g, 1)
	advanceToUpkeepOfSeat(t, g, 2)
	passPriorityAroundTable(t, g)
	if other.Life != otherLife || me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Errorf("a bystander's upkeep moved life: me %d opp %d other %d", me.Life, opp.Life, other.Life)
	}
}

// Night turns the Aura into the creature, and a creature is attached to
// nothing (CR 704.5p): it lets go of its player and stays on the
// battlefield, a 4/4 lifelinker.
func TestLeechingLurkerLetsGoOfThePlayerAtNight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := leechesCast(t, g, me, opp.ID)
	dnaNight(g)
	g.RunStateChecksForTest()

	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Leeching Lurker left the battlefield")
	}
	if c.ActiveFace != 1 || c.Name != "Leeching Lurker" {
		t.Fatalf("at night it is %q on face %d, want Leeching Lurker", c.Name, c.ActiveFace)
	}
	if c.IsAttached() {
		t.Errorf("the Lurker is still attached to %+v", c.AttachedTo)
	}
	if !c.IsCreature() {
		t.Error("the Lurker is not a creature")
	}
	if got := dnaPower(t, g, id); got != 4 {
		t.Errorf("power %d, want 4", got)
	}
	for _, k := range []string{"lifelink", "nightbound"} {
		if !slices.Contains(effectiveAbilities(t, g, id), k) {
			t.Errorf("the Lurker lacks %q: %v", k, effectiveAbilities(t, g, id))
		}
	}
}

// A Curse cast at night is a Lurker from the start: it enters
// transformed, so "as this transforms" never runs, and the player its
// spell targeted is not kept (CR 704.5p).
func TestLeechingLurkerCastAtNightEntersUnattachedWithNoQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dnaNight(g)
	id := leechesCast(t, g, me, opp.ID)
	g.RunStateChecksForTest()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the card did not stay on the battlefield")
	}
	if c.ActiveFace != 1 || c.IsAttached() {
		t.Errorf("face %d attached %v, want the unattached Lurker", c.ActiveFace, c.IsAttached())
	}
	if p := latestOptionPickFor(g, me.ID); p != nil {
		t.Errorf("a prompt is open for a permanent that never transformed: %+v", p)
	}
}

// Day returns: the Curse's controller chooses a player as it turns
// over, the Curse survives the state-based check while the question is
// open, and the answer attaches it. Any player, the controller too.
func TestCurseOfLeechesAsItTransformsAttachesToTheChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	id := leechesCast(t, g, me, opp.ID)
	dnaNight(g)
	g.RunStateChecksForTest()
	dnaDay(g)

	prompt := latestOptionPickFor(g, me.ID)
	if prompt == nil {
		t.Fatalf("no attach question for the controller: %+v", g.PendingChoices)
	}
	if got := len(prompt.PickOptions); got != len(g.Seats) {
		t.Errorf("%d players offered, want all %d (the controller included)", got, len(g.Seats))
	}
	// CR 704.5m must wait for the answer.
	g.RunStateChecksForTest()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the Curse was swept away while the question was open")
	}
	if c.Name != "Curse of Leeches" || c.IsAttached() {
		t.Fatalf("mid-question: %q attached %v", c.Name, c.IsAttached())
	}

	answerAttachPrompt(t, g, me, third)
	if host := leechesHost(t, g, id); host.Kind != game.TargetPlayer || host.ID != third.ID {
		t.Fatalf("attached to %+v, want the chosen player %s", host, third.ID)
	}
	g.RunStateChecksForTest()
	if _, ok := battlefieldCard(g, id); !ok {
		t.Error("the Curse left the battlefield after attaching legally")
	}
}

// The chosen player is whoever the controller names, not the player the
// spell first targeted, and the drain follows the new host.
func TestCurseOfLeechesDrainsTheNewlyChosenPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	leechesCast(t, g, me, opp.ID)
	dnaNight(g)
	dnaDay(g)
	answerAttachPrompt(t, g, me, third)

	myLife, oppLife, thirdLife := me.Life, opp.Life, third.Life
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != myLife {
		t.Fatalf("the old host's upkeep drained: opp %d me %d", opp.Life, me.Life)
	}
	leechesCastASpell(t, g, 1)
	advanceToUpkeepOfSeat(t, g, 2)
	passPriorityAroundTable(t, g)
	if third.Life != thirdLife-1 || me.Life != myLife+1 {
		t.Errorf("new host %d (want %d), controller %d (want %d)", third.Life, thirdLife-1, me.Life, myLife+1)
	}
}

// No player may be enchanted (each has protection from everything, CR
// 702.16j): nothing is offered, nothing is asked, and the Aura that is
// attached to nothing goes to its owner's graveyard (CR 704.5m).
func TestCurseOfLeechesWithNoLegalPlayerGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := leechesCast(t, g, me, opp.ID)
	dnaNight(g)
	g.RunStateChecksForTest()
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			d := g.UntilYourNextTurnDuration(p.ID)
			g.GrantPlayerStaticForEffect(p.ID, "protection from everything", "Test — protection from everything", uuid.Nil, d)
		}
	})
	dnaDay(g)
	if p := latestOptionPickFor(g, me.ID); p != nil {
		t.Fatalf("a question was asked with nobody to attach to: %+v", p)
	}
	g.RunStateChecksForTest()
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("an unattached Aura stayed on the battlefield")
	}
	if !leechesInGraveyard(g, me, id) {
		t.Error("the Curse is not in its owner's graveyard")
	}
}

// A protected seat is not offered; the others are.
func TestCurseOfLeechesDoesNotOfferAProtectedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	leechesCast(t, g, me, opp.ID)
	dnaNight(g)
	g.WithWriteLock(func() {
		d := g.UntilYourNextTurnDuration(third.ID)
		g.GrantPlayerStaticForEffect(third.ID, "protection from everything", "Test — protection from everything", uuid.Nil, d)
	})
	dnaDay(g)
	prompt := latestOptionPickFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no attach question")
	}
	for _, o := range prompt.PickOptions {
		if o.Player == third.ID {
			t.Errorf("the protected player %s is offered", third.ID)
		}
	}
	if len(prompt.PickOptions) != len(g.Seats)-1 {
		t.Errorf("%d players offered, want %d", len(prompt.PickOptions), len(g.Seats)-1)
	}
}

// A second Curse on the same turn over asks its own question, and each
// answer attaches its own Curse.
func TestTwoCursesTurningOverAskTwoQuestions(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := leechesCast(t, g, me, opp.ID)
	b := leechesCast(t, g, me, opp.ID)
	dnaNight(g)
	dnaDay(g)
	if n := countOptionPicksFor(g, me.ID); n != 2 {
		t.Fatalf("%d questions open, want 2", n)
	}
	answerAttachPrompt(t, g, me, third)
	answerAttachPrompt(t, g, me, fourth)
	hosts := []uuid.UUID{leechesHost(t, g, a).ID, leechesHost(t, g, b).ID}
	slices.SortFunc(hosts, func(x, y uuid.UUID) int { return slices.Compare(x[:], y[:]) })
	want := []uuid.UUID{third.ID, fourth.ID}
	slices.SortFunc(want, func(x, y uuid.UUID) int { return slices.Compare(x[:], y[:]) })
	if !slices.Equal(hosts, want) {
		t.Errorf("hosts %v, want %v", hosts, want)
	}
}

func countOptionPicksFor(g *game.Game, chooser uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			n++
		}
	}
	return n
}

func TestCurseOfLeechesBothFacesAreRegisteredWithTheirKeywords(t *testing.T) {
	front, ok := Lookup(curseOfLeechesOracle)
	if !ok || front.Name != "Curse of Leeches" {
		t.Fatalf("front face: %+v", front)
	}
	back, ok := Lookup(curseOfLeechesOracle + "#1")
	if !ok || back.Name != "Leeching Lurker" {
		t.Fatalf("back face: %+v", back)
	}
	if !slices.Contains(front.PrintedKeywords, "daybound") || !slices.Contains(back.PrintedKeywords, "nightbound") ||
		!slices.Contains(back.PrintedKeywords, "lifelink") {
		t.Errorf("keywords: front %v back %v", front.PrintedKeywords, back.PrintedKeywords)
	}
	if front.AsTransformsInto == nil {
		t.Error("the front face has no as-it-transforms clause")
	}
}
