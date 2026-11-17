package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dice_5681_tables_test.go — #5681: the four cards whose roll has a
// modifier and a results table (CR 706.2, 706.3): Revivify, Song of
// Inspiration, Danse Macabre and The Deck of Many Things.

const (
	revivifyOracle          = "55eb3c21-6834-4c62-8ab9-54b901531fed"
	songOfInspirationOracle = "bdb80d7b-672c-4e2a-b93b-9b96721b93f2"
	danseMacabreOracle      = "a9c58cfe-9630-48fa-84ad-50268eb725ed"
	deckOfManyThingsOracle  = "91df4cf5-d7ec-4fcd-87ed-e075ef6ceba9"
)

// d5681Table is a table at seat 0's main phase whose next d20 from
// `source` shows `natural`.
func d5681Table(t *testing.T, source uuid.UUID, natural int) *game.Game {
	t.Helper()
	key := d5681Key(t, source, 20, 1, func(r []int) bool { return r[0] == natural })
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(key)
	advanceTo(t, g, game.StepPrecombatMain)
	return g
}

// d5681Spell puts a spell with a chosen instance ID in the player's hand.
func d5681Spell(g *game.Game, p *game.Player, id uuid.UUID, name, typeLine, cost, oracle string) {
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID,
		})
	})
}

// d5681Dies moves a creature from the battlefield to its owner's
// graveyard.
func d5681Dies(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("to the graveyard: %v", err)
	}
}

// d5681OneRoll is the single die rolled after event index from.
func d5681OneRoll(t *testing.T, g *game.Game, from int) game.Event {
	t.Helper()
	got := d5681Rolled(g, from)
	if len(got) != 1 {
		t.Fatalf("rolled %+v, want one die", got)
	}
	return got[0]
}

// --- Revivify -----------------------------------------------------------

// Two creature cards died this turn: a 12 is a result of 14 (to hand), a
// 13 a result of 15 (to the battlefield). A creature card that was never
// on the battlefield is neither counted nor returned.
func TestRevivifyAddsTheDeadAndReadsTheTable(t *testing.T) {
	for _, tc := range []struct {
		natural     int
		battlefield bool
	}{{12, false}, {13, true}} {
		id := uuid.New()
		g := d5681Table(t, id, tc.natural)
		me := g.Seats[0]
		a := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
		b := pushVanillaCreature(g, me.ID, "Ox", 2, 4)
		d5681Dies(t, g, me.ID, a)
		d5681Dies(t, g, me.ID, b)
		milled := pushGraveyardCreature(g, me.ID, "Milled Bear", "{1}{G}")
		d5681Spell(g, me, id, "Revivify", "Instant", "{2}{W}", revivifyOracle)
		floatForTest(g, me, "WWW")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		from := sfkEventCount(g)
		passPriorityAroundTable(t, g)
		if die := d5681OneRoll(t, g, from); die.Amount != tc.natural || die.Modifier != 2 {
			t.Fatalf("rolled natural %d modifier %d, want %d and 2", die.Amount, die.Modifier, tc.natural)
		}
		want := game.ZoneHand
		if tc.battlefield {
			want = game.ZoneBattlefield
		}
		for _, c := range []uuid.UUID{a, b} {
			if !d5681InZone(g, c, want) {
				t.Errorf("natural %d: a creature card that died is not in %s", tc.natural, want)
			}
		}
		if !d5681InZone(g, milled, game.ZoneGraveyard) {
			t.Errorf("natural %d: a creature card that never died was returned", tc.natural)
		}
	}
}

// --- Song of Inspiration --------------------------------------------------

// Mana values 3 and 4: a 7 is a result of 14 (cards only), an 8 a result
// of 15 (cards and 7 life).
func TestSongOfInspirationAddsTheManaValues(t *testing.T) {
	for _, tc := range []struct {
		natural, life int
	}{{7, 0}, {8, 7}} {
		id := uuid.New()
		g := d5681Table(t, id, tc.natural)
		me := g.Seats[0]
		a := pushGraveyardCreature(g, me.ID, "Bear", "{2}{G}")
		b := pushGraveyardCreature(g, me.ID, "Golem", "{4}")
		d5681Spell(g, me, id, "Song of Inspiration", "Instant", "{3}{G}{G}", songOfInspirationOracle)
		floatForTest(g, me, "GGGGG")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{p2Card(a), p2Card(b)}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		life := me.Life
		from := sfkEventCount(g)
		passPriorityAroundTable(t, g)
		if die := d5681OneRoll(t, g, from); die.Amount != tc.natural || die.Modifier != 7 {
			t.Fatalf("rolled natural %d modifier %d, want %d and 7", die.Amount, die.Modifier, tc.natural)
		}
		if !d5681InZone(g, a, game.ZoneHand) || !d5681InZone(g, b, game.ZoneHand) {
			t.Errorf("natural %d: the targets did not return to hand", tc.natural)
		}
		if got := me.Life - life; got != tc.life {
			t.Errorf("natural %d: gained %d life, want %d", tc.natural, got, tc.life)
		}
	}
}

