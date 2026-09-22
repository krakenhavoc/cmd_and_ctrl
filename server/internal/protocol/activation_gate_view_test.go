package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_gate_view_test.go — the VIEW half of #1210 (ADR 0073's
// amendment of 2026-09-22). The client greys an ability row and names
// the clause from server data, so the stamp has to be the same gate's
// answer the engine and the enumerator read.
//
// Stubs the catalog hook rather than registering a real card: the
// question is about the projection, not about any card, and this
// package must be able to ask it without the catalog.

const viewActivationOracle = "test-view-activation-restriction"

func stubViewActivationRestrictions(t *testing.T, oracleID string, rules []game.ActivationRestriction) {
	t.Helper()
	prev := game.CatalogActivationRestrictions
	game.CatalogActivationRestrictions = func(id string) []game.ActivationRestriction {
		if id == oracleID {
			return rules
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogActivationRestrictions = prev })
}

// TestCantActivateIsStampedOnBothAbilityKinds is the projection's
// whole job: the refusal the engine would give, on the row, before
// the click — and on the MANA row too, because the restriction and
// not the call site decides whether CR 605.1a exempts it.
func TestCantActivateIsStampedOnBothAbilityKinds(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const clause = "Test Totem — activated abilities of creatures can't be activated."

	bear := uuid.New()
	totem := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: bear,
			Name:       "Test Bear",
			TypeLine:   "Creature — Bear",
			Owner:      me.ID,
			Controller: me.ID,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "{T}: Draw a card.", Cost: game.AbilityCost{Tap: true}},
			},
			ManaAbilities: []game.ManaAbilityShape{
				{TapCost: true, Produced: "{G}", Label: "Add {G}"},
			},
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: totem,
			Name:       "Test Totem",
			TypeLine:   "Artifact",
			OracleID:   viewActivationOracle,
			Owner:      me.ID,
			Controller: me.ID,
		})
	})

	// Nothing restricting: both rows are clean, and the field stays
	// off the wire entirely.
	c := conditionCardView(t, g, bear)
	if c.ActivatedAbilities[0].CantActivate != "" || c.ManaAbilities[0].CantActivate != "" {
		t.Fatalf("with nothing restricting, cant_activate = %q / %q, want empty",
			c.ActivatedAbilities[0].CantActivate, c.ManaAbilities[0].CantActivate)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "cant_activate") {
		t.Errorf("an unrestricted card must leave cant_activate off the wire: %s", raw)
	}

	stubViewActivationRestrictions(t, viewActivationOracle, []game.ActivationRestriction{{
		Label:   clause,
		Forbids: func(q game.ActivationQuery) bool { return q.Card.IsCreature() },
	}})

	c = conditionCardView(t, g, bear)
	if c.ActivatedAbilities[0].CantActivate != clause {
		t.Errorf("activated row: cant_activate = %q, want the printed clause", c.ActivatedAbilities[0].CantActivate)
	}
	if c.ManaAbilities[0].CantActivate != clause {
		t.Errorf("mana row: cant_activate = %q, want the printed clause — this clause prints no mana exemption",
			c.ManaAbilities[0].CantActivate)
	}
	// The restricting permanent itself is not a creature, so its own
	// (absent) rows say nothing — and neither does anything else on
	// the board. The stamp is per ABILITY, not per card.
	if c.ActivatedAbilities[0].ConditionUnmet || c.ActivatedAbilities[0].Exhausted {
		t.Error("cant_activate must not be confused with its two siblings — the three recover differently")
	}
}

// TestCantActivateExemptsManaAbilitiesWhenTheClauseDoes is the other
// half of the same decision, on the wire: a Pithing-Needle-shaped
// clause greys the non-mana row and leaves the mana row alone.
func TestCantActivateExemptsManaAbilitiesWhenTheClauseDoes(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const clause = "Test Needle — … can't be activated unless they're mana abilities."

	bear := uuid.New()
	needle := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: bear,
			Name:       "Test Bear",
			TypeLine:   "Creature — Bear",
			Owner:      me.ID,
			Controller: me.ID,
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "{T}: Draw a card.", Cost: game.AbilityCost{Tap: true}},
			},
			ManaAbilities: []game.ManaAbilityShape{
				{TapCost: true, Produced: "{G}", Label: "Add {G}"},
			},
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: needle,
			Name:       "Test Needle",
			TypeLine:   "Artifact",
			OracleID:   viewActivationOracle,
			Owner:      me.ID,
			Controller: me.ID,
		})
	})
	stubViewActivationRestrictions(t, viewActivationOracle, []game.ActivationRestriction{{
		Label: clause,
		Forbids: func(q game.ActivationQuery) bool {
			return !q.Ability.Mana && q.Card.IsCreature()
		},
	}})

	c := conditionCardView(t, g, bear)
	if c.ActivatedAbilities[0].CantActivate != clause {
		t.Errorf("activated row: cant_activate = %q, want the printed clause", c.ActivatedAbilities[0].CantActivate)
	}
	if c.ManaAbilities[0].CantActivate != "" {
		t.Errorf("mana row: cant_activate = %q, want empty — the clause exempts mana abilities",
			c.ManaAbilities[0].CantActivate)
	}
}
