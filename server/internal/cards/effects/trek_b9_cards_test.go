package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b9_cards_test.go — Star Trek, batch 9 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB9JockeyOracle     = "d41ff489-7d74-4b06-b3a1-638f3e3303d1"
	trekB9TPolOracle       = "e6973668-4ea7-4920-82a5-73abfc1c9fd5"
	trekB9SuurokOracle     = "5d2016ed-3ea0-4938-b95e-ee582307e475"
	trekB9SuusMahnaOracle  = "e7d22ef6-a926-4964-bfd6-f7345b20e74e"
	trekB9PrioritiesOracle = "cd24de4f-ce88-4fbb-9e8d-93ea7c260b4c"
	trekB9ScientistsOracle = "5ca4e8c4-8358-49be-bafb-85849bf9bd01"
	trekB9VGerOracle       = "f50ec16b-13f5-4e20-ae09-9c7dee9d223b"
	trekB9MindMeldOracle   = "a6aa204a-0be9-4a38-b54c-82538adfe7f1"
	trekB9WesleyOracle     = "5c3cbb08-cbe1-42f5-9ac7-e558b85648bb"
	trekB9ZeframOracle     = "fea67446-f413-4a8c-b79a-49172e798fff"
)

func TestTrekB9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB9JockeyOracle:     "Starfleet Spanner Jockey",
		trekB9TPolOracle:       "Subcommander T'Pol",
		trekB9SuurokOracle:     "Suurok-Class Science Vessel",
		trekB9SuusMahnaOracle:  "Suus Mahna Practitioner",
		trekB9PrioritiesOracle: "Test of Priorities",
		trekB9ScientistsOracle: "Those Old Scientists",
		trekB9VGerOracle:       "V'Ger, the Intruder",
		trekB9MindMeldOracle:   "Vulcan Mind-Meld",
		trekB9WesleyOracle:     "Wesley Crusher, Starfleet Cadet",
		trekB9ZeframOracle:     "Zefram Cochrane, Warp Pioneer",
	} {
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

// --- shared test helpers ------------------------------------------------------

// trekB9Hand puts a card in p's hand and returns its id.
func trekB9Hand(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	return c.InstanceID
}

// trekB9ModePick waits for the mode prompt addressed to `chooser` and
// answers it with `modes`.
func trekB9ModePick(t *testing.T, g *game.Game, chooser uuid.UUID, modes ...int) {
	t.Helper()
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, chooser)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, chooser, modes); err != nil {
		t.Fatalf("ResolveModePick %v: %v", modes, err)
	}
}

// trekB9Settle resolves the stack, answering every discard prompt for
// `p` off the top of the hand, and returns how many it answered.
func trekB9Settle(t *testing.T, g *game.Game, p uuid.UUID) int {
	t.Helper()
	discards := 0
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if discardChoiceFor(g, p) == nil {
			return discards
		}
		discardFromHand(t, g, p)
		discards++
	}
	t.Fatal("the stack never settled")
	return discards
}

// --- Starfleet Spanner Jockey -------------------------------------------------

func TestTrekB9JockeyAttacksOnlyAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jockey := b27Push(g, me.ID, "Starfleet Spanner Jockey", "Creature — Human Artificer", trekB9JockeyOracle, "{1}{U}", 3, 3, "U")
	advanceToMain(t, g)

	// A creature spell does not count.
	castFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, jockey); len(c.Effective().DefenderWaivers) != 0 {
		t.Fatal("a creature spell lifted defender")
	}

	castFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, jockey)
	if !game.HasKeyword(&c, "defender") {
		t.Error("the Jockey lost defender; it should keep it and attack anyway")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jockey, opp.ID); err != nil {
		t.Fatalf("after a noncreature spell the Jockey could not attack: %v", err)
	}
}

func TestTrekB9JockeyCantAttackWithoutANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jockey := b27Push(g, me.ID, "Starfleet Spanner Jockey", "Creature — Human Artificer", trekB9JockeyOracle, "{1}{U}", 3, 3, "U")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jockey, opp.ID); !errors.Is(err, game.ErrDefender) {
		t.Fatalf("DeclareAttacker = %v, want ErrDefender", err)
	}
}

// --- Subcommander T'Pol -------------------------------------------------------