// The targets are permanent cards in your graveyard; none is a legal
// cast, and it still rolls.
func TestSongOfInspirationTargets(t *testing.T) {
	id := uuid.New()
	g := d5681Table(t, id, 20)
	me, opp := g.Seats[0], g.Seats[1]
	var spell uuid.UUID
	g.WithWriteLock(func() { spell = pushGraveyardCardTyped(me, "Shock", "Instant") })
	theirs := pushGraveyardCreature(g, opp.ID, "Their Bear", "{1}{G}")
	d5681Spell(g, me, id, "Song of Inspiration", "Instant", "{3}{G}{G}", songOfInspirationOracle)
	floatForTest(g, me, "GGGGG")
	for name, target := range map[string]uuid.UUID{"an instant card": spell, "a card in an opponent's graveyard": theirs} {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{p2Card(target)}}); err == nil {
			t.Fatalf("cast targeting %s", name)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with no targets: %v", err)
	}
	life := me.Life
	from := sfkEventCount(g)
	passPriorityAroundTable(t, g)
	if die := d5681OneRoll(t, g, from); die.Modifier != 0 {
		t.Errorf("modifier %d with no targets", die.Modifier)
	}
	if me.Life != life {
		t.Errorf("gained %d life with no targets", me.Life-life)
	}
}

// --- Danse Macabre --------------------------------------------------------

// danseSacrifices answers every seat's sacrifice leg with its pick.
func danseSacrifices(t *testing.T, g *game.Game, pick map[uuid.UUID]uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		c := pendingOfKind(g, game.PendingChoiceOwnPermanents)
		if c == nil {
			return
		}
		if err := g.ResolveOwnPermanents(c.ID, c.Chooser, []uuid.UUID{pick[c.Chooser]}); err != nil {
			t.Fatalf("sacrifice for %s: %v", c.Chooser, err)
		}
	}
	t.Fatal("the sacrifice prompts did not stop")
}

// danseScene casts Danse Macabre with the d20 showing natural: the caster
// sacrifices a 2/3 and seats 1 and 2 a creature each; seat 3 has only a
// token and sacrifices nothing.
func danseScene(t *testing.T, natural int) (g *game.Game, mine, theirs, third, token uuid.UUID, from int) {
	t.Helper()
	id := uuid.New()
	g = d5681Table(t, id, natural)
	me := g.Seats[0]
	mine = pushVanillaCreature(g, me.ID, "Ox", 2, 3)
	theirs = pushVanillaCreature(g, g.Seats[1].ID, "Dragon", 5, 5)
	third = pushVanillaCreature(g, g.Seats[2].ID, "Bear", 2, 2)
	g.WithWriteLock(func() {
		token = uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: token, Name: "Goblin", TypeLine: "Token Creature — Goblin",
			Power: 1, Toughness: 1, Owner: g.Seats[3].ID, Controller: g.Seats[3].ID,
		})
	})
	d5681Spell(g, me, id, "Danse Macabre", "Sorcery", "{3}{B}{B}", danseMacabreOracle)
	floatForTest(g, me, "BBBBB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	from = sfkEventCount(g)
	passPriorityAroundTable(t, g)
	danseSacrifices(t, g, map[uuid.UUID]uuid.UUID{me.ID: mine, g.Seats[1].ID: theirs, g.Seats[2].ID: third})
	passPriorityAroundTable(t, g)
	return g, mine, theirs, third, token, from
}

// An 11 plus the Ox's toughness of 3 is 14: one creature card, any
// player's, under the caster's control. The table survives a restore
// while the choice is open.
func TestDanseMacabreLowRowReturnsOne(t *testing.T) {
	g, mine, theirs, third, token, from := danseScene(t, 11)
	if die := d5681OneRoll(t, g, from); die.Amount != 11 || die.Modifier != 3 {
		t.Fatalf("rolled natural %d modifier %d, want 11 and the sacrificed creature's toughness, 3", die.Amount, die.Modifier)
	}
	if !d5681InZone(g, token, game.ZoneBattlefield) {
		t.Error("a token was sacrificed")
	}
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 3 {
		t.Fatalf("choice %+v, want the three sacrificed creature cards", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err == nil {
		t.Fatal("the 1—14 row returned nothing by choice")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{theirs, third}); err == nil {
		t.Fatal("the 1—14 row returned two cards")
	}
	answerChooseCards(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	var controller uuid.UUID
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(theirs); ok {
			controller = c.Controller
		}
	})
	if !d5681InZone(g, theirs, game.ZoneBattlefield) || controller != me.ID {
		t.Error("the chosen card did not enter under the caster's control")
	}
	if !d5681InZone(g, mine, game.ZoneGraveyard) || !d5681InZone(g, third, game.ZoneGraveyard) {
		t.Error("a card that was not chosen left its graveyard")
	}
}

// A 12 plus 3 is 15: up to two.
func TestDanseMacabreHighRowReturnsUpToTwo(t *testing.T) {
	g, mine, theirs, third, _, from := danseScene(t, 12)
	if die := d5681OneRoll(t, g, from); die.RollResult() != 15 {
		t.Fatalf("result %d, want 15", die.RollResult())
	}
	me := g.Seats[0]
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choice")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{mine, theirs, third}); err == nil {
		t.Fatal("the 15+ row returned three cards")
	}
	answerChooseCards(t, g, me.ID, mine, theirs)
	passPriorityAroundTable(t, g)
	if !d5681InZone(g, mine, game.ZoneBattlefield) || !d5681InZone(g, theirs, game.ZoneBattlefield) || !d5681InZone(g, third, game.ZoneGraveyard) {
		t.Error("the two chosen cards did not enter, or the third did")
	}
}

