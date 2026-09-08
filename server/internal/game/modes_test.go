package game

import (
	"testing"

	"github.com/google/uuid"
)

// modes_test.go — S20 sub-PR 4: modal spells. Mode validation at
// announce, the effective target clause derived from the chosen
// option, and the resolution re-check keyed on it.

func withCatalogModeSpec(t *testing.T, fn func(oracleID string) *ModeSpec) {
	t.Helper()
	prev := CatalogModeSpec
	CatalogModeSpec = fn
	t.Cleanup(func() { CatalogModeSpec = prev })
}

// charmSpec is a Rakdos Charm-shaped "choose one": target player /
// target artifact / untargeted.
func charmSpec() *ModeSpec {
	return &ModeSpec{
		Prompt: "Choose one",
		Options: []ModeOption{
			{Label: "target player", Targets: &TargetSpec{Mode: "player", Players: true, Min: 1, Max: 1}},
			{Label: "target artifact", Targets: &TargetSpec{
				Mode: "permanent", Zones: []ZoneKind{ZoneBattlefield},
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsArtifact() },
				Min:    1, Max: 1,
			}},
			{Label: "untargeted"},
		},
		Min: 1, Max: 1,
	}
}

func TestValidateModes(t *testing.T) {
	one := charmSpec()
	two := &ModeSpec{Options: make([]ModeOption, 4), Min: 2, Max: 2}
	cases := []struct {
		name  string
		spec  *ModeSpec
		modes []int
		want  error
	}{
		{"non-modal, none", nil, nil, nil},
		{"non-modal, some (free-form)", nil, []int{0}, nil},
		{"choose one, ok", one, []int{1}, nil},
		{"choose one, none", one, nil, ErrInvalidParam},
		{"choose one, two", one, []int{0, 1}, ErrInvalidParam},
		{"out of range", one, []int{3}, ErrInvalidParam},
		{"negative", one, []int{-1}, ErrInvalidParam},
		{"choose two, ok", two, []int{0, 3}, nil},
		{"choose two, duplicate", two, []int{1, 1}, ErrInvalidParam},
		{"choose two, one", two, []int{2}, ErrInvalidParam},
	}
	for _, tc := range cases {
		if got := validateModes(tc.spec, tc.modes); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCastTargetSpecFollowsChosenMode(t *testing.T) {
	const oracle = "test-charm"
	withCatalogModeSpec(t, func(id string) *ModeSpec {
		if id == oracle {
			return charmSpec()
		}
		return nil
	})
	withCatalogTargetSpec(t, func(string) *TargetSpec { return nil })

	if spec, err := castTargetSpec(oracle, []int{0}); err != nil || spec == nil || spec.Mode != "player" {
		t.Errorf("mode 0: spec %+v err %v, want player spec", spec, err)
	}
	if spec, err := castTargetSpec(oracle, []int{2}); err != nil || spec != nil {
		t.Errorf("untargeted mode: spec %+v err %v, want nil, nil", spec, err)
	}
	if _, err := castTargetSpec(oracle, []int{0, 1}); err != ErrInvalidParam {
		t.Errorf("two targeted modes: err %v, want ErrInvalidParam", err)
	}
	if spec, err := castTargetSpec("not-modal", nil); err != nil || spec != nil {
		t.Errorf("non-modal, no targets: %+v %v", spec, err)
	}
}

func modalCharmInHand(t *testing.T, g *Game, me *Player, oracle string) uuid.UUID {
	t.Helper()
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	c := NewCard("Charm", me.ID)
	c.TypeLine = "Instant"
	c.OracleID = oracle
	me.Hand.PushTop(c)
	return c.InstanceID
}

func TestCastSpellValidatesModeAndItsTarget(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-charm"
	withCatalogModeSpec(t, func(id string) *ModeSpec {
		if id == oracle {
			return charmSpec()
		}
		return nil
	})
	withCatalogTargetSpec(t, func(string) *TargetSpec { return nil })
	rock := pushArtifact(g, opp, "Sol Ring")
	bear := pushColoredCreature(g, opp, "Bear", "{1}{G}")
	charm := modalCharmInHand(t, g, me, oracle)

	// No mode at all.
	if err := g.CastSpell(me.ID, charm, CastSpellParams{}); err != ErrInvalidParam {
		t.Fatalf("modal card without modes: %v, want ErrInvalidParam", err)
	}
	// Artifact mode with a creature target.
	err := g.CastSpell(me.ID, charm, CastSpellParams{Modes: []int{1}, Targets: []TargetRef{{Kind: TargetCard, ID: bear}}})
	if err != ErrIllegalTarget {
		t.Fatalf("artifact mode + creature: %v, want ErrIllegalTarget", err)
	}
	// Player mode with a card target.
	err = g.CastSpell(me.ID, charm, CastSpellParams{Modes: []int{0}, Targets: []TargetRef{{Kind: TargetCard, ID: rock}}})
	if err != ErrIllegalTarget {
		t.Fatalf("player mode + card: %v, want ErrIllegalTarget", err)
	}
	// Untargeted mode must not carry a target.
	err = g.CastSpell(me.ID, charm, CastSpellParams{Modes: []int{2}, Targets: []TargetRef{{Kind: TargetCard, ID: rock}}})
	if err != ErrInvalidParam {
		t.Fatalf("untargeted mode + target: %v, want ErrInvalidParam", err)
	}
	// Targeted mode must carry one.
	err = g.CastSpell(me.ID, charm, CastSpellParams{Modes: []int{1}})
	if err != ErrInvalidParam {
		t.Fatalf("artifact mode without target: %v, want ErrInvalidParam", err)
	}
	if !me.Hand.Contains(charm) {
		t.Fatalf("rejected casts must leave the card in hand")
	}
	// The legal pairing.
	err = g.CastSpell(me.ID, charm, CastSpellParams{Modes: []int{1}, Targets: []TargetRef{{Kind: TargetCard, ID: rock}}})
	if err != nil {
		t.Fatalf("artifact mode + artifact: %v", err)
	}
	item := g.StackMeta[charm]
	if item == nil || len(item.Modes) != 1 || item.Modes[0] != 1 {
		t.Fatalf("stack item modes = %+v, want [1]", item)
	}
}

// The resolution re-check uses the chosen mode's predicate: an
// artifact-mode charm whose target stopped being an artifact in
// response fizzles.
func TestModalSpellRecheckUsesChosenModeSpec(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-charm"
	withCatalogModeSpec(t, func(id string) *ModeSpec {
		if id == oracle {
			return charmSpec()
		}
		return nil
	})
	withCatalogTargetSpec(t, func(string) *TargetSpec { return nil })
	rock := pushArtifact(g, opp, "Sol Ring")
	charm := modalCharmInHand(t, g, me, oracle)
	if err := g.CastSpell(me.ID, charm, CastSpellParams{Modes: []int{1}, Targets: []TargetRef{{Kind: TargetCard, ID: rock}}}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].TypeLine = "Creature — Golem"
			}
		}
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatal(err)
		}
	})
	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == EventFizzle && ev.CardID == charm {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("charm whose artifact target became a creature should be countered by game rules")
	}
}
