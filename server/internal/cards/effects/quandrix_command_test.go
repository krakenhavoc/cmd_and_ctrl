package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// quandrix_command_test.go — #4862. Quandrix Command's fourth bullet is a
// player clause and a dependent clause of up to three cards that must be
// in that player's graveyard (FromTargetPlayersGraveyard).

const quandrixCommandOracle = "9c0f7cdc-9095-40ae-bf83-75a5ffda72b9"

// The bullets, by printed index.
const (
	qcBounce = iota
	qcCounter
	qcCounters
	qcShuffle
)

// qcShuffleRefs is the fourth bullet's group at occurrence occ: the
// player, then the cards.
func qcShuffleRefs(occ int, player uuid.UUID, cards ...uuid.UUID) []game.TargetRef {
	out := []game.TargetRef{modeRef(game.TargetPlayer, player, occ, 0)}
	for _, id := range cards {
		out = append(out, modeRef(game.TargetCard, id, occ, 1))
	}
	return out
}

// qcGraveyard puts n named cards into p's graveyard.
func qcGraveyard(p *game.Player, prefix string, n int) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = pushGraveyardCardForTest(p, prefix+" "+string(rune('A'+i)))
	}
	return out
}

// qcCastWithCounters casts the counters bullet (occurrence 0) on bear and
// the shuffle bullet (occurrence 1) with the given player and cards.
func qcCastWithCounters(t *testing.T, g *game.Game, bear, player uuid.UUID, cards ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	targets := append([]game.TargetRef{modeRef(game.TargetCard, bear, 0, 0)}, qcShuffleRefs(1, player, cards...)...)
	return castModalErr(t, g, "Quandrix Command", "Instant", quandrixCommandOracle, []int{qcCounters, qcShuffle}, targets)
}

func qcPlusOnes(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterPlusOne]
}

func TestQuandrixCommandBouncesAndCountersAnArtifactSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: rock, Name: "Mind Stone", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	castModal(t, g, "Quandrix Command", "Instant", quandrixCommandOracle, []int{qcBounce, qcCounter},
		[]game.TargetRef{modeRef(game.TargetCard, bear, 0, 0), modeRef(game.TargetCard, rock, 1, 0)})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if g.Battlefield.Contains(rock) || !me.Graveyard.Contains(rock) {
		t.Error("the artifact spell was not countered")
	}
}

func TestQuandrixCommandRefusesACreatureSpellForTheCounterBullet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	_, err := castModalErr(t, g, "Quandrix Command", "Instant", quandrixCommandOracle, []int{qcCounter, qcCounters},
		[]game.TargetRef{modeRef(game.TargetCard, spell, 0, 0), modeRef(game.TargetCard, bear, 1, 0)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature spell for the counter bullet: err %v, want ErrIllegalTarget", err)
	}
}

// Mode 4 with three picks, and the counters bullet beside it.
func TestQuandrixCommandShufflesThreeCardsAndPutsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dead := qcGraveyard(opp, "Dead", 4)
	library := opp.Library.Size()
	shuffles := shuffleCount(g, opp.ID)

	if _, err := qcCastWithCounters(t, g, bear, opp.ID, dead[0], dead[1], dead[2]); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := qcPlusOnes(g, bear); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	for _, id := range dead[:3] {
		if opp.Graveyard.Contains(id) || !opp.Library.Contains(id) {
			t.Errorf("picked card %s was not shuffled into its owner's library", id)
		}
	}
	if !opp.Graveyard.Contains(dead[3]) {
		t.Error("the unpicked card left the graveyard")
	}
	if got := opp.Library.Size(); got != library+3 {
		t.Errorf("library size %d, want %d", got, library+3)
	}
	if got := shuffleCount(g, opp.ID); got != shuffles+1 {
		t.Errorf("the target player shuffled %d times, want once", got-shuffles)
	}
}

// One pick, from the caster's own graveyard.
func TestQuandrixCommandShufflesOneCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dead := qcGraveyard(me, "Mine", 2)
	if _, err := qcCastWithCounters(t, g, bear, me.ID, dead[1]); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Library.Contains(dead[1]) || me.Graveyard.Contains(dead[1]) {
		t.Error("the picked card was not shuffled in")
	}
	if !me.Graveyard.Contains(dead[0]) {
		t.Error("the unpicked card left the graveyard")
	}
}

