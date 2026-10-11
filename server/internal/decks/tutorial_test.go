package decks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	oraclefix "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/coverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// tutorial_test.go pins what ADR 0076 §2.2 asks of the practice table's
// two decks beyond what every deck owes (decks_test.go walks both for
// that). The dump-gated half — printed power, oracle text — is in
// realdump_manual_test.go.

// TestTutorialDecksAreNotOffered: the pickers list `all`, and the
// tutorial pair is not a choice. A player browsing GET /decks for a
// real deck must not be offered seventy-two Forests.
func TestTutorialDecksAreNotOffered(t *testing.T) {
	for _, d := range tutorialDecks() {
		if _, ok := Lookup(d.ID); ok {
			t.Errorf("Lookup(%q) found a tutorial deck; the pickers would offer it", d.ID)
		}
		for _, id := range IDs() {
			if id == d.ID {
				t.Errorf("IDs() lists the tutorial deck %q", d.ID)
			}
		}
		for _, other := range All() {
			if other.ID == d.ID {
				t.Errorf("All() includes the tutorial deck %q", d.ID)
			}
		}
	}
	if TutorialPlayer().ID == TutorialBot().ID {
		t.Fatal("the tutorial's two decks share an ID")
	}
}

// TestTutorialDecksPlayExactlyAsPrinted holds the pair to more than
// registration: a new player's first game is the worst place for a
// card that does something other than what it says, so every card is
// graded CompletenessFull, not merely present.
func TestTutorialDecksPlayExactlyAsPrinted(t *testing.T) {
	for _, d := range tutorialDecks() {
		t.Run(d.ID, func(t *testing.T) {
			if d.Name == "" || d.Summary == "" || d.Archetype == "" {
				t.Errorf("Name, Summary and Archetype must be set")
			}
			for _, c := range d.Cards() {
				if c.Basic {
					continue
				}
				spec, ok := effects.Lookup(c.OracleID)
				if !ok {
					continue // reported by TestEveryCardResolvesToARegisteredSpec
				}
				if spec.Completeness != effects.CompletenessFull {
					t.Errorf("%s is graded %s; the tutorial decks take only cards that play as printed",
						c.Name, spec.Completeness)
				}
			}
		})
	}
}

// botEvasion is the keyword list the practice bot's deck must not
// print. "No evasion" (ADR 0076 §2.2) is read broadly: anything that
// makes a small attacker hard to block, or a small blocker deadly, is
// out, because the point is that the bot cannot hurt a player who is
// still reading the coach card.
var botEvasion = []string{
	"flying", "trample", "menace", "fear", "intimidate", "shadow", "horsemanship", "skulk",
	"deathtouch", "first strike", "double strike", "haste", "infect", "wither", "toxic",
	"lifelink", "flash",
}

// TestTutorialBotDeckIsSlow is §2.2's "lands and small bodies, no
// removal, no evasion", as far as the effect catalog can see it without
// a Scryfall dump: no printed evasion keyword, and no ability on any
// card that picks a target — creature-borne removal is always a
// targeted trigger or activation. Printed power is the realdump test's.
func TestTutorialBotDeckIsSlow(t *testing.T) {
	d := TutorialBot()
	lands, spells := 0, 0
	for _, c := range d.Cards() {
		if c.Basic {
			lands += c.Copies()
			continue
		}
		spells++
		spec, ok := effects.Lookup(c.OracleID)
		if !ok {
			continue
		}
		for _, kw := range spec.PrintedKeywords {
			for _, bad := range botEvasion {
				if strings.EqualFold(kw, bad) {
					t.Errorf("%s prints %s; the practice bot's deck has no evasion", c.Name, kw)
				}
			}
		}
		if spec.Targets != nil {
			t.Errorf("%s targets; the practice bot's deck has no removal", c.Name)
		}
		for _, a := range spec.Activated {
			if a.Targets != nil {
				t.Errorf("%s has a targeted activated ability (%q)", c.Name, a.Label)
			}
		}
		for _, tr := range spec.Triggered {
			if tr.Targets != nil || tr.TargetsFrom != nil {
				t.Errorf("%s has a targeted triggered ability (%q)", c.Name, tr.Key)
			}
		}
	}
	// Lands, mostly. Two thirds of the deck is the floor.
	if lands < 66 {
		t.Errorf("the practice bot's deck has %d lands; it is meant to be mostly land", lands)
	}
	if spells == 0 {
		t.Error("the practice bot's deck has no spells at all; the bot has nothing to show the player")
	}
}

