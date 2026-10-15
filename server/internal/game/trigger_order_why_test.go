package game

import (
	"strings"
	"testing"
)

// trigger_order_why_test.go — #3015: the "Order your triggers" prompt
// says why it was asked. One sentence per reason, in the player's
// terms; the board-driven cases for the reasons a real board can reach
// are in cards/effects/trigger_order_why_3015_test.go.

// Every reason the engine can give has a sentence, and the sentence
// reads as the owner's examples do: it names the card, says what it
// does in plain words, and uses none of the check's vocabulary.
func TestEveryTriggerOrderReasonHasASentence(t *testing.T) {
	g := newActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]

	cases := map[TriggerOrderReason][]struct {
		name string
		why  triggerOrderWhy
		want string
	}{
		TriggerOrderReasonAlways: {
			{"", triggerOrderWhy{reason: TriggerOrderReasonAlways}, "You've set your triggers to ask every time."},
		},
		TriggerOrderReasonUndescribed: {
			{"names the card", triggerOrderWhy{reason: TriggerOrderReasonUndescribed, card: "Wolverine Riders"},
				"Wolverine Riders' trigger is handled by the players, so the game can't tell whether the order matters."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonUndescribed},
				"The game can't tell whether the order of these triggers matters."},
		},
		TriggerOrderReasonBoardOpaque: {
			{"names the card", triggerOrderWhy{reason: TriggerOrderReasonBoardOpaque, card: "Humility"},
				"Humility could change how these triggers resolve."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonBoardOpaque},
				"Something in play could change how these triggers resolve."},
		},
		TriggerOrderReasonReactor: {
			{"life totals", triggerOrderWhy{reason: TriggerOrderReasonReactor, card: "Ajani's Pridemate", event: EventChangeLife},
				"Ajani's Pridemate watches life totals."},
			{"life totals, whatever direction the batch moves them", triggerOrderWhy{reason: TriggerOrderReasonReactor, card: "Sanguine Bond", event: EventChangeLife, life: lifeUp},
				"Sanguine Bond watches life totals."},
			{"a draw", triggerOrderWhy{reason: TriggerOrderReasonReactor, card: "Chasm Skulker", event: EventDrawCard},
				"Chasm Skulker watches cards being drawn."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonReactor, event: EventETB},
				"Something in play watches permanents entering the battlefield."},
			{"a token the batch makes", triggerOrderWhy{reason: TriggerOrderReasonReactor, event: EventETB, token: true},
				"A token these triggers create watches permanents entering the battlefield."},
			{"an event with no words", triggerOrderWhy{reason: TriggerOrderReasonReactor, card: "Mystery", event: "no_such_event"},
				"Mystery watches what these triggers do."},
		},
		TriggerOrderReasonRipple: {
			{"names the card", triggerOrderWhy{reason: TriggerOrderReasonRipple, card: "Oblivion Ring"},
				"Something else in play is keeping track of Oblivion Ring."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonRipple},
				"Something else in play is keeping track of what these triggers change."},
		},
		TriggerOrderReasonElimination: {
			{"the chooser", triggerOrderWhy{reason: TriggerOrderReasonElimination, player: me.ID},
				"You could leave the game partway through these triggers, which changes what the rest do."},
			{"another player", triggerOrderWhy{reason: TriggerOrderReasonElimination, player: other.ID},
				other.Name + " could leave the game partway through these triggers, which changes what the rest do."},
			{"nobody found", triggerOrderWhy{reason: TriggerOrderReasonElimination},
				"A player could leave the game partway through these triggers, which changes what the rest do."},
		},
		TriggerOrderReasonObjectTie: {
			{"both change it", triggerOrderWhy{reason: TriggerOrderReasonObjectTie, card: "Marwyn", both: true},
				"These triggers both change Marwyn."},
			{"one reads it", triggerOrderWhy{reason: TriggerOrderReasonObjectTie, card: "Marwyn"},
				"One of these triggers changes Marwyn, which another looks at."},
			{"no name", triggerOrderWhy{reason: TriggerOrderReasonObjectTie, both: true},
				"These triggers both change the same permanent."},
		},
		TriggerOrderReasonRules: {
			{"names the card", triggerOrderWhy{reason: TriggerOrderReasonRules, card: "Sokka"},
				"Sokka's trigger changes the rules the other triggers resolve under."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonRules},
				"One of these triggers changes the rules the others resolve under."},
		},
		TriggerOrderReasonLibrary: {
			{"", triggerOrderWhy{reason: TriggerOrderReasonLibrary}, "These triggers both work on the same library or graveyard."},
		},
		TriggerOrderReasonMonarch: {
			{"", triggerOrderWhy{reason: TriggerOrderReasonMonarch}, "These triggers both make you the monarch."},
		},
		TriggerOrderReasonCharacteristic: {
			{"copies of one trigger", triggerOrderWhy{reason: TriggerOrderReasonCharacteristic, card: "Righteous Valkyrie", life: lifeUp, identical: true},
				"The first life gain can change Righteous Valkyrie's bonus, which another trigger counts."},
			{"different triggers", triggerOrderWhy{reason: TriggerOrderReasonCharacteristic, card: "Righteous Valkyrie", life: lifeUp},
				"The life gain from one trigger can change Righteous Valkyrie's bonus, which another trigger counts."},
			{"not life", triggerOrderWhy{reason: TriggerOrderReasonCharacteristic, card: "Elvish Archdruid"},
				"What one of these triggers does can change Elvish Archdruid's bonus, which another trigger counts."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonCharacteristic},
				"What one of these triggers does can change a bonus that another trigger counts."},
		},
		TriggerOrderReasonCondition: {
			{"card", triggerOrderWhy{reason: TriggerOrderReasonCondition, card: "Field of the Dead"},
				"Field of the Dead's trigger checks its \"if\" again as it resolves, and another of these triggers can change the answer."},
			{"no card", triggerOrderWhy{reason: TriggerOrderReasonCondition},
				"One of these triggers checks its \"if\" again as it resolves, and another can change the answer."},
		},
	}

	// Nothing the engine can say is left without a sentence here.
	for _, r := range TriggerOrderReasons() {
		if len(cases[r]) == 0 {
			t.Errorf("reason %q has no sentence case", r)
		}
	}

	banned := []string{"footprint", "characteristic", "batch", "engine", "undescribed", "opaque", "ripple", "reactor", "object_tie"}
	for r, cs := range cases {
		for _, c := range cs {
			t.Run(string(r)+"/"+c.name, func(t *testing.T) {
				var got string
				g.WithWriteLock(func() { got = g.triggerOrderSentenceLocked(c.why, me.ID) })
				if got != c.want {
					t.Errorf("sentence\n got %q\nwant %q", got, c.want)
				}
				for _, b := range banned {
					if strings.Contains(strings.ToLower(got), b) {
						t.Errorf("%q uses the word %q", got, b)
					}
				}
			})
		}
	}
}