func TestTrekB9TPolHasFlashAndTapsAndStunsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The draw step: only flash lets a creature be cast here.
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("at %s, want the draw step", g.Turn.Step)
	}
	floatMana(t, g, me, "{C}{C}{U}")
	castFromHandForTest(t, g, me, "Subcommander T'Pol", "Legendary Creature — Vulcan Scientist", "{2}{U}", trekB9TPolOracle, game.CastSpellParams{Strict: true})
	trekB9ModePick(t, g, me.ID, 1)
	if err := sgAnswerTargets(t, g, theirs); err != nil {
		t.Fatalf("targeting the bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, theirs)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("the bear tapped=%v stun=%d, want tapped with one stun counter", c.Tapped, c.Counters[game.CounterStun])
	}
}

func TestTrekB9TPolCountersATriggeredAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vessel := castCatalogSpell(t, g, "Suurok-Class Science Vessel", "Artifact — Spacecraft", trekB9SuurokOracle, nil)
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	draw := triggerOnStack(g, vessel)
	if draw == nil {
		t.Fatal("the Vessel's draw trigger is not on the stack")
	}
	before := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}{U}")
	castFromHandForTest(t, g, me, "Subcommander T'Pol", "Legendary Creature — Vulcan Scientist", "{2}{U}", trekB9TPolOracle, game.CastSpellParams{Strict: true})
	trekB9ModePick(t, g, me.ID, 0)
	if err := sgAnswerTargets(t, g, draw.ID); err != nil {
		t.Fatalf("targeting the draw trigger: %v", err)
	}
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, vessel) != nil {
		t.Error("the countered trigger is still on the stack")
	}
	// T'Pol was put in hand and cast; nothing was drawn.
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (the draw was countered)", got, before)
	}
}

func TestTrekB9TPolWithoutFlashWouldBeRefusedAtThisTiming(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	floatMana(t, g, me, "{C}{C}{U}")
	id := trekB9Hand(g, me, game.Card{Name: "Plain Vulcan", TypeLine: "Creature — Vulcan", ManaCost: "{2}{U}"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature without flash was cast in the draw step; the flash test proves nothing")
	}
}

// --- Suurok-Class Science Vessel ----------------------------------------------

func TestTrekB9SuurokDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{U}")
	before := me.Hand.Size()
	castFromHandForTest(t, g, me, "Suurok-Class Science Vessel", "Artifact — Spacecraft", "{2}{U}", trekB9SuurokOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	// The helper puts the spell in hand and casts it; the draw is the +1.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (drew one)", got, before+1)
	}
}

func TestTrekB9SuurokIsAFlierAtFive(t *testing.T) {
	g, me, ship, crew := stationBoard(t, "Suurok-Class Science Vessel", "Artifact — Spacecraft", trekB9SuurokOracle, 4)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, ship); c.IsCreature() || game.HasKeyword(&c, "flying") {
		t.Fatal("four charge counters is below 5+")
	}
	second := b12Creature(g, me.ID, "Ensign", "Creature — Human", 1, 1)
	station(t, g, me, ship, second)
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || c.CurrentPower() != 3 || c.CurrentToughness() != 2 || !game.HasKeyword(&c, "flying") {
		t.Errorf("at 5+ want a 3/2 flying artifact creature; creature=%v %d/%d flying=%v",
			c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
}

// --- Suus Mahna Practitioner --------------------------------------------------

func TestTrekB9SuusMahnaCostsLessPerDifferentlyNamedLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := b27Push(g, me.ID, "Suus Mahna Practitioner", "Creature — Vulcan Officer", trekB9SuusMahnaOracle, "{2}{U}", 3, 2, "U")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// The lands are tapped so they count but pay nothing; only the floated
	// mana pays.
	for _, n := range []string{"Island", "Island", "Plains"} {
		tmtB4Tap(g, b12Permanent(g, me.ID, n, "Basic Land — "+n))
	}
	// An opponent's land is not yours.
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}

	// Two names: {4}{U} costs {2}{U}.
	floatMana(t, g, me, "{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{Targets: targets, Strict: true}); err == nil {
		t.Fatal("activated for {1}{U} with two differently named lands")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{Targets: targets, Strict: true}); err != nil {
		t.Fatalf("activate for {2}{U} with two differently named lands: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d after paying {2}{U}, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, theirs).Tapped {
		t.Error("the target did not tap")
	}
	if !layeredCard(t, g, monk).Tapped {
		t.Error("the {T} cost did not tap the Practitioner")
	}
}