// TestTutorialPlayerDeckTeachesTheLessons checks the player's deck can
// actually stage the tutorial (ADR 0076 §2.1): enough Forests that the
// opening hand has a land to play, and enough permanents with an
// ability behind right-click that step 7 has something to open. A mana
// creature counts — its mana ability is what the right-click menu
// lists first.
//
// It also asks nothing at resolution: no card picks a target. The
// tutorial teaches the client, and a prompt it has not explained is a
// lesson it did not plan.
func TestTutorialPlayerDeckTeachesTheLessons(t *testing.T) {
	d := TutorialPlayer()
	lands, rightClick := 0, 0
	for _, c := range d.Cards() {
		if c.Basic {
			lands += c.Copies()
			continue
		}
		spec, ok := effects.Lookup(c.OracleID)
		if !ok {
			continue
		}
		if len(spec.ManaAbilities) > 0 || len(spec.Activated) > 0 {
			rightClick++
		}
		if spec.Targets != nil {
			t.Errorf("%s targets; the player's tutorial deck asks no questions", c.Name)
		}
		for _, a := range spec.Activated {
			if a.Targets != nil {
				t.Errorf("%s has a targeted activated ability (%q)", c.Name, a.Label)
			}
		}
		for _, tr := range spec.Triggered {
			if tr.Targets != nil || tr.TargetsFrom != nil {
				t.Errorf("%s has a targeted triggered ability (%q)", c.Name, tr.Key)
			}
		}
	}
	// Half the deck: under one opening seven in a hundred has no land.
	if lands < 50 {
		t.Errorf("the player's tutorial deck has %d lands; step 3 needs one in the opening hand", lands)
	}
	if rightClick < 12 {
		t.Errorf("only %d cards in the player's tutorial deck have a right-click ability; step 9 needs one in play", rightClick)
	}
}

// TestTutorialOpeningHandIsInTheDeck pins the practice table's scripted
// opening seven (ADR 0076 §2.2, amended 2026-10-11, #2992, #3012) to the
// deck it is dealt from: seven names, each one the deck holds enough
// copies of, three Forests for the land lessons, Llanowar Elves, a
// one-drop whose mana ability the right-click step uses on turn two, and
// Memnite, the free creature cast beside it on turn one that attacks on
// turn two.
func TestTutorialOpeningHandIsInTheDeck(t *testing.T) {
	hand := TutorialOpeningHand()
	if len(hand) != 7 {
		t.Fatalf("the scripted opening hand has %d cards, want 7", len(hand))
	}
	want := map[string]int{}
	for _, n := range hand {
		want[n]++
	}
	have := map[string]int{}
	oracle := map[string]string{}
	for _, c := range TutorialPlayer().Cards() {
		have[c.Name] += c.Copies()
		oracle[c.Name] = c.OracleID
	}
	for n, k := range want {
		if have[n] < k {
			t.Errorf("the scripted hand deals %d %s; the deck holds %d", k, n, have[n])
		}
	}
	if want["Forest"] != 3 {
		t.Errorf("the scripted hand has %d Forests, want 3", want["Forest"])
	}
	spec, ok := effects.Lookup(oracle["Llanowar Elves"])
	if want["Llanowar Elves"] != 1 || !ok || len(spec.ManaAbilities) == 0 {
		t.Error("the scripted hand needs Llanowar Elves, a one-drop with a mana ability for the right-click step")
	}
	if want["Memnite"] != 1 {
		t.Error("the scripted hand needs Memnite, the free creature that attacks on turn two while the one-drop taps for mana")
	}
}