// --- The Deck of Many Things ----------------------------------------------

// deckScene is the Deck under seat 0's control with `hand` cards in hand
// and its next d20 showing natural.
func deckScene(t *testing.T, natural, hand int) (*game.Game, uuid.UUID) {
	t.Helper()
	deck := uuid.New()
	g := d5681Table(t, deck, natural)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: deck, Name: "The Deck of Many Things", TypeLine: "Legendary Artifact",
		OracleID: deckOfManyThingsOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[:hand] })
	return g, deck
}

func deckActivate(t *testing.T, g *game.Game, deck uuid.UUID) int {
	t.Helper()
	me := g.Seats[0]
	if err := g.ActivateCatalogAbility(me.ID, deck, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	if err := p2Activate(t, g, me, deck, 0, "CC"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	from := sfkEventCount(g)
	passPriorityAroundTable(t, g)
	return from
}

func TestTheDeckOfManyThingsTable(t *testing.T) {
	// 3 minus a hand of 3 is 0: discard the hand.
	g, deck := deckScene(t, 3, 3)
	me := g.Seats[0]
	from := deckActivate(t, g, deck)
	if die := d5681OneRoll(t, g, from); die.Amount != 3 || die.Modifier != -3 || die.RollResult() != 0 {
		t.Fatalf("rolled %+v, want a 3 minus 3", die)
	}
	if handSize(me) != 0 {
		t.Errorf("hand = %d after a result of 0, want it discarded", handSize(me))
	}

	// 4 minus 3 is 1: a card at random from the graveyard, here the only one.
	g, deck = deckScene(t, 4, 3)
	me = g.Seats[0]
	only := pushGraveyardCreature(g, me.ID, "Bear", "{1}{G}")
	deckActivate(t, g, deck)
	if !d5681InZone(g, only, game.ZoneHand) || handSize(me) != 4 {
		t.Errorf("a result of 1 did not return the graveyard's card (hand %d)", handSize(me))
	}

	// 12 minus 3 is 9: still the first row. 13 minus 3 is 10: draw two.
	g, deck = deckScene(t, 12, 3)
	deckActivate(t, g, deck)
	if got := handSize(g.Seats[0]); got != 3 {
		t.Errorf("hand = %d after a result of 9 with an empty graveyard, want 3", got)
	}
	g, deck = deckScene(t, 13, 3)
	deckActivate(t, g, deck)
	if got := handSize(g.Seats[0]); got != 5 {
		t.Errorf("hand = %d after a result of 10, want 5", got)
	}

	// A natural 20 minus 1 is 19: two cards, no creature, and Critical Hit
	// in the graveyard still sees the natural 20.
	g, deck = deckScene(t, 20, 1)
	me = g.Seats[0]
	crit := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{
			InstanceID: crit, Name: "Critical Hit", TypeLine: "Instant", ManaCost: "{1}{R}",
			OracleID: criticalHitOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	pushGraveyardCreature(g, me.ID, "Bear", "{1}{G}")
	deckActivate(t, g, deck)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a result of 19 offered a creature card")
	}
	if got := handSize(me); got != 4 || !d5681InZone(g, crit, game.ZoneHand) {
		t.Errorf("hand = %d, want 1 + 2 drawn + Critical Hit on the natural 20", got)
	}
}

// A result of 20: a creature card from any graveyard enters under your
// control, and when it dies, on a later turn and a restored table, its
// owner loses the game.
func TestTheDeckOfManyThingsTwenty(t *testing.T) {
	g, deck := deckScene(t, 20, 0)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushGraveyardCreature(g, opp.ID, "Their Dragon", "{4}{R}{R}")
	deckActivate(t, g, deck)
	answerChooseCards(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	var controller uuid.UUID
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(theirs); ok {
			controller = c.Controller
		}
	})
	if !d5681InZone(g, theirs, game.ZoneBattlefield) || controller != me.ID {
		t.Fatal("the creature card did not enter under the activator's control")
	}
	advanceToUpkeepOfSeat(t, g, 1)
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	d5681Dies(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Error("the creature's owner did not lose the game")
	}
	if me.Eliminated {
		t.Error("the Deck's controller lost the game")
	}
}
