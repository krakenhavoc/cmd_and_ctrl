package game

import (
	"testing"

	"github.com/google/uuid"
)

// amass_test.go — #1236, the engine half of CR 701.47:
//
//	To amass [subtype] N means "If you don't control an Army creature,
//	create a 0/0 black [subtype] Army creature token. Choose an Army
//	creature you control. Put N +1/+1 counters on that creature. If it
//	isn't a [subtype], it becomes a [subtype] in addition to its other
//	types."
//
// Four parts in one verb, so the tests are about the seams between
// them: find-or-create is a live board query and not a template pick,
// the choice is a prompt when there is one to make, the counters go
// through the CR 614 placement window, the subtype is a layer-4 effect
// pinned to the object, and the whole thing rides the keyword-action
// window so its count is replaceable.
//
// The card half (Orcish Bowmasters, Dreadhorde Invasion, Eternal
// Skylord, Widespread Brutality) is in
// cards/effects/amass_cards_test.go, where a real catalog entry can
// build the Army token.

// amassTestToken is effects.ArmyToken's output, spelled here because
// internal/game owns no token templates — the same 0/0 black Army the
// catalog hands AmassForEffect.
func amassTestToken(subtype string) Card {
	return Card{
		Name:           subtype + " " + ArmySubtype,
		TypeLine:       "Token Creature — " + subtype + " " + ArmySubtype,
		Power:          0,
		Toughness:      0,
		Colors:         []string{"B"},
		PrintedPTKnown: true,
	}
}

// amass takes the keyword action under the write lock and returns the
// Army the verb chose (CR 701.47c).
func amass(t *testing.T, g *Game, actor *Player, subtype string, n int) uuid.UUID {
	t.Helper()
	var army uuid.UUID
	g.WithWriteLock(func() {
		err := g.AmassForEffect(actor.ID, uuid.Nil, amassTestToken(subtype), subtype, n,
			func(_ *Game, chosen uuid.UUID) error {
				army = chosen
				return nil
			})
		if err != nil {
			t.Fatalf("AmassForEffect(%s, %d): %v", subtype, n, err)
		}
	})
	return army
}

// amassView is an Army's post-layer characteristics.
func amassView(t *testing.T, g *Game, army uuid.UUID) *Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findCard(g, army)
	if c == nil {
		t.Fatalf("the Army %s is not on the battlefield", army)
	}
	return c
}

// seedArmy parks an Army creature on the battlefield under `owner`.
// `subtypes` goes into the type line after the dash, so
// seedArmy(g, me, "Zombie") is a Zombie Army and seedArmy(g, me) is a
// plain one.
//
// It carries ONE +1/+1 counter, which is what an amassed Army looks
// like and what keeps it alive: a 0/0 with no counters is killed by
// CR 704.5f at the next state check, and the prompt test below runs
// one when it answers.
func seedArmy(g *Game, owner uuid.UUID, subtypes ...string) uuid.UUID {
	line := "Token Creature — "
	for _, s := range subtypes {
		line += s + " "
	}
	id := pushBattlefieldForTest(g, owner, "Army", line+ArmySubtype, "")
	c := findCard(g, id)
	c.Power, c.Toughness = 0, 0
	c.PrintedPTKnown = true
	c.Counters = map[string]int{CounterPlusOne: 1}
	c.EnteredBattlefieldAt = timeNowUnixNano()
	c.SummonedThisTurn = false
	return id
}

// --- part 1: find or create ------------------------------------------

// TestAmassCreatesAnArmyWhenYouControlNone is CR 701.47a's first
// sentence, and the half #1236 said no primitive could express: the
// token is created only BECAUSE the board has no Army on it.
func TestAmassCreatesAnArmyWhenYouControlNone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	army := amass(t, g, me, "Zombie", 1)

	if army == uuid.Nil {
		t.Fatal("amass chose no Army; it should have created one")
	}
	c := amassView(t, g, army)
	if !c.IsCreature() || !c.HasSubtype(ArmySubtype) || !c.HasSubtype("Zombie") {
		t.Errorf("the token is %q, want a Zombie Army creature", c.TypeLine)
	}
	if !c.IsToken() || c.Controller != me.ID {
		t.Errorf("token=%v controller=%v, want a token under the amassing player", c.IsToken(), c.Controller)
	}
	if got := colorsOfForTest(c); got != "B" {
		t.Errorf("colors = %q, want B — the keyword says a BLACK Army token", got)
	}
	if got := counterOn(g, army, CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 1 || tough != 1 {
		t.Errorf("power/toughness = %d/%d, want 1/1 — a 0/0 body with one counter", p, tough)
	}
}

