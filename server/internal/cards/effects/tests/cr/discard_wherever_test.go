package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	before func(t *testing.T, d effectstest.DboTable)
	// discard runs the discard; nil means a plain Mind Rot.
	discard func(t *testing.T, d effectstest.DboTable, card, filler uuid.UUID)
	// check asserts the ability's own result.
	check func(t *testing.T, d effectstest.DboTable, card uuid.UUID)
}

func TestTheDiscardFamilyTriggersFromExileUnderRestInPeace(t *testing.T) {
	cases := []restInPeaceCase{
		{
			name: "Ajani's Last Stand", oracle: effectstest.AjanisLastStandOracle, typeLine: "Enchantment", cost: "{2}{W}{W}",
			before: func(t *testing.T, d effectstest.DboTable) {
				d.Permanent(d.Victim, "Plains", "Basic Land — Plains", "")
			},
			check: func(t *testing.T, d effectstest.DboTable, _ uuid.UUID) {
				if n := d.CountNamedUnder(d.Victim, "Avatar"); n != 1 {
					t.Errorf("Avatar tokens = %d, want 1", n)
				}
			},
		},
		{
			name: "Metrognome", oracle: effectstest.MetrognomeOracle, typeLine: "Artifact", cost: "{4}",
			check: func(t *testing.T, d effectstest.DboTable, _ uuid.UUID) {
				if n := d.CountNamedUnder(d.Victim, "Gnome"); n != 4 {
					t.Errorf("Gnome tokens = %d, want 4", n)
				}
			},
		},
		{
			name: "Guerrilla Tactics", oracle: effectstest.GuerrillaTacticsOracle, typeLine: "Instant", cost: "{1}{R}",
			discard: func(t *testing.T, d effectstest.DboTable, card, filler uuid.UUID) {
				effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
					[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
				effectstest.PassPriorityAroundTable(t, d.G)
				effectstest.AnswerDiscard(t, d.G, d.Victim.ID, card, filler)
				effectstest.PassPriorityAroundTable(t, d.G)
				effectstest.PickTargetRefs(t, d.G, d.Victim.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: d.Caster.ID})
				effectstest.PassPriorityAroundTable(t, d.G)
			},
			check: func(t *testing.T, d effectstest.DboTable, _ uuid.UUID) {
				if d.Caster.Life != 40-4 {
					t.Errorf("caster life = %d, want %d", d.Caster.Life, 40-4)
				}
			},
		},
		{
			name: "Psychic Purge", oracle: effectstest.PsychicPurgeOracle, typeLine: "Sorcery", cost: "{U}",
			check: func(t *testing.T, d effectstest.DboTable, _ uuid.UUID) {
				if d.Caster.Life != 40-5 {
					t.Errorf("caster life = %d, want %d", d.Caster.Life, 40-5)
				}
			},
		},
		{
			name: "Mangara's Blessing", oracle: effectstest.MangarasBlessingOracle, typeLine: "Instant", cost: "{2}{W}",
			check: func(t *testing.T, d effectstest.DboTable, card uuid.UUID) {
				if d.Victim.Life != 40+2 {
					t.Errorf("victim life = %d, want %d", d.Victim.Life, 40+2)
				}
				d.ToEndStep(t)
				// "from your graveyard": the card never was in one, so
				// the delayed return finds nothing.
				if got := d.ZoneOf(card); got != "exile" {
					t.Errorf("at the end step the card is in %q, want exile", got)
				}
			},
		},
		{
			name: "Pure Intentions", oracle: effectstest.PureIntentionsOracle, typeLine: "Instant — Arcane", cost: "{W}",
			check: func(t *testing.T, d effectstest.DboTable, card uuid.UUID) {
				if n := triggerFrom(d.G, card); n != 1 {
					t.Errorf("the discard ability triggered %d times, want 1", n)
				}
				d.ToEndStep(t)
				if got := d.ZoneOf(card); got != "exile" {
					t.Errorf("at the end step the card is in %q, want exile", got)
				}
			},
		},
		{
			name: "Sand Golem", oracle: effectstest.SandGolemOracle, typeLine: "Artifact Creature — Golem", cost: "{5}", power: 3, tough: 3,
			check: func(t *testing.T, d effectstest.DboTable, card uuid.UUID) {
				if n := triggerFrom(d.G, card); n != 1 {
					t.Errorf("the discard ability triggered %d times, want 1", n)
				}
				d.ToEndStep(t)
				if got := d.ZoneOf(card); got != "exile" {
					t.Errorf("at the end step the card is in %q, want exile (nothing in a graveyard to return)", got)
				}
			},
		},
		{
			name: "Quagnoth", oracle: effectstest.QuagnothOracle, typeLine: "Creature — Beast", cost: "{5}{G}", power: 4, tough: 5,
			check: func(t *testing.T, d effectstest.DboTable, card uuid.UUID) {
				if !d.Victim.Hand.Contains(card) {
					t.Errorf("Quagnoth is in %q, want the hand: \"return it to your hand\" names no zone", d.ZoneOf(card))
				}
			},
		},
		{
			name: "Orvar, the All-Form", oracle: effectstest.OrvarOracle, typeLine: "Legendary Creature — Shapeshifter", cost: "{3}{U}", power: 3, tough: 3,
			before: func(t *testing.T, d effectstest.DboTable) {
				d.Permanent(d.Caster, "Copyable Ox", "Creature — Ox", "")
			},
			discard: func(t *testing.T, d effectstest.DboTable, card, filler uuid.UUID) {
				effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
					[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
				effectstest.PassPriorityAroundTable(t, d.G)
				effectstest.AnswerDiscard(t, d.G, d.Victim.ID, card, filler)
				effectstest.PassPriorityAroundTable(t, d.G)
				var ox uuid.UUID
				for _, c := range d.G.Battlefield.Cards {
					if c.Name == "Copyable Ox" {
						ox = c.InstanceID
					}
				}
				effectstest.PickTargetRefs(t, d.G, d.Victim.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: ox})
				effectstest.PassPriorityAroundTable(t, d.G)
			},
			check: func(t *testing.T, d effectstest.DboTable, _ uuid.UUID) {
				if n := d.CountNamedUnder(d.Victim, "Copyable Ox"); n != 1 {
					t.Errorf("token copies of the Ox = %d, want 1", n)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := effectstest.NewDBOTable(t)
			effectstest.CastRestInPeace(t, d.G)
			card := d.HandCard(d.Victim, tc.name, tc.typeLine, tc.cost, tc.oracle, tc.power, tc.tough)
			filler := d.Filler(d.Victim)
			if tc.before != nil {
				tc.before(t, d)
			}
			if tc.discard != nil {
				tc.discard(t, d, card, filler)
			} else {
				d.TheyDiscard(t, card, filler)
			}
			if got := d.ZoneOf(card); got != "exile" && tc.name != "Quagnoth" {
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
	setup := func(t *testing.T, oracle, name, typeLine, cost string, power, tough int) (effectstest.DboTable, uuid.UUID) {
		d := effectstest.NewDBOTable(t)
		effectstest.PushTokenReplacementCard(d.G, effectstest.LibraryOfLengOracle, "Library of Leng", "Artifact", d.Victim.ID)
		card := d.HandCard(d.Victim, name, typeLine, cost, oracle, power, tough)
		filler := d.Filler(d.Victim)
		effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
		effectstest.PassPriorityAroundTable(t, d.G)
		effectstest.AnswerDiscard(t, d.G, d.Victim.ID, card, filler)
		// Library of Leng asks once per card; take it for the card
		// under test and decline the filler.
		for len(d.G.PendingChoices) > 0 {
			pc := d.G.PendingChoices[0]
			if pc.Kind != game.PendingChoiceOptionalReplacement {
				break
			}
			accept := d.Victim.Hand.Contains(card)
			if err := d.G.ResolveOptionalReplacement(pc.ID, d.Victim.ID, accept); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
		}
		if !d.Victim.Library.Contains(card) {
			t.Fatalf("setup: the card is in %q, want the library", d.ZoneOf(card))
		}
		return d, card
	}

	t.Run("Quagnoth returns from the top of the library", func(t *testing.T) {
		d, card := setup(t, effectstest.QuagnothOracle, "Quagnoth", "Creature — Beast", "{5}{G}", 4, 5)
		effectstest.PassPriorityAroundTable(t, d.G)
		if !d.Victim.Hand.Contains(card) {
			t.Errorf("Quagnoth is in %q, want the hand", d.ZoneOf(card))
		}
	})
	t.Run("Mangara's Blessing gains its life", func(t *testing.T) {
		life := 40
		d, card := setup(t, effectstest.MangarasBlessingOracle, "Mangara's Blessing", "Instant", "{2}{W}", 0, 0)
		effectstest.PassPriorityAroundTable(t, d.G)
		if d.Victim.Life != life+2 {
			t.Errorf("victim life = %d, want %d", d.Victim.Life, life+2)
		}
		if !d.Victim.Library.Contains(card) {
			t.Errorf("the card is in %q, want it left in the library", d.ZoneOf(card))
		}
	})
}

// Madness exiles the discard; the card's own discard ability still
// triggers from exile.
func TestADiscardAbilityTriggersFromMadnessExile(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	card := d.HandCard(d.Victim, "Madness Discard Probe", "Creature — Test", "{3}{R}", effectstest.MadnessProbeOracle, 2, 2)
	filler := d.Filler(d.Victim)
	effectstest.MadnessProbeFired = 0

	d.TheyDiscard(t, card, filler)

	if got := d.ZoneOf(card); got != "exile" {
		t.Fatalf("a madness discard is in %q, want exile", got)
	}
	// The card's own ability and the madness trigger went on the stack
	// together, so its owner orders them; the madness offer holds
	// priority until it is answered, so decline it.
	if order := effectstest.LatestChoiceOfKind(d.G, game.PendingChoiceTriggerOrder); order != nil {
		if err := d.G.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, d.G)
	offer := effectstest.LatestChoiceOfKind(d.G, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("a madness discard offered no cast")
	}
	if err := d.G.ResolvePendingChoice(offer.ID, d.Victim.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, d.G)
	if effectstest.MadnessProbeFired != 1 {
		t.Errorf("the discard ability fired %d times, want 1", effectstest.MadnessProbeFired)
	}
}

// A card that left the graveyard and came back before the trigger resolved
// is a different object (CR 400.7), so "return this card from your
// graveyard" does not follow it.
func TestAReturnFromTheGraveyardDoesNotFollowANewObject(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	card := d.HandCard(d.Victim, "Mangara's Blessing", "Instant", "{2}{W}", effectstest.MangarasBlessingOracle, 0, 0)
	filler := d.Filler(d.Victim)

	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)
	effectstest.AnswerDiscard(t, d.G, d.Victim.ID, card, filler)
	// The trigger is queued; before it resolves the card is exiled and
	// buried again, which is the path a declined madness cast takes.
	d.G.WithWriteLock(func() {
		if err := d.G.ExileCardForEffect(card); err != nil {
			t.Fatalf("exile: %v", err)
		}
		if err := d.G.PutIntoGraveyardForEffect(card); err != nil {
			t.Fatalf("bury: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, d.G)
	d.ToEndStep(t)

	if d.Victim.Hand.Contains(card) {
		t.Error("the card came back to the hand: the ability followed a different object")
	}
}
