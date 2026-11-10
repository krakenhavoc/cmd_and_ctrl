package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b63_cards_test.go — Lorwyn Eclipsed leftovers, batch 63 (#4458).
// Each test plays the real catalog entry and asserts what its text says.

const (
	stdB63RecluseOracle     = "ad99e430-e13f-4819-a2ee-bf3332ef517d"
	stdB63ShinestrikerOracl = "1b94ee87-a2f4-47f1-9fc7-4a167cad1b3b"
	stdB63PeddlerOracle     = "08688903-d9cc-4d57-ba91-14a8523551b9"
	stdB63SnareOracle       = "38877c68-b69b-4ee1-8e8f-5137a184a8f7"
	stdB63StratoOracle      = "36ab41b7-bcfb-4618-bbfb-70d608dba15a"
	stdB63SentinelOracle    = "e7b88da3-7dd1-45f8-b555-98ad838b2917"
	stdB63SunderflockOracle = "c01b221f-3993-43ab-8cde-b42fb344e2a7"
	stdB63SwatAwayOracle    = "79090b00-d920-4134-a838-fb434cf3ee4d"
	stdB63TanufelOracle     = "bf8de96b-0a75-4459-a123-dcf95e1b89f2"
	stdB63ThirstOracle      = "045ece1e-23aa-4122-8726-86b2af4345ba"
)

func TestStdB63EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		stdB63RecluseOracle:     "Rimekin Recluse",
		stdB63ShinestrikerOracl: "Shinestriker",
		stdB63PeddlerOracle:     "Silvergill Peddler",
		stdB63SnareOracle:       "Spell Snare",
		stdB63StratoOracle:      "Stratosoarer",
		stdB63SentinelOracle:    "Summit Sentinel",
		stdB63SunderflockOracle: "Sunderflock",
		stdB63SwatAwayOracle:    "Swat Away",
		stdB63TanufelOracle:     "Tanufel Rimespeaker",
		stdB63ThirstOracle:      "Thirst for Identity",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// stdB63Hand puts a catalog card into p's hand with a mana cost and colours.
func stdB63Hand(p *game.Player, name, typeLine, manaCost, oracle string, colors []string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleID: oracle,
		Colors: colors, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// stdB63CastCreature casts a creature from the active seat's hand and lets
// it settle, leaving any prompt open.
func stdB63CastCreature(t *testing.T, g *game.Game, name, typeLine, manaCost, oracle string, colors []string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := stdB63Hand(me, name, typeLine, manaCost, oracle, colors, power, toughness)
	toMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestStdB63RimekinRecluseBouncesAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	stdB63CastCreature(t, g, "Rimekin Recluse", "Creature — Elemental Wizard", "{2}{U}", stdB63RecluseOracle, []string{"U"}, 3, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("their creature was not returned to its owner's hand")
	}
}

func TestStdB63RimekinRecluseCannotTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rec := stdB63CastCreature(t, g, "Rimekin Recluse", "Creature — Elemental Wizard", "{2}{U}", stdB63RecluseOracle, []string{"U"}, 3, 2)
	if pick := latestPickTarget(g, me.ID); pick != nil && hasID(pick.PickTargetCards, rec) {
		t.Error("the Recluse offered itself as a target")
	}
	if !rfdOnBattlefield(g, rec) {
		t.Error("the Recluse left the battlefield")
	}
}

func TestStdB63ShinestrikerDrawsOnePerColorAmongPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b16Creature(g, me.ID, "Red Guy", "Creature — Goblin", 1, 1, "R")
	b16Creature(g, me.ID, "Red Twin", "Creature — Goblin", 1, 1, "R")
	b16Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Green", "Creature — Elf", 1, 1, "G")
	hand := me.Hand.Size()
	stdB63CastCreature(t, g, "Shinestriker", "Creature — Elemental", "{4}{U}{U}", stdB63ShinestrikerOracl, []string{"U"}, 3, 3)
	// Blue (itself) and red; an opponent's green permanent is not yours.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (blue and red)", got)
	}
}

func TestStdB63SilvergillPeddlerLootsWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	peddler := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Silvergill Peddler", TypeLine: "Creature — Merfolk Citizen",
		OracleID: stdB63PeddlerOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	tapForTest(t, g, peddler)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("owes %d discards, want 1", discardOwed(g, me.ID))
	}
}

func TestStdB63SpellSnareCountersOnlyManaValueTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Two Drop", lightningBoltOracle, "{1}{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Spell Snare", "Instant", stdB63SnareOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Fatal("the mana value 2 spell was not countered")
	}
	three := batch01OpponentCasts(t, g, opp, "Three Drop", lightningBoltOracle, "{1}{U}{U}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if !b03CastRefused(t, g, "Spell Snare", "Instant", stdB63SnareOracle, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: three}},
	}) {
		t.Error("a mana value 3 spell was accepted as a target")
	}
}

func TestStdB63StratosoarerGrantsFlyingAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	stdB63CastCreature(t, g, "Stratosoarer", "Creature — Elemental", "{4}{U}", stdB63StratoOracle, []string{"U"}, 3, 5)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target creature did not gain flying")
	}
	spec, _ := Lookup(stdB63StratoOracle)
	if len(spec.Activated) != 1 {
		t.Errorf("want basic landcycling as the one activated ability, got %d", len(spec.Activated))
	}
}