func TestTrekB9SuusMahnaNeedsItsBlueEvenWithManyLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monk := b27Push(g, me.ID, "Suus Mahna Practitioner", "Creature — Vulcan Officer", trekB9SuusMahnaOracle, "{2}{U}", 3, 2, "U")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, n := range []string{"Island", "Plains", "Swamp", "Mountain", "Forest"} {
		tmtB4Tap(g, b12Permanent(g, me.ID, n, "Basic Land — "+n))
	}
	advanceToMain(t, g)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{Targets: targets, Strict: true}); err == nil {
		t.Fatal("five names took the {U} off the cost")
	}
	floatMana(t, g, me, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, monk, 0, game.ActivateAbilityParams{Targets: targets, Strict: true}); err != nil {
		t.Fatalf("activate for {U} with five differently named lands: %v", err)
	}
}

// --- Test of Priorities -------------------------------------------------------

func TestTrekB9PrioritiesPowerToEnginesSurveilsThenDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	floatMana(t, g, me, "{C}{U}")
	before := me.Hand.Size()
	libBefore := me.Library.Size()
	castFromHandForTest(t, g, me, "Test of Priorities", "Instant", "{1}{U}", trekB9PrioritiesOracle,
		game.CastSpellParams{Strict: true, Modes: []int{0}})
	passPriorityAroundTable(t, g)
	if ch := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); ch != nil {
		if err := g.ResolveSurveil(ch.ID, me.ID, ch.ScryCards, nil); err != nil {
			t.Fatalf("ResolveSurveil: %v", err)
		}
	} else {
		t.Fatal("no surveil prompt")
	}
	passPriorityAroundTable(t, g)
	// The helper puts the spell in hand and casts it; the draw is the +1.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (drew one)", got, before+1)
	}
	if got := me.Library.Size(); got != libBefore-3 {
		t.Errorf("library %d, want %d (two surveilled away, one drawn)", got, libBefore-3)
	}
}

func TestTrekB9PrioritiesPowerToLifeSupportReturnsAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{C}{U}")
	id := trekB9Hand(g, me, game.Card{Name: "Test of Priorities", TypeLine: "Instant", ManaCost: "{1}{U}", OracleID: trekB9PrioritiesOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("an opponent's creature was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err != nil {
		t.Fatalf("cast at my creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Fatal("the creature was not exiled")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	back := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "My Bear" && c.Controller == me.ID {
			back++
		}
	}
	if back != 1 {
		t.Errorf("%d My Bears back at the end step, want 1", back)
	}
}

// --- Those Old Scientists -----------------------------------------------------

// trekB9Scientists puts the Saga on seat 0's side with `lore` lore
// counters already on it; the precombat-main turn-based action (CR 714.3b)
// then adds the next one.
func trekB9Scientists(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Those Old Scientists", TypeLine: "Enchantment — Saga",
		OracleID: trekB9ScientistsOracle, ManaCost: "{2}{U}", Colors: []string{"U"},
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLore: lore},
	})
}

func TestTrekB9ScientistsChapterOneDrawsTwoThenDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Those Old Scientists", "Enchantment — Saga", trekB9ScientistsOracle, nil)
	if n := trekB9Settle(t, g, me.ID); n != 1 {
		t.Errorf("%d discard prompts, want one prompt for two cards", n)
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (cast it, draw two, discard two)", got, before)
	}
}

func TestTrekB9ScientistsChapterTwoReturnsAnArtifactInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB9Scientists(g, me.ID, 1)
	creature := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	theirs := batch01GraveyardCard(opp, "Their Spell", "Sorcery")
	sorcery := batch01GraveyardCard(me, "My Spell", "Sorcery")
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	for _, bad := range []uuid.UUID{creature, theirs} {
		if err := sgAnswerTargets(t, g, bad); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("%s: err = %v, want ErrIllegalTarget", cardNameFor(g, bad), err)
		}
	}
	if err := sgAnswerTargets(t, g, sorcery); err != nil {
		t.Fatalf("my sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(sorcery) {
		t.Error("the sorcery did not return to hand")
	}
}

