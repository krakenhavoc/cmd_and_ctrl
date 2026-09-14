package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// summoning_sick_view_test.go — #530. `summoning_sick` on the wire
// used to be a bare "entered the battlefield this turn" flag, so the
// client greyed the {T} row of every Treasure, fetchland and mana
// rock that arrived this turn. CR 302.6 restricts a CREATURE's
// attack and its {T}/{Q} abilities; an artifact or a land has no
// such restriction, and the server never refused those activations —
// only the wire said it would.
//
// These assertions are the contract the client's `abilityBlocked`
// reads. They are deliberately at the view layer rather than on
// game.HasSummoningSickness alone, because view.go is where the
// creature guard was missing.

// seatFreshPermanent drops a permanent that entered the battlefield
// this turn under the first seat and returns its instance ID.
func seatFreshPermanent(g *game.Game, name, typeLine string, mutate func(*game.Card)) uuid.UUID {
	owner := g.Seats[0].ID
	c := game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
		// Every battlefield entry is stamped with this by
		// layer_listener.go, tokens included.
		SummonedThisTurn: true,
	}
	if mutate != nil {
		mutate(&c)
	}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// sickFlagOf reads `summoning_sick` off one battlefield permanent in
// the rendered view.
func sickFlagOf(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	v := ViewOfGame(g)
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c.SummoningSick
		}
	}
	t.Fatalf("permanent %s missing from the battlefield view", id)
	return false
}

// TestSummoningSickIsCreaturesOnly is the #530 regression: a
// noncreature permanent that entered this turn must not be reported
// sick, or the client greys an ability the server would have let the
// player activate.
func TestSummoningSickIsCreaturesOnly(t *testing.T) {
	tests := []struct {
		name     string
		typeLine string
		mutate   func(*game.Card)
		want     bool
	}{
		{
			// #365. A Treasure minted this turn — by Smothering
			// Tithe, Pitiless Plunderer or a Dockside — has to be
			// crackable the moment it exists. That is the entire
			// point of the card.
			name:     "Treasure token made this turn",
			typeLine: "Token Artifact — Treasure",
			mutate: func(c *game.Card) {
				c.ManaAbilities = []game.ManaAbilityShape{{
					TapCost:       true,
					SacrificeCost: true,
					Produced:      "{W|U|B|R|G}",
					Label:         "{T}, Sacrifice: Add one mana of any color",
				}}
			},
			want: false,
		},
		{
			// #368. Fabled Passage and every other fetchland: land
			// drop, then {T}, Sacrifice, on the same turn.
			name:     "fetchland played this turn",
			typeLine: "Land",
			mutate: func(c *game.Card) {
				c.ActivatedAbilities = []game.ActivatedAbilityShape{{
					Label: "{T}, Sacrifice this: Search for a basic land",
					Cost:  game.AbilityCost{Tap: true, SacrificeSelf: true},
				}}
			},
			want: false,
		},
		{
			// Sol Ring, Arcane Signet, Mind Stone — a mana rock is
			// cast to be tapped the same turn.
			name:     "mana rock cast this turn",
			typeLine: "Artifact",
			mutate: func(c *game.Card) {
				c.ManaAbilities = []game.ManaAbilityShape{{
					TapCost: true, Produced: "{C}{C}", Label: "{T}: Add {C}{C}",
				}}
			},
			want: false,
		},
		{
			// An uncrewed Vehicle is not a creature (CR 301.7), so
			// it is not sick either — it simply cannot attack for
			// want of being a creature at all.
			name:     "uncrewed Vehicle that entered this turn",
			typeLine: "Artifact — Vehicle",
			want:     false,
		},
		{
			name:     "basic land played this turn",
			typeLine: "Basic Land — Forest",
			want:     false,
		},
		{
			// The rule itself, unchanged: CR 302.6 still gates a
			// creature.
			name:     "creature that entered this turn",
			typeLine: "Creature — Human Wizard",
			want:     true,
		},
		{
			name:     "creature with haste that entered this turn",
			typeLine: "Creature — Goblin",
			mutate: func(c *game.Card) {
				// Lowercase: deck.printedKeywords canonicalises on
				// the import road, and HasKeyword compares verbatim.
				c.Keywords = []string{"haste"}
			},
			want: false,
		},
		{
			// A land animated into a creature this turn is a
			// creature that has not been controlled continuously
			// since the turn began — CR 302.6 catches it.
			name:     "manland animated the turn it entered",
			typeLine: "Creature Land — Elemental",
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := buildActiveGame(t)
			id := seatFreshPermanent(g, tc.name, tc.typeLine, tc.mutate)
			if got := sickFlagOf(t, g, id); got != tc.want {
				t.Errorf("summoning_sick for %q (%s) = %v, want %v",
					tc.name, tc.typeLine, got, tc.want)
			}
		})
	}
}

// TestSettledCreatureIsNotSummoningSick pins the other half: the
// flag is about this turn, not about being a creature.
func TestSettledCreatureIsNotSummoningSick(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Settled Bear",
		TypeLine:   "Creature — Bear",
		Owner:      owner,
		Controller: owner,
	})
	if sickFlagOf(t, g, id) {
		t.Error("a creature that did not enter this turn must not be summoning sick")
	}
}