// "Up to three": no card at all is a legal announcement, and the player
// still shuffles (the Quandrix Command ruling).
func TestQuandrixCommandWithNoCardsStillShuffles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dead := qcGraveyard(opp, "Dead", 1)
	shuffles := shuffleCount(g, opp.ID)
	if _, err := qcCastWithCounters(t, g, bear, opp.ID); err != nil {
		t.Fatalf("cast with no cards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := shuffleCount(g, opp.ID); got != shuffles+1 {
		t.Errorf("the target player shuffled %d times, want once", got-shuffles)
	}
	if !opp.Graveyard.Contains(dead[0]) {
		t.Error("an unpicked card left the graveyard")
	}
	if got := qcPlusOnes(g, bear); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

// CR 601.2c: every card must be in the target player's graveyard, and
// there are at most three.
func TestQuandrixCommandRefusesAnUnrelatedOrFourthCard(t *testing.T) {
	t.Run("a card from another player's graveyard", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := qcGraveyard(opp, "Theirs", 1)
		elsewhere := qcGraveyard(other, "Elsewhere", 1)
		if _, err := qcCastWithCounters(t, g, bear, opp.ID, theirs[0], elsewhere[0]); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("a card from a different graveyard: err %v, want ErrIllegalTarget", err)
		}
		if _, err := qcCastWithCounters(t, g, bear, opp.ID, elsewhere[0]); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("only a card from a different graveyard: err %v, want ErrIllegalTarget", err)
		}
	})
	t.Run("a fourth card", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		dead := qcGraveyard(opp, "Dead", 4)
		if _, err := qcCastWithCounters(t, g, bear, opp.ID, dead...); err == nil {
			t.Fatal("a fourth card was accepted")
		}
	})
	t.Run("a card on the battlefield", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if _, err := qcCastWithCounters(t, g, bear, opp.ID, theirs); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("a permanent for the card clause: err %v, want ErrIllegalTarget", err)
		}
	})
}

// CR 608.2b: each card is judged again on its own as the spell resolves.
// A card that has left the graveyard stays where it went, and the rest
// are still shuffled in.
func TestQuandrixCommandSkipsCardsThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dead := qcGraveyard(opp, "Dead", 3)
	if _, err := qcCastWithCounters(t, g, bear, opp.ID, dead...); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// In response, two of the cards are exiled.
	g.WithWriteLock(func() {
		for _, id := range dead[:2] {
			c, err := opp.Graveyard.Remove(id)
			if err != nil {
				t.Fatal(err)
			}
			g.Exile.PushTop(c)
		}
	})
	passPriorityAroundTable(t, g)
	for _, id := range dead[:2] {
		if !g.Exile.Contains(id) || opp.Library.Contains(id) {
			t.Errorf("card %s left the graveyard and was still moved", id)
		}
	}
	if !opp.Library.Contains(dead[2]) {
		t.Error("the card still in the graveyard was not shuffled in")
	}
}

// The rulings: with every picked card gone but the player still legal,
// the player still shuffles; with the player an illegal target (here,
// hexproof from Leyline of Sanctity), nothing is shuffled even though the
// cards are still legal targets (Memory's Journey's ruling).
func TestQuandrixCommandReadsThePlayersLegality(t *testing.T) {
	t.Run("every card gone, the player legal", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		dead := qcGraveyard(opp, "Dead", 1)
		if _, err := qcCastWithCounters(t, g, bear, opp.ID, dead[0]); err != nil {
			t.Fatalf("cast: %v", err)
		}
		shuffles := shuffleCount(g, opp.ID)
		g.WithWriteLock(func() {
			c, err := opp.Graveyard.Remove(dead[0])
			if err != nil {
				t.Fatal(err)
			}
			g.Exile.PushTop(c)
		})
		passPriorityAroundTable(t, g)
		if got := shuffleCount(g, opp.ID); got != shuffles+1 {
			t.Errorf("the target player shuffled %d times, want once", got-shuffles)
		}
	})
	t.Run("the player becomes hexproof", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
		dead := qcGraveyard(opp, "Dead", 2)
		if _, err := qcCastWithCounters(t, g, bear, opp.ID, dead...); err != nil {
			t.Fatalf("cast: %v", err)
		}
		shuffles := shuffleCount(g, opp.ID)
		pushCatalogPermanent(g, opp.ID, "Leyline of Sanctity", "Enchantment", "492e0e6c-8c27-4376-938b-f8a8b6205810", false)
		passPriorityAroundTable(t, g)
		for _, id := range dead {
			if !opp.Graveyard.Contains(id) {
				t.Errorf("card %s moved though its player is an illegal target", id)
			}
		}
		if got := shuffleCount(g, opp.ID); got != shuffles {
			t.Errorf("an illegal player shuffled %d times", got-shuffles)
		}
		if got := qcPlusOnes(g, bear); got != 2 {
			t.Errorf("+1/+1 counters = %d, want 2: the other bullet still happens", got)
		}
	})
}