func TestStdB63SummitSentinelDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	s := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Summit Sentinel", TypeLine: "Creature — Elemental Soldier",
		OracleID: stdB63SentinelOracle, Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(s) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestStdB63SunderflockCostsTheGreatestElementalValueLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Creature — Elemental"
	if got := selfPricedMV(t, g, me, stdB63SunderflockOracle, tl, "{7}{U}{U}"); got != 9 {
		t.Errorf("no Elementals: %d, want 9", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Small", TypeLine: tl, ManaCost: "{1}{U}", Power: 1, Toughness: 1})
	pushPermanent(g, me.ID, game.Card{Name: "Big", TypeLine: tl, ManaCost: "{4}{U}", Power: 1, Toughness: 1})
	pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{8}", Power: 1, Toughness: 1})
	pushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: tl, ManaCost: "{7}", Power: 1, Toughness: 1})
	if got := selfPricedMV(t, g, me, stdB63SunderflockOracle, tl, "{7}{U}{U}"); got != 4 {
		t.Errorf("greatest Elemental value 5 (not the Bear's 8 or theirs): %d, want 4", got)
	}
}

func TestStdB63SunderflockBouncesNonElementalsOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ele := b12Creature(g, me.ID, "Spark", "Creature — Elemental", 1, 1)
	flock := stdB63CastCreature(t, g, "Sunderflock", "Creature — Elemental", "{7}{U}{U}", stdB63SunderflockOracle, []string{"U"}, 5, 5)
	if !opp.Hand.Contains(bear) || !me.Hand.Contains(mine) {
		t.Error("the non-Elementals were not returned")
	}
	if !rfdOnBattlefield(g, ele) || !rfdOnBattlefield(g, flock) {
		t.Error("an Elemental was returned")
	}
}

func TestStdB63SunderflockPutOntoTheBattlefieldBouncesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sunderflock", TypeLine: "Creature — Elemental",
			OracleID: stdB63SunderflockOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
		if err := g.ReturnFromGraveyardForEffect(id, game.ZoneBattlefield); err != nil {
			t.Errorf("put onto the battlefield: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !rfdOnBattlefield(g, bear) {
		t.Error("an uncast Sunderflock bounced a creature")
	}
}

func TestStdB63SwatAwayCostsTwoLessWhileAttackedAndOwnerTucksACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Instant"
	if got := selfPricedMV(t, g, me, stdB63SwatAwayOracle, tl, "{2}{U}{U}"); got != 4 {
		t.Errorf("not attacked: %d, want 4", got)
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Raider", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2, AttackingTarget: opp.ID})
	if got := selfPricedMV(t, g, me, stdB63SwatAwayOracle, tl, "{2}{U}{U}"); got != 4 {
		t.Errorf("a creature attacking somebody else: %d, want 4", got)
	}
	pushPermanent(g, opp.ID, game.Card{Name: "Raider Two", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2, AttackingTarget: me.ID})
	if got := selfPricedMV(t, g, me, stdB63SwatAwayOracle, tl, "{2}{U}{U}"); got != 2 {
		t.Errorf("a creature attacking you: %d, want 2", got)
	}
}

func TestStdB63SwatAwayPutsACreatureOnItsOwnersChoiceOfLibraryEnd(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Swat Away", "Instant", stdB63SwatAwayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	c := latestPutInLibraryChoice(g)
	if c == nil {
		t.Fatal("no top-or-bottom prompt")
	}
	if c.Chooser != opp.ID {
		t.Errorf("chooser %v, want the owner %v", c.Chooser, opp.ID)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{target}, nil); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if !opp.Library.Contains(target) || rfdOnBattlefield(g, target) {
		t.Error("the creature is not in its owner's library")
	}
}

func TestStdB63SwatAwayPutsASpellIntoTheLibraryWithoutCounteringIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := batch01OpponentCasts(t, g, opp, "Their Spell", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	castCatalogSpell(t, g, "Swat Away", "Instant", stdB63SwatAwayOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	c := latestPutInLibraryChoice(g)
	if c == nil {
		t.Fatal("no top-or-bottom prompt for the spell")
	}
	if c.Chooser != opp.ID {
		t.Errorf("chooser %v, want the spell's owner %v", c.Chooser, opp.ID)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{spell}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Library.Contains(spell) {
		t.Error("the spell is not in its owner's library")
	}
	if opp.Graveyard.Contains(spell) {
		t.Error("the spell went to the graveyard")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell is still on the stack")
	}
}

func latestPutInLibraryChoice(g *game.Game) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePutInLibrary {
			out = c
		}
	}
	return out
}

func TestStdB63TanufelDrawsOnSpellsWithManaValueFourOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tanufel Rimespeaker", TypeLine: "Creature — Elemental Wizard",
		OracleID: stdB63TanufelOracle, Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	hand := me.Hand.Size()
	stdB63CastCreature(t, g, "Three Drop", "Creature — Bear", "{1}{G}{G}", "", nil, 2, 2)
	if me.Hand.Size() != hand {
		t.Fatalf("a mana value 3 spell drew: hand %d, want %d", me.Hand.Size(), hand)
	}
	stdB63CastCreature(t, g, "Four Drop", "Creature — Bear", "{2}{G}{G}", "", nil, 2, 2)
	if me.Hand.Size() != hand+1 {
		t.Errorf("the mana value 4 spell: hand %d, want %d (cast one, drew one)", me.Hand.Size(), hand+1)
	}
}

func TestStdB63ThirstForIdentityAcceptsOneCreatureOrTwoOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	bear := handCardFull(me, "Bear", "Creature — Bear", "", "", nil)
	land := handCardFull(me, "Temple", "Land", "", "", nil)
	castCatalogSpell(t, g, "Thirst for Identity", "Instant", stdB63ThirstOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2+3 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+5)
	}
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land alone was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("one creature card is a complete discard: %v", err)
	}
	if me.Hand.Contains(bear) || !me.Hand.Contains(land) {
		t.Error("only the creature card was discarded")
	}
}