// TestTutorialPlayerTriggersNeverShareAnEvent is #3012's invariant: no
// two triggered abilities of the player's practice deck, commander
// included, can trigger on one event. Two triggers of one player's on
// one event raise the CR 603.3b ordering prompt ("Order your
// triggers"), which the tutorial never explains. Marwyn and Wolverine
// Riders raised it on every Elf the player cast.
//
// It reads each card's printed text from the oracle fixture the catalog
// keeps in the repo (internal/cards/coverage, testdata/oracle), so it
// runs without a Scryfall dump. Reminder text is kept, because a keyword
// trigger such as evolve is only spelled out there. The fixture carries
// no type line, so the test is strict where it cannot see: a trigger on
// "another Elf" entering counts as a trigger on any nonland permanent
// entering. That is why the deck has no "when this creature enters" at
// all beside its commander's trigger.
func TestTutorialPlayerTriggersNeverShareAnEvent(t *testing.T) {
	d := TutorialPlayer()
	var all []tutorialTrigger
	for _, c := range d.Cards() {
		if c.Basic {
			continue
		}
		b, err := os.ReadFile(filepath.Join("..", "cards", "coverage", oraclefix.OracleFixtureFile(c.OracleID)))
		if err != nil {
			t.Errorf("%s: no oracle fixture to read its triggers from: %v", c.Name, err)
			continue
		}
		var oc oraclefix.OracleCard
		if err := json.Unmarshal(b, &oc); err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		for _, f := range oc.AllFaces() {
			all = append(all, triggersIn(c.Name, f.Text)...)
		}
	}
	// The classifier has to see the trigger that started this, or a
	// change to the text it reads would pass by finding nothing.
	commander := 0
	for _, tr := range all {
		if tr.card == d.Commander.Name && tr.event == "enters" && !tr.self {
			commander++
		}
	}
	if commander != 1 {
		t.Fatalf("found %d enters triggers on %s, want 1; the trigger classifier no longer reads the fixture",
			commander, d.Commander.Name)
	}
	for i, a := range all {
		for _, b := range all[i+1:] {
			if a.collides(b) {
				t.Errorf("%s (%q) and %s (%q) can trigger on one event; the player would be asked to order them "+
					"(CR 603.3b), which the tutorial never explains", a.card, a.cond, b.card, b.cond)
			}
		}
	}
}

// TestTutorialTriggerClassifier pins the reading the deck test relies
// on, against the cards #3012 took out of the deck.
func TestTutorialTriggerClassifier(t *testing.T) {
	marwyn := triggersIn("Marwyn, the Nurturer", "Whenever another Elf you control enters, put a +1/+1 counter on Marwyn.\n{T}: Add an amount of {G} equal to Marwyn's power.")
	cases := []struct {
		card, text string
		collides   bool
	}{
		{"Wolverine Riders", "At the beginning of each upkeep, create a 1/1 green Elf Warrior creature token.\nWhenever another Elf you control enters, you gain life equal to its toughness.", true},
		{"Elvish Visionary", "When this creature enters, draw a card.", true},
		{"Gyre Sage", "Evolve (Whenever a creature you control enters, if that creature has greater power or toughness than this creature, put a +1/+1 counter on this creature.)\n{T}: Add {G} for each +1/+1 counter on this creature.", true},
		{"Sporemound", "Landfall — Whenever a land you control enters, create a 1/1 green Saproling creature token.", false},
		{"Beast Whisperer", "Whenever you cast a creature spell, draw a card.", false},
		{"Llanowar Elves", "{T}: Add {G}.", false},
	}
	for _, tc := range cases {
		got := false
		for _, a := range marwyn {
			for _, b := range triggersIn(tc.card, tc.text) {
				got = got || a.collides(b)
			}
		}
		if got != tc.collides {
			t.Errorf("%s beside Marwyn: collides = %v, want %v", tc.card, got, tc.collides)
		}
	}
	// Two landfall triggers fire on the same land; two "this creature
	// enters" triggers on different cards do not.
	sporemound := triggersIn("Sporemound", "Landfall — Whenever a land you control enters, create a 1/1 green Saproling creature token.")
	baloths := triggersIn("Rampaging Baloths", "Trample\nLandfall — Whenever a land you control enters, create a 4/4 green Beast creature token.")
	if !sporemound[0].collides(baloths[0]) {
		t.Error("two landfall triggers should collide")
	}
	wall := triggersIn("Wall of Blossoms", "Defender\nWhen this creature enters, draw a card.")
	visionary := triggersIn("Elvish Visionary", "When this creature enters, draw a card.")
	if wall[0].collides(visionary[0]) {
		t.Error("two cards' own enters triggers fire on different events")
	}
	upkeep := triggersIn("Bugenhagen, Wise Elder", "At the beginning of your upkeep, if you control a creature with power 7 or greater, draw a card.")
	if len(upkeep) != 1 || upkeep[0].event != "beginning:upkeep" || !upkeep[0].collides(triggersIn("Wolverine Riders", "At the beginning of each upkeep, create a token.")[0]) {
		t.Errorf("two upkeep triggers should collide: %+v", upkeep)
	}
}

