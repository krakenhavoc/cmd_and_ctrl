package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_wherever_test.go — #3859, ADR 0153's 2026-10-15 amendment. A
// card discarded is discarded wherever it goes (CR 701.9a), and "when a
// spell or ability an opponent controls causes you to discard this card"
// looks back at that move (CR 603.10a). So madness (CR 702.35a), Rest in
// Peace and Library of Leng sending the card to exile or a library do not
// stop its own ability. These tests run the ten-card family under Rest in
// Peace (the discard lands in exile), and then the other zones.

const (
	madnessProbeOracle = "test-discard-madness-probe"
)

var madnessProbeFired int

func init() {
	// A madness card whose own discard ability counts its firings, so the
	// two ways a discard can end in exile (madness and Rest in Peace) are
	// proved against the real madness machinery and not a stand-in.
	Register(Spec{
		OracleID: madnessProbeOracle,
		Name:     "Madness Discard Probe",
		Madness:  "{R}",
		Triggered: []game.TriggeredAbility{
			WhereverDiscarded(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Madness Discard Probe — discarded",
				func(*game.Game, *game.StackItem) error { madnessProbeFired++; return nil })),
		},
	})
}

// triggerFrom counts the EventTrigger entries a source card put on the
// stack.
func triggerFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventTrigger && ev.Source == source {
			n++
		}
	}
	return n
}

// restInPeaceDiscard is the family's one setup: Rest in Peace is on the
// table, the card is in the victim's hand beside a filler, and an
// opponent's Mind Rot makes the victim discard both.
type restInPeaceCase struct {
	name     string
	oracle   string
	typeLine string
	cost     string
	power    int
	tough    int
	// before runs after the table is built and the card is in hand.
	before func(t *testing.T, d dboTable)
	// discard runs the discard; nil means a plain Mind Rot.
	discard func(t *testing.T, d dboTable, card, filler uuid.UUID)
	// check asserts the ability's own result.
	check func(t *testing.T, d dboTable, card uuid.UUID)
}