// TestAmassFindsTheArmyYouAlreadyControl is the other arm of the same
// branch and the reason Dreadhorde Invasion is a two-mana enchantment
// rather than a bad one: the second amass grows the Army it already
// made instead of building a second 0/0.
func TestAmassFindsTheArmyYouAlreadyControl(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	first := amass(t, g, me, "Zombie", 1)
	second := amass(t, g, me, "Zombie", 1)

	if first != second {
		t.Fatalf("the second amass chose %v, want the Army the first one made (%v)", second, first)
	}
	if n := len(armiesOnBattlefieldForTest(g, me.ID)); n != 1 {
		t.Errorf("%d Armies on the battlefield, want 1 — the second amass must not create one", n)
	}
	if got := counterOn(g, first, CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 — one from each amass", got)
	}
}

// TestAmassIgnoresArmiesItDoesNotOwnOrThatAreNotCreatures pins both
// halves of CR 701.47a's "an Army CREATURE YOU CONTROL". An
// opponent's Army is not yours, and an Army that has stopped being a
// creature is not amassable — either one wrongly matched would silence
// the find-or-create branch and hand an opponent your counters.
func TestAmassIgnoresArmiesItDoesNotOwnOrThatAreNotCreatures(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]

	theirs := seedArmy(g, them.ID, "Zombie")
	// Mine, and not a creature: an Army enchanted into a plain
	// artifact. IsArmy is IsCreature AND HasSubtype, so this is
	// invisible to the verb.
	inert := pushBattlefieldForTest(g, me.ID, "Stilled Army", "Artifact — "+ArmySubtype, "")

	army := amass(t, g, me, "Zombie", 1)

	if army == theirs {
		t.Error("amass chose an opponent's Army")
	}
	if army == inert {
		t.Error("amass chose a non-creature Army (CR 701.47a says an Army CREATURE)")
	}
	if got := counterOn(g, theirs, CounterPlusOne); got != 1 {
		t.Errorf("the opponent's Army has %d counters, want the 1 it was seeded with", got)
	}
	if army == uuid.Nil {
		t.Fatal("amass made no Army; neither permanent on the board was a candidate")
	}
}

// TestAmassZeroStillCreatesTheArmy is `actsAtZeroCount` for amass, and
// the War of the Spark release note: "if you're instructed to amass 0,
// you'll create an Army token if you don't control one, but you won't
// put any counters on it". Summons of Saruman off an empty graveyard
// is the printed card.
//
// The 0/0 with no counters then dies to CR 704.5f, which is what
// PrintedPTKnown on the template is for.
func TestAmassZeroStillCreatesTheArmy(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	army := amass(t, g, me, "Orc", 0)

	if army == uuid.Nil {
		t.Fatal("amass 0 created no Army; the count is the only part it skips")
	}
	if got := counterOn(g, army, CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", got)
	}
	c := amassView(t, g, army)
	if c.CurrentToughness() != 0 {
		t.Fatalf("toughness = %d, want 0", c.CurrentToughness())
	}
	runSBAsForTest(g)
	if findCard(g, army) != nil {
		t.Error("the 0/0 Army survived the state-based check (CR 704.5f)")
	}
}

// --- part 2: the choice ----------------------------------------------

// TestAmassWithOneArmyAsksNothing: a forced choice is not a decision,
// and Orcish Bowmasters amasses on every opponent draw. One Army means
// no prompt and the counters land immediately.
func TestAmassWithOneArmyAsksNothing(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	only := seedArmy(g, me.ID, "Zombie")

	army := amass(t, g, me, "Zombie", 2)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts queued for a single-Army amass, want none", len(g.PendingChoices))
	}
	if army != only {
		t.Errorf("chose %v, want the only Army (%v)", army, only)
	}
	if got := counterOn(g, only, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — the seeded one plus the amass's two", got)
	}
}