// tutorialTrigger is one trigger condition read off a card's text: the
// event it waits for, and what has to be involved in it.
type tutorialTrigger struct {
	card string
	// cond is the condition as printed, up to its first comma.
	cond string
	// event is the family of events it fires on: "enters", "leaves",
	// "cast", "beginning:<step>", "attacks", or the condition itself.
	event string
	// subject is, for "enters", what enters: "land", "nonland" or "any".
	subject string
	// self: only the card's own event fires it ("When this creature
	// enters"). Two cards' self triggers fire on one event only when both
	// cards take part in it at once, which nothing in the deck does.
	self bool
}

// triggerCondition finds every "When …", "Whenever …" and "At the
// beginning of …" condition, reminder text included.
var triggerCondition = regexp.MustCompile(`\b(When(?:ever)?|At the beginning of)\b[^,.]*`)

// beginningSteps names the steps an "At the beginning of …" trigger can
// wait for. "end of combat" is before "combat", which it contains.
var beginningSteps = []string{
	"precombat main phase", "postcombat main phase", "end step", "draw step", "upkeep", "end of combat", "combat",
}

// triggersIn reads every trigger condition out of one face's text.
func triggersIn(card, text string) []tutorialTrigger {
	var out []tutorialTrigger
	short, _, _ := strings.Cut(card, ",")
	for _, cond := range triggerCondition.FindAllString(text, -1) {
		cond = strings.TrimSpace(cond)
		lower := strings.ToLower(cond)
		tr := tutorialTrigger{card: card, cond: cond}
		tr.self = strings.Contains(lower, "this creature") || strings.Contains(lower, "this card") ||
			strings.Contains(lower, "this permanent") || strings.Contains(cond, short)
		switch {
		case strings.HasPrefix(cond, "At the beginning of"):
			tr.event = "beginning:" + lower
			for _, s := range beginningSteps {
				if strings.Contains(lower, s) {
					tr.event = "beginning:" + s
					break
				}
			}
		case strings.Contains(lower, " enters"):
			tr.event = "enters"
			switch {
			case strings.Contains(lower, "permanent"):
				tr.subject = "any"
			case strings.Contains(lower, " land") && !strings.Contains(lower, "nonland"):
				tr.subject = "land"
			default:
				tr.subject = "nonland"
			}
		case strings.Contains(lower, "leaves the battlefield") || strings.Contains(lower, " dies") ||
			strings.Contains(lower, "put into a graveyard"):
			tr.event = "leaves"
		case strings.Contains(lower, "cast"):
			tr.event = "cast"
		case strings.Contains(lower, "attacks"):
			tr.event = "attacks"
		default:
			tr.event = lower
		}
		out = append(out, tr)
	}
	return out
}

// collides reports whether a and b can trigger on one event.
func (a tutorialTrigger) collides(b tutorialTrigger) bool {
	if a.event != b.event {
		return false
	}
	if a.self && b.self && a.card != b.card {
		return false
	}
	if a.subject == "" || b.subject == "" || a.subject == "any" || b.subject == "any" {
		return true
	}
	return a.subject == b.subject
}