func TestTrekB9ScientistsChapterThreeShufflesCardsIntoTheirOwnersLibraries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saga := trekB9Scientists(g, me.ID, 2)
	mine := batch01GraveyardCard(me, "Mine", "Sorcery")
	theirs := batch01GraveyardCard(opp, "Theirs", "Creature — Bear")
	myLib, theirLib := me.Library.Size(), opp.Library.Size()
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if err := sgAnswerTargets(t, g, mine, theirs); err != nil {
		t.Fatalf("two cards in two graveyards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Library.Contains(mine) || !opp.Library.Contains(theirs) {
		t.Error("each card should be in its owner's library")
	}
	if me.Library.Size() != myLib+1 || opp.Library.Size() != theirLib+1 {
		t.Errorf("libraries %d/%d, want %d/%d", me.Library.Size(), opp.Library.Size(), myLib+1, theirLib+1)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

// --- V'Ger, the Intruder ------------------------------------------------------

func TestTrekB9VGerLooksAtAHandAndDrawsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "V'Ger, the Intruder", "Legendary Artifact Creature — Construct", trekB9VGerOracle, nil)
	trekB9ModePick(t, g, me.ID, 0)
	if err := g.ResolvePickTargets(pickTargetPrompt(t, g).ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("V'Ger's controller was a legal target opponent")
	}
	if err := g.ResolvePickTargets(pickTargetPrompt(t, g).ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("targeting an opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The helper puts the spell in hand and casts it; the draw is the +1.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
	for _, c := range opp.Hand.Cards {
		if !c.IsKnownTo(me.ID) {
			t.Fatalf("%s in the opponent's hand is not known to V'Ger's controller", c.Name)
		}
		if c.IsKnownTo(g.Seats[2].ID) {
			t.Fatalf("%s was shown to another player", c.Name)
		}
	}
}

func TestTrekB9VGerShrinksOpponentsCreaturesUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vger := b27Push(g, me.ID, "V'Ger, the Intruder", "Legendary Artifact Creature — Construct", trekB9VGerOracle, "{6}{U}", 5, 7, "U")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(vger, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("V'Ger's attack trigger asked no mode: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 1 || c.CurrentToughness() != 2 {
		t.Errorf("their bear %d/%d, want 1/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, mine); c.CurrentPower() != 2 {
		t.Errorf("my bear %d power, want 2", c.CurrentPower())
	}
	// Still shrunk on the next player's turn; gone on mine.
	advanceToUpkeepOf(t, g, 1)
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 1 {
		t.Errorf("on the opponent's turn their bear has %d power, want 1", c.CurrentPower())
	}
	advanceToUpkeepOf(t, g, 0)
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("on my next turn their bear has %d power, want 2", c.CurrentPower())
	}
}

// --- Vulcan Mind-Meld ---------------------------------------------------------

func TestTrekB9MindMeldSwapsTopCardsFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mineTop := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	n := len(opp.Library.Cards)
	theirTop := []uuid.UUID{opp.Library.Cards[n-1].InstanceID, opp.Library.Cards[n-2].InstanceID}
	floatMana(t, g, me, "{U}")
	id := trekB9Hand(g, me, game.Card{Name: "Vulcan Mind-Meld", TypeLine: "Sorcery", ManaCost: "{U}", OracleID: trekB9MindMeldOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("the caster was a legal target opponent")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast at an opponent: %v", err)
	}
	passPriorityAroundTable(t, g)

	if perm := e2Permission(g, mineTop); perm == nil || perm.Player != opp.ID || perm.CastOnly || !perm.AnyType {
		t.Errorf("my top card's permission %+v, want an any-type play grant to the opponent", perm)
	}
	e2AssertFaceDownFor(t, g, mineTop, opp.ID)
	for _, c := range theirTop {
		if perm := e2Permission(g, c); perm == nil || perm.Player != me.ID || perm.CastOnly || !perm.AnyType {
			t.Errorf("their top card's permission %+v, want an any-type play grant to the caster", perm)
		}
		e2AssertFaceDownFor(t, g, c, me.ID)
	}
	if got := n - len(opp.Library.Cards); got != 2 {
		t.Errorf("exiled %d of the opponent's cards, want 2", got)
	}
}

func TestTrekB9MindMeldCardIsCastableWithAnyType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seer := game.NewCard("Their Seer", opp.ID)
	seer.TypeLine = "Creature — Eldrazi"
	seer.ManaCost = "{2}{C}"
	opp.Library.PushTop(seer)
	floatMana(t, g, me, "{U}")
	id := trekB9Hand(g, me, game.Card{Name: "Vulcan Mind-Meld", TypeLine: "Sorcery", ManaCost: "{U}", OracleID: trekB9MindMeldOracle})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	e2Swamps(g, me.ID, 3)
	if err := g.CastSpell(me.ID, seer.InstanceID, game.CastSpellParams{FromZone: "exile", AutoTap: true}); err != nil {
		t.Fatalf("cast the exiled {2}{C} card off three Swamps: %v", err)
	}
}

// --- Wesley Crusher, Starfleet Cadet ------------------------------------------

func TestTrekB9WesleyBouncesASmallCreatureAndMakesAnotherUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b27Push(g, opp.ID, "Small Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	big := b27Push(g, opp.ID, "Big Bear", "Creature — Bear", "", "{2}{G}", 3, 3, "G")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{U}{U}")
	castFromHandForTest(t, g, me, "Wesley Crusher, Starfleet Cadet", "Legendary Creature — Human Officer", "{2}{U}{U}", trekB9WesleyOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err == nil {
		t.Fatal("one mode was accepted for \"choose two\"")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := sgAnswerTargets(t, g, big); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a mana value 3 creature: err = %v, want ErrIllegalTarget", err)
	}
	if err := sgAnswerTargets(t, g, small); err != nil {
		t.Fatalf("bounce target: %v", err)
	}
	if err := sgAnswerTargets(t, g, mine); err != nil {
		t.Fatalf("unblockable target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(small) {
		t.Error("the small bear did not return to its owner's hand")
	}
	if auraRestrictions(t, g, mine)&game.CantBeBlocked == 0 {
		t.Error("my bear can be blocked")
	}
	if auraRestrictions(t, g, big)&game.CantBeBlocked != 0 {
		t.Error("an untargeted creature became unblockable")
	}
}

func TestTrekB9WesleyCountersANoncreatureSpellUnlessPaidAndLootsAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The opponent casts a sorcery-speed-free instant; Wesley flashes in
	// over it.
	spell := trekB9Hand(g, opp, game.Card{Name: "Their Instant", TypeLine: "Instant"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent casts: %v", err)
	}
	for i := 0; i < 4 && g.Turn.PriorityHolder != 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	floatMana(t, g, me, "{C}{C}{U}{U}")
	castFromHandForTest(t, g, me, "Wesley Crusher, Starfleet Cadet", "Legendary Creature — Human Officer", "{2}{U}{U}", trekB9WesleyOracle, game.CastSpellParams{Strict: true})
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0, 3}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := sgAnswerTargets(t, g, spell); err != nil {
		t.Fatalf("counter target: %v", err)
	}
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("loot target: %v", err)
	}
	oppHand := opp.Hand.Size()
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerPayUnless(t, g, opp.ID, false)
	trekB9Settle(t, g, opp.ID)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the unpaid spell was not countered")
	}
	if got := opp.Hand.Size(); got != oppHand {
		t.Errorf("opponent hand %d, want %d (drew two, discarded two)", got, oppHand)
	}
}

