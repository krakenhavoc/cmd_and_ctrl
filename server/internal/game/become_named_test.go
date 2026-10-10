package game

import (
	"reflect"
	"strings"
	"testing"
)

// become_named_test.go pins #2562's two mods: CR 205.1a's type set and
// CR 612.8's name set, and what registration and restore refuse.

func TestSetTypesFollowsTheTypeRules(t *testing.T) {
	cases := []struct {
		name                         string
		types, subtypes, supertypes  []string
		modTypes, modSubs, modSupers []string
		wantTypes, wantSubs          []string
		wantSupers                   []string
	}{
		{
			// A Dryad Arbor made an Equipment artifact: both its land and
			// its creature types go with their card types.
			name: "card types replaced", types: []string{"Land", "Creature"}, subtypes: []string{"Forest", "Dryad"},
			modTypes: []string{"Artifact"}, modSubs: []string{"Equipment"},
			wantTypes: []string{"Artifact"}, wantSubs: []string{"Equipment"},
		},
		{
			// Only the artifact types are replaced; the creature type
			// stays with the creature type that stays (CR 205.1a, "from
			// the appropriate set").
			name: "other sets kept", types: []string{"Artifact", "Creature"}, subtypes: []string{"Vehicle", "Golem"},
			modTypes: []string{"Artifact", "Creature"}, modSubs: []string{"Equipment"},
			wantTypes: []string{"Artifact", "Creature"}, wantSubs: []string{"Golem", "Equipment"},
		},
		{
			// CR 205.4b: a supertype is gained, and the others stay.
			name: "supertypes gained", types: []string{"Artifact"}, supertypes: []string{"Snow"},
			modTypes: []string{"Artifact"}, modSupers: []string{"Legendary"},
			wantTypes: []string{"Artifact"}, wantSubs: []string{}, wantSupers: []string{"Snow", "Legendary"},
		},
		{
			// CR 205.1a's exception: an instant keeps that type.
			name: "instant kept", types: []string{"Instant"}, subtypes: []string{"Arcane"},
			modTypes:  []string{"Artifact"},
			wantTypes: []string{"Instant", "Artifact"}, wantSubs: []string{"Arcane"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Characteristic{Types: tc.types, Subtypes: tc.subtypes, Supertypes: tc.supertypes}
			c.setTypes(tc.modTypes, tc.modSubs, tc.modSupers)
			if !reflect.DeepEqual(c.Types, tc.wantTypes) {
				t.Errorf("types = %v, want %v", c.Types, tc.wantTypes)
			}
			if len(c.Subtypes) != len(tc.wantSubs) || (len(tc.wantSubs) > 0 && !reflect.DeepEqual(c.Subtypes, tc.wantSubs)) {
				t.Errorf("subtypes = %v, want %v", c.Subtypes, tc.wantSubs)
			}
			if tc.wantSupers != nil && !reflect.DeepEqual(c.Supertypes, tc.wantSupers) {
				t.Errorf("supertypes = %v, want %v", c.Supertypes, tc.wantSupers)
			}
		})
	}
}

// "Is every creature type" goes with the creature type (CR 205.3m).
func TestSetTypesTakesEveryCreatureTypeWithTheCreatureType(t *testing.T) {
	c := Characteristic{Types: []string{"Creature"}, AllCreatureTypes: true}
	c.setTypes([]string{"Artifact"}, nil, nil)
	if c.AllCreatureTypes {
		t.Error("a noncreature artifact is still every creature type")
	}
}

func TestBecomeNamedModProblems(t *testing.T) {
	cases := []struct {
		mod  Mod
		want string
	}{
		{SetNameMod(""), "names no name"},
		{SetTypesMod(nil, nil), "no card type"},
		{SetTypesMod([]string{"Artifact"}, []string{"Goblin"}), "not one of its card types"},
		{SetTypesMod([]string{"Artifact"}, nil, "Mythic"), "not a supertype"},
		{Mod{Kind: ModAddTypes, Types: []string{"Artifact"}, Supertypes: []string{"Legendary"}}, "only a setTypes mod reads"},
	}
	for _, tc := range cases {
		if got := becomeNamedModProblem(tc.mod); !strings.Contains(got, tc.want) {
			t.Errorf("%+v: problem %q, want one mentioning %q", tc.mod, got, tc.want)
		}
	}
	for _, ok := range []Mod{
		SetNameMod("Everflame, Heroes' Legacy"),
		SetTypesMod([]string{"Artifact"}, []string{"Equipment"}, "Legendary"),
	} {
		if got := becomeNamedModProblem(ok); got != "" {
			t.Errorf("%+v refused: %s", ok, got)
		}
	}
}