// No reason, no sentence; a reason this binary has no case for still
// gets a generic line rather than none.
func TestTriggerOrderSentenceFallbacks(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var none, unknown string
	g.WithWriteLock(func() {
		none = g.triggerOrderSentenceLocked(triggerOrderWhy{}, me.ID)
		unknown = g.triggerOrderSentenceLocked(triggerOrderWhy{reason: "some-future-reason"}, me.ID)
	})
	if none != "" {
		t.Errorf("no reason gave %q", none)
	}
	if unknown != "The order of these triggers can change how they resolve." {
		t.Errorf("unknown reason gave %q", unknown)
	}
}

// The sentence is stored on the prompt as it is queued, and it is the
// same reason the metric counted. A queue-time batch of two plain
// triggers is undescribed; "always" is said as the seat's own choice.
func TestTheQueuedPromptCarriesItsSentence(t *testing.T) {
	for _, c := range []struct {
		mode TriggerOrderMode
		want string
	}{
		{TriggerOrderWhenItMatters, "The game can't tell whether the order of these triggers matters."},
		{TriggerOrderAlways, "You've set your triggers to ask every time."},
	} {
		g := newActiveGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		if err := g.SetTriggerOrderPreference(me.ID, c.mode); err != nil {
			t.Fatal(err)
		}
		var log []string
		queueTwoDifferentForTest(g, me, &log)
		if drainForTest(g) {
			t.Fatalf("mode %q: the batch drained, want it held", c.mode)
		}
		prompt := findTriggerOrderPrompt(g, me.ID)
		if prompt == nil {
			t.Fatalf("mode %q: no prompt", c.mode)
		}
		if prompt.TriggerOrderWhy != c.want {
			t.Errorf("mode %q: prompt says %q, want %q", c.mode, prompt.TriggerOrderWhy, c.want)
		}
		// Asking again does not queue a second prompt or change the words.
		drainForTest(g)
		if again := findTriggerOrderPrompt(g, me.ID); again != prompt || again.TriggerOrderWhy != c.want {
			t.Errorf("mode %q: the held drain changed the prompt", c.mode)
		}
	}
}

// A legend is named by the part before the comma, and a plain card by
// its whole name.
func TestWhyNamesALegendByItsShortName(t *testing.T) {
	for _, c := range []struct {
		card Card
		want string
	}{
		{Card{Name: "Marwyn, the Nurturer", TypeLine: "Legendary Creature — Elf Druid"}, "Marwyn"},
		{Card{Name: "Ajani's Pridemate", TypeLine: "Creature — Cat Soldier"}, "Ajani's Pridemate"},
		{Card{Name: "Lotus Bloom, Reborn", TypeLine: "Artifact"}, "Lotus Bloom, Reborn"},
	} {
		if got := whyCardName(c.card); got != c.want {
			t.Errorf("%q named as %q, want %q", c.card.Name, got, c.want)
		}
	}
}