func TestTheDiscardFamilyTriggersFromExileUnderRestInPeace(t *testing.T) {
	cases := []restInPeaceCase{
		{
			name: "Ajani's Last Stand", oracle: ajanisLastStandOracle, typeLine: "Enchantment", cost: "{2}{W}{W}",
			before: func(t *testing.T, d dboTable) {
				d.permanent(d.victim, "Plains", "Basic Land — Plains", "")
			},
			check: func(t *testing.T, d dboTable, _ uuid.UUID) {
				if n := d.countNamedUnder(d.victim, "Avatar"); n != 1 {
					t.Errorf("Avatar tokens = %d, want 1", n)
				}
			},
		},
		{
			name: "Metrognome", oracle: metrognomeOracle, typeLine: "Artifact", cost: "{4}",
			check: func(t *testing.T, d dboTable, _ uuid.UUID) {
				if n := d.countNamedUnder(d.victim, "Gnome"); n != 4 {
					t.Errorf("Gnome tokens = %d, want 4", n)
				}
			},
		},
		{
			name: "Guerrilla Tactics", oracle: guerrillaTacticsOracle, typeLine: "Instant", cost: "{1}{R}",
			discard: func(t *testing.T, d dboTable, card, filler uuid.UUID) {
				castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
					[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
				passPriorityAroundTable(t, d.g)
				answerDiscard(t, d.g, d.victim.ID, card, filler)
				passPriorityAroundTable(t, d.g)
				pickTargetRefs(t, d.g, d.victim.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: d.caster.ID})
				passPriorityAroundTable(t, d.g)
			},
			check: func(t *testing.T, d dboTable, _ uuid.UUID) {
				if d.caster.Life != 40-4 {
					t.Errorf("caster life = %d, want %d", d.caster.Life, 40-4)
				}
			},
		},
		{
			name: "Psychic Purge", oracle: psychicPurgeOracle, typeLine: "Sorcery", cost: "{U}",
			check: func(t *testing.T, d dboTable, _ uuid.UUID) {
				if d.caster.Life != 40-5 {
					t.Errorf("caster life = %d, want %d", d.caster.Life, 40-5)
				}
			},
		},
		{
			name: "Mangara's Blessing", oracle: mangarasBlessingOracle, typeLine: "Instant", cost: "{2}{W}",
			check: func(t *testing.T, d dboTable, card uuid.UUID) {
				if d.victim.Life != 40+2 {
					t.Errorf("victim life = %d, want %d", d.victim.Life, 40+2)
				}
				d.toEndStep(t)
				// "from your graveyard": the card never was in one, so
				// the delayed return finds nothing.
				if got := d.zoneOf(card); got != "exile" {
					t.Errorf("at the end step the card is in %q, want exile", got)
				}
			},
		},
		{
			name: "Pure Intentions", oracle: pureIntentionsOracle, typeLine: "Instant — Arcane", cost: "{W}",
			check: func(t *testing.T, d dboTable, card uuid.UUID) {
				if n := triggerFrom(d.g, card); n != 1 {
					t.Errorf("the discard ability triggered %d times, want 1", n)
				}
				d.toEndStep(t)
				if got := d.zoneOf(card); got != "exile" {
					t.Errorf("at the end step the card is in %q, want exile", got)
				}
			},
		},
		{
			name: "Sand Golem", oracle: sandGolemOracle, typeLine: "Artifact Creature — Golem", cost: "{5}", power: 3, tough: 3,
			check: func(t *testing.T, d dboTable, card uuid.UUID) {
				if n := triggerFrom(d.g, card); n != 1 {
					t.Errorf("the discard ability triggered %d times, want 1", n)
				}
				d.toEndStep(t)
				if got := d.zoneOf(card); got != "exile" {
					t.Errorf("at the end step the card is in %q, want exile (nothing in a graveyard to return)", got)
				}
			},
		},
		{
			name: "Quagnoth", oracle: quagnothOracle, typeLine: "Creature — Beast", cost: "{5}{G}", power: 4, tough: 5,
			check: func(t *testing.T, d dboTable, card uuid.UUID) {
				if !d.victim.Hand.Contains(card) {
					t.Errorf("Quagnoth is in %q, want the hand: \"return it to your hand\" names no zone", d.zoneOf(card))
				}
			},
		},
		{
			name: "Orvar, the All-Form", oracle: orvarOracle, typeLine: "Legendary Creature — Shapeshifter", cost: "{3}{U}", power: 3, tough: 3,
			before: func(t *testing.T, d dboTable) {
				d.permanent(d.caster, "Copyable Ox", "Creature — Ox", "")
			},
			discard: func(t *testing.T, d dboTable, card, filler uuid.UUID) {
				castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
					[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
				passPriorityAroundTable(t, d.g)
				answerDiscard(t, d.g, d.victim.ID, card, filler)
				passPriorityAroundTable(t, d.g)
				var ox uuid.UUID
				for _, c := range d.g.Battlefield.Cards {
					if c.Name == "Copyable Ox" {
						ox = c.InstanceID
					}
				}
				pickTargetRefs(t, d.g, d.victim.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: ox})
				passPriorityAroundTable(t, d.g)
			},
			check: func(t *testing.T, d dboTable, _ uuid.UUID) {
				if n := d.countNamedUnder(d.victim, "Copyable Ox"); n != 1 {
					t.Errorf("token copies of the Ox = %d, want 1", n)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDBOTable(t)
			castRestInPeace(t, d.g)
			card := d.handCard(d.victim, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.power, tc.tough)
			filler := d.filler(d.victim)
			if tc.before != nil {
				tc.before(t, d)
			}
			if tc.discard != nil {
				tc.discard(t, d, card, filler)
			} else {
				d.theyDiscard(t, card, filler)
			}
			if got := d.zoneOf(card); got != "exile" && tc.name != "Quagnoth" {
				t.Fatalf("setup: the discarded card is in %q, want exile (Rest in Peace)", got)
			}
			tc.check(t, d, card)
		})
	}
}

// Library of Leng puts the discarded card on top of its owner's library.
// It was discarded all the same, so Quagnoth still comes back to hand and
// Mangara's Blessing still gains its life.
func TestTheDiscardFamilyTriggersFromTheLibraryUnderLibraryOfLeng(t *testing.T) {
	setup := func(t *testing.T, oracle, name, typeLine, cost string, power, tough int) (dboTable, uuid.UUID) {
		d := newDBOTable(t)
		pushTokenReplacementCard(d.g, libraryOfLengOracle, "Library of Leng", "Artifact", d.victim.ID)
		card := d.handCard(d.victim, name, typeLine, cost, oracle, power, tough)
		filler := d.filler(d.victim)
		castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
		passPriorityAroundTable(t, d.g)
		answerDiscard(t, d.g, d.victim.ID, card, filler)
		// Library of Leng asks once per card; take it for the card
		// under test and decline the filler.
		for len(d.g.PendingChoices) > 0 {
			pc := d.g.PendingChoices[0]
			if pc.Kind != game.PendingChoiceOptionalReplacement {
				break
			}
			accept := d.victim.Hand.Contains(card)
			if err := d.g.ResolveOptionalReplacement(pc.ID, d.victim.ID, accept); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
		}
		if !d.victim.Library.Contains(card) {
			t.Fatalf("setup: the card is in %q, want the library", d.zoneOf(card))
		}
		return d, card
	}

	t.Run("Quagnoth returns from the top of the library", func(t *testing.T) {
		d, card := setup(t, quagnothOracle, "Quagnoth", "Creature — Beast", "{5}{G}", 4, 5)
		passPriorityAroundTable(t, d.g)
		if !d.victim.Hand.Contains(card) {
			t.Errorf("Quagnoth is in %q, want the hand", d.zoneOf(card))
		}
	})
	t.Run("Mangara's Blessing gains its life", func(t *testing.T) {
		life := 40
		d, card := setup(t, mangarasBlessingOracle, "Mangara's Blessing", "Instant", "{2}{W}", 0, 0)
		passPriorityAroundTable(t, d.g)
		if d.victim.Life != life+2 {
			t.Errorf("victim life = %d, want %d", d.victim.Life, life+2)
		}
		if !d.victim.Library.Contains(card) {
			t.Errorf("the card is in %q, want it left in the library", d.zoneOf(card))
		}
	})
}

// Madness exiles the discard; the card's own discard ability still
// triggers from exile.
func TestADiscardAbilityTriggersFromMadnessExile(t *testing.T) {
	d := newDBOTable(t)
	card := d.handCard(d.victim, "Madness Discard Probe", "Creature — Test", "{3}{R}", madnessProbeOracle, 2, 2)
	filler := d.filler(d.victim)
	madnessProbeFired = 0

	d.theyDiscard(t, card, filler)

	if got := d.zoneOf(card); got != "exile" {
		t.Fatalf("a madness discard is in %q, want exile", got)
	}
	// The card's own ability and the madness trigger went on the stack
	// together, so its owner orders them; the madness offer holds
	// priority until it is answered, so decline it.
	if order := latestChoiceOfKind(d.g, game.PendingChoiceTriggerOrder); order != nil {
		if err := d.g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	passPriorityAroundTable(t, d.g)
	offer := latestChoiceOfKind(d.g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatal("a madness discard offered no cast")
	}
	if err := d.g.ResolveMayCast(offer.ID, d.victim.ID, false); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	passPriorityAroundTable(t, d.g)
	if madnessProbeFired != 1 {
		t.Errorf("the discard ability fired %d times, want 1", madnessProbeFired)
	}
}

// A card that left the graveyard and came back before the trigger resolved
// is a different object (CR 400.7), so "return this card from your
// graveyard" does not follow it.
func TestAReturnFromTheGraveyardDoesNotFollowANewObject(t *testing.T) {
	d := newDBOTable(t)
	card := d.handCard(d.victim, "Mangara's Blessing", "Instant", "{2}{W}", mangarasBlessingOracle, 0, 0)
	filler := d.filler(d.victim)

	castCatalogSpell(t, d.g, "Mind Rot", "Sorcery", mindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.victim.ID}})
	passPriorityAroundTable(t, d.g)
	answerDiscard(t, d.g, d.victim.ID, card, filler)
	// The trigger is queued; before it resolves the card is exiled and
	// buried again, which is the path a declined madness cast takes.
	d.g.WithWriteLock(func() {
		if err := d.g.ExileCardForEffect(card); err != nil {
			t.Fatalf("exile: %v", err)
		}
		if err := d.g.PutIntoGraveyardForEffect(card); err != nil {
			t.Fatalf("bury: %v", err)
		}
	})
	passPriorityAroundTable(t, d.g)
	d.toEndStep(t)

	if d.victim.Hand.Contains(card) {
		t.Error("the card came back to the hand: the ability followed a different object")
	}
}