// --- Zefram Cochrane, Warp Pioneer --------------------------------------------

func TestTrekB9ZeframCastsOneArtifactFromTheTopEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Zefram Cochrane, Warp Pioneer", "Legendary Creature — Human Pilot", trekB9ZeframOracle, "{2}{U}", 2, 3, "U")
	advanceToMain(t, g)
	second := pushLibraryCard(me, game.Card{Name: "Second Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	first := pushLibraryCard(me, game.Card{Name: "First Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	floatMana(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a card under the top was cast")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the top artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a second artifact was cast from the top in one turn")
	}
}

func TestTrekB9ZeframDoesNotOpenANonartifactOrAnotherTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Zefram Cochrane, Warp Pioneer", "Legendary Creature — Human Pilot", trekB9ZeframOracle, "{2}{U}", 2, 3, "U")
	advanceToMain(t, g)
	sorcery := pushLibraryCard(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{1}"})
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a sorcery was cast from the top")
	}
	var mine, theirs bool
	g.ReadSnapshot(func() {
		mine = g.LibraryTopVisibleToLocked(me.ID, me.ID)
		theirs = g.LibraryTopVisibleToLocked(me.ID, opp.ID)
	})
	if !mine || theirs {
		t.Errorf("top visible to Zefram's controller %v, to an opponent %v; want true, false", mine, theirs)
	}
}

func TestTrekB9ZeframSpacecraftEnterWithTwoMoreChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Zefram Cochrane, Warp Pioneer", "Legendary Creature — Human Pilot", trekB9ZeframOracle, "{2}{U}", 2, 3, "U")
	advanceToMain(t, g)
	ship := castAndResolveCreature(t, g, "Suurok-Class Science Vessel", "Artifact — Spacecraft", trekB9SuurokOracle)
	passPriorityAroundTable(t, g)
	if got := layeredCard(t, g, ship).Counters[game.CounterCharge]; got != 2 {
		t.Errorf("the Spacecraft has %d charge counters, want 2", got)
	}
	rock := castAndResolveCreature(t, g, "Plain Rock", "Artifact", "")
	if got := layeredCard(t, g, rock).Counters[game.CounterCharge]; got != 0 {
		t.Errorf("a non-Spacecraft artifact has %d charge counters, want 0", got)
	}
}