// The picker's pairs: for each player, the cards in that player's
// graveyard, and the dependent clause allows up to three.
func TestQuandrixCommandBotEnumerationRespectsTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := qcGraveyard(opp, "Theirs", 2)
	mine := qcGraveyard(me, "Mine", 1)
	ms := game.ModeSpecFor(quandrixCommandOracle)
	if ms == nil {
		t.Fatal("no mode spec")
	}
	spec := ms.Options[qcShuffle].Targets
	if c := spec.Clause(1); c == nil || c.Min != 0 || c.Max != 3 || c.Relation == nil || c.Relation.Kind != game.TargetInTargetPlayersGraveyard {
		t.Fatalf("the card clause is %+v, want 0..3 from the target player's graveyard", c)
	}
	var pairs map[uuid.UUID][]uuid.UUID
	g.ReadSnapshot(func() {
		pairs = g.TargetRelationPairsForEffect(game.SourceChooser(me.ID), spec, 1)
	})
	if !qcSameIDSet(pairs[opp.ID], theirs) {
		t.Errorf("opponent's pairs %v, want %v", pairs[opp.ID], theirs)
	}
	if !qcSameIDSet(pairs[me.ID], mine) {
		t.Errorf("caster's pairs %v, want %v", pairs[me.ID], mine)
	}
}

func qcSameIDSet(got, want []uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[uuid.UUID]bool, len(got))
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			return false
		}
	}
	return true
}

// CR 707.10c: a copy may change the player, and then its cards must
// come from the new player's graveyard; each offered answer is accepted.
func TestQuandrixCommandCopyCanMoveToAnotherPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := qcGraveyard(opp, "Theirs", 2)
	elsewhere := qcGraveyard(other, "Elsewhere", 3)
	spell, err := qcCastWithCounters(t, g, bear, opp.ID, theirs...)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	copySpellNow(t, g, spell, me.ID)

	// The bullets' clauses are asked in order: the creature, the player,
	// then the cards.
	answer := func(want string, targets []game.TargetRef) {
		t.Helper()
		c := copyPromptFor(g, me.ID)
		if c == nil {
			t.Fatalf("no copy prompt open for %s", want)
		}
		if err := g.ResolvePickTargets(c.ID, me.ID, targets); err != nil {
			t.Fatalf("answer %s (%q): %v", want, c.Reason, err)
		}
	}
	c := copyPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no copy prompt")
	}
	if strings.Contains(c.Reason, "creature") {
		answer("the creature", []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	}
	answer("the player", []game.TargetRef{{Kind: game.TargetPlayer, ID: other.ID}})
	c = copyPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("changing the player opened no prompt for the cards")
	}
	for _, id := range elsewhere[:2] {
		found := false
		for _, offered := range c.PickTargetCards {
			found = found || offered == id
		}
		if !found {
			t.Errorf("the new player's card %s was not offered", id)
		}
	}
	// Keeping the old player's cards with the new player is refused.
	if err := g.ResolvePickTargets(c.ID, me.ID, cardTargets(theirs...)); err == nil {
		t.Fatal("the copy kept cards from a graveyard that is no longer its player's")
	}
	answer("the cards", cardTargets(elsewhere[0], elsewhere[1]))
	if copyPromptFor(g, me.ID) != nil {
		t.Fatal("a copy prompt is still open")
	}
	passPriorityAroundTable(t, g)

	for _, id := range elsewhere[:2] {
		if !other.Library.Contains(id) {
			t.Errorf("the copy did not shuffle %s into the new player's library", id)
		}
	}
	if other.Library.Contains(elsewhere[2]) {
		t.Error("the copy moved an unpicked card")
	}
	for _, id := range theirs {
		if !opp.Library.Contains(id) {
			t.Errorf("the original did not shuffle %s in", id)
		}
	}
	if got := qcPlusOnes(g, bear); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4 (the original and the copy)", got)
	}
}

// Every answer the bot enumerator offers for the copy's prompts is
// accepted, down every branch (the #2622 harness).
func TestQuandrixCommandCopyPromptsAlwaysHaveAnAcceptedAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := qcGraveyard(opp, "Theirs", 2)
	elsewhere := qcGraveyard(other, "Elsewhere", 2)
	spell, err := qcCastWithCounters(t, g, bear, opp.ID, theirs...)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	copySpellNow(t, g, spell, me.ID)
	finals := walkCopyPrompt(t, g, me.ID, 1, "quandrix")
	if !hasFinal(finals, bear, opp.ID, theirs[0], theirs[1]) {
		t.Error("leaving every target unchanged was not offered (CR 707.10c)")
	}
	reached := false
	for _, f := range finals {
		if len(f) == 4 && f[1].ID == other.ID && qcSameIDSet([]uuid.UUID{f[2].ID, f[3].ID}, elsewhere) {
			reached = true
		}
	}
	if !reached {
		t.Error("the copy could not move to the other player and both of their cards")
	}
	for _, f := range finals {
		if len(f) == 4 && f[1].ID == other.ID && (f[2].ID == theirs[0] || f[2].ID == theirs[1] || f[3].ID == theirs[0] || f[3].ID == theirs[1]) {
			t.Errorf("a final set %v keeps the old player's card with the new player", f)
		}
	}
}