// TestAmassPromptsWhenYouControlTwoArmies is CR 701.47a's "choose an
// Army creature you control" as a real question. The counters are NOT
// placed when the verb returns — the prompt owns the rest of the
// action — and the continuation is handed the Army the player picked.
func TestAmassPromptsWhenYouControlTwoArmies(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	first := seedArmy(g, me.ID, "Zombie")
	second := seedArmy(g, me.ID, "Zombie")

	var chosen uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		if err := g.AmassForEffect(me.ID, uuid.Nil, amassTestToken("Zombie"), "Zombie", 3,
			func(_ *Game, army uuid.UUID) error {
				ran++
				chosen = army
				return nil
			}); err != nil {
			t.Fatalf("AmassForEffect: %v", err)
		}
	})

	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts queued, want 1", len(g.PendingChoices))
	}
	choice := g.PendingChoices[0]
	if choice.Kind != PendingChoiceChooseCards || choice.Chooser != me.ID {
		t.Fatalf("prompt kind=%v chooser=%v, want a choose-cards prompt for the amassing player",
			choice.Kind, choice.Chooser)
	}
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 {
		t.Errorf("bounds = %d..%d, want exactly one Army", choice.ChooseMin, choice.ChooseMax)
	}
	if len(choice.ChooseCards) != 2 {
		t.Errorf("%d candidates, want both Armies", len(choice.ChooseCards))
	}
	if ran != 0 || counterOn(g, first, CounterPlusOne) != 1 || counterOn(g, second, CounterPlusOne) != 1 {
		t.Fatal("the amass finished before the prompt was answered")
	}

	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if ran != 1 || chosen != second {
		t.Errorf("continuation ran %d time(s) with %v, want once with the picked Army (%v)", ran, chosen, second)
	}
	if got := counterOn(g, second, CounterPlusOne); got != 4 {
		t.Errorf("chosen Army has %d counters, want 4 — the seeded one plus the amass's three", got)
	}
	if got := counterOn(g, first, CounterPlusOne); got != 1 {
		t.Errorf("the Army that was NOT chosen has %d counters, want the 1 it was seeded with", got)
	}
}

// --- part 3: the counters --------------------------------------------

// TestAmassCountersGoThroughThePlacementWindow: the counters are
// PLACED on a permanent already on the battlefield, so the CR 614
// counter window opens and Hardened Scales sees them — including on
// the token the amass just made, which enters 0/0 and is counted up
// afterwards exactly as printed.
func TestAmassCountersGoThroughThePlacementWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { g.RegisterReplacementForTest(hardenedScalesForTest()) })

	army := amass(t, g, me, "Zombie", 2)

	if got := counterOn(g, army, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — Hardened Scales adds one to the amass's two", got)
	}
}

// --- part 4: the subtype ---------------------------------------------

// TestAmassGivesTheChosenArmyTheKeywordsSubtype is CR 701.47a's last
// sentence — "if it isn't a [subtype], it becomes a [subtype] in
// addition to its other types" — and the `in addition` is the half
// worth pinning: a Zombie Army amassed as Orcs is an Orc Zombie Army,
// not an Orc Army.
func TestAmassGivesTheChosenArmyTheKeywordsSubtype(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	existing := seedArmy(g, me.ID, "Zombie")

	army := amass(t, g, me, "Orc", 1)
	if army != existing {
		t.Fatalf("amass chose %v, want the Army already out (%v)", army, existing)
	}

	c := amassView(t, g, army)
	if !c.HasSubtype("Orc") {
		t.Error(`the Army did not become an Orc ("it's also an Orc")`)
	}
	if !c.HasSubtype("Zombie") {
		t.Error("the Army stopped being a Zombie — the subtype is ADDED, not set")
	}
	if !c.HasSubtype(ArmySubtype) || !c.IsCreature() {
		t.Error("the Army stopped being an Army creature")
	}
}

// TestAmassRegistersNoSubtypeEffectWhenItAlreadyHasOne is the "if it
// isn't" guard, and it is cost control as much as rules: Orcish
// Bowmasters amasses on EVERY opponent draw, and a layer-4 effect per
// amass would leave a hundred identical entries in the snapshot census.
func TestAmassRegistersNoSubtypeEffectWhenItAlreadyHasOne(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	amass(t, g, me, "Zombie", 1)
	if n := len(g.ScopedEffects); n != 0 {
		t.Fatalf("%d layer effects after the first amass, want 0 — the token is created AS a Zombie", n)
	}
	for i := 0; i < 5; i++ {
		amass(t, g, me, "Zombie", 1)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d layer effects after six amasses, want 0", n)
	}

	// One Orc amass on the same Zombie Army registers exactly one, and
	// the next five register none.
	amass(t, g, me, "Orc", 1)
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("%d layer effects after the first Orc amass, want 1", n)
	}
	for i := 0; i < 5; i++ {
		amass(t, g, me, "Orc", 1)
	}
	if n := len(g.ScopedEffects); n != 1 {
		t.Errorf("%d layer effects after six Orc amasses, want 1 — the guard reads the EFFECTIVE subtype", n)
	}
}

