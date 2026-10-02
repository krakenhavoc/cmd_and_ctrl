package decks

import (
	"strings"
	"testing"

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
		t.Errorf("only %d cards in the player's tutorial deck have a right-click ability; step 7 needs one in play", rightClick)
	}
}