// TestTheAmassSubtypeEndsWhenTheArmyLeaves: the grant states no
// duration (CR 611.2a) but the object it names is gone the moment it
// leaves (CR 400.7), so the pin drops the registry entry rather than
// leaving a husk per amass for the rest of the game.
func TestTheAmassSubtypeEndsWhenTheArmyLeaves(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	existing := seedArmy(g, me.ID, "Zombie")

	amass(t, g, me, "Orc", 1)
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("%d layer effects registered, want 1", n)
	}

	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield},
		ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, existing); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d layer effects survive the Army leaving, want 0 (PinnedTo + CR 400.7)", n)
	}
}

// --- the CR 614 window on the ACTION ---------------------------------

// TestAmassRidesTheKeywordActionWindow: the COUNT is on the event, so
// "if you would amass, amass that much plus one instead" is a
// card-side replacement and no engine change. #1236's third bullet.
func TestAmassRidesTheKeywordActionWindow(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCountReplacement(KeywordActionAmass,
			func(n int) int { return n + 1 }, "amass one more (probe)"))
	})

	army := amass(t, g, me, "Zombie", 2)

	if got := counterOn(g, army, CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — the window rewrote the amass's count", got)
	}
}

// TestACancelledAmassStillRunsTheRestOfTheSentence is CR 614.10 with a
// null replacement beside CR 701.47b's "a player amassed after the
// process is complete, even if some or all of those actions were
// impossible": no Army is made, no counters are placed, and Widespread
// Brutality's second half still runs — with nothing to point at.
func TestACancelledAmassStillRunsTheRestOfTheSentence(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(keywordCancelReplacement(KeywordActionAmass))
	})

	ran := 0
	var chosen uuid.UUID = uuid.New()
	g.WithWriteLock(func() {
		if err := g.AmassForEffect(me.ID, uuid.Nil, amassTestToken("Zombie"), "Zombie", 2,
			func(_ *Game, army uuid.UUID) error {
				ran++
				chosen = army
				return nil
			}); err != nil {
			t.Fatalf("AmassForEffect: %v", err)
		}
	})

	if ran != 1 {
		t.Fatalf("the continuation ran %d time(s), want exactly once", ran)
	}
	if chosen != uuid.Nil {
		t.Errorf("the continuation was handed %v, want uuid.Nil — nothing was amassed", chosen)
	}
	if n := len(armiesOnBattlefieldForTest(g, me.ID)); n != 0 {
		t.Errorf("%d Armies on the battlefield after a cancelled amass, want 0", n)
	}
}

// TestTheAmassContinuationRunsExactlyOnce: it is cleared THROUGH the
// tail pointer, so the settled path and the abandoned path cannot both
// run it. Without the clear, a clause that re-enters the pipeline on
// the same tail fires twice.
func TestTheAmassContinuationRunsExactlyOnce(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]

	ran := 0
	g.WithWriteLock(func() {
		if err := g.AmassForEffect(me.ID, uuid.Nil, amassTestToken("Zombie"), "Zombie", 1,
			func(_ *Game, _ uuid.UUID) error {
				ran++
				return nil
			}); err != nil {
			t.Fatalf("AmassForEffect: %v", err)
		}
	})
	if ran != 1 {
		t.Errorf("the continuation ran %d time(s), want exactly once", ran)
	}
}

// --- the read side ---------------------------------------------------

// TestArmiesControlledForEffectIsTheReadSide: "an Army you control" is
// a board query eight printed cards make without amassing at all
// (Sauron, the Dark Lord; March from the Black Gate), and it is the
// same walk the verb uses.
func TestArmiesControlledForEffectIsTheReadSide(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := seedArmy(g, me.ID, "Orc")
	seedArmy(g, them.ID, "Orc")
	pushBattlefieldForTest(g, me.ID, "Grizzly Bears", "Creature — Bear", "")

	var got []uuid.UUID
	g.WithWriteLock(func() { got = g.ArmiesControlledForEffect(me.ID) })

	if len(got) != 1 || got[0] != mine {
		t.Errorf("ArmiesControlledForEffect = %v, want just my own Army (%v)", got, mine)
	}
}

// armiesOnBattlefieldForTest counts a player's Armies without the
// write lock the effect-time read wants.
func armiesOnBattlefieldForTest(g *Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && IsArmy(*c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// colorsOfForTest renders a card's colours as a stable string.
func colorsOfForTest(c *Card) string {
	out := ""
	for _, col := range c.EffectiveColors() {
		out += col
	}
	return out
}
