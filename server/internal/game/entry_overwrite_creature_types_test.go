package game

import "testing"

// An entry overwrite's CreatureTypes (assimilate, #4973 batch 11) ride
// beside AddTypes only, and every one must be a creature type.
func TestEntersAsCreatureTypesProblem(t *testing.T) {
	for _, tc := range []struct {
		name string
		e    EntersAs
		ok   bool
	}{
		{"add Artifact, become a Borg", EntersAs{AddTypes: []string{"Artifact"}, CreatureTypes: []string{"Borg"}}, true},
		{"set types and creature types", EntersAs{Types: []string{"Artifact", "Creature"}, CreatureTypes: []string{"Borg"}}, false},
		{"a creature type that is not one", EntersAs{AddTypes: []string{"Artifact"}, CreatureTypes: []string{"Food"}}, false},
	} {
		if got := tc.e.Problem() == ""; got != tc.ok {
			t.Errorf("%s: Problem() = %q, want ok %v", tc.name, tc.e.Problem(), tc.ok)
		}
	}
	e := &EntersAs{AddTypes: []string{"Artifact"}, CreatureTypes: []string{"Borg"}}
	c := copyEntersAs(e)
	c.CreatureTypes[0] = "Elf"
	if e.CreatureTypes[0] != "Borg" {
		t.Error("copyEntersAs shares the CreatureTypes backing array")
	}
	mods := e.mods()
	if len(mods) != 2 || mods[0].Kind != ModAddTypes || mods[1].Kind != ModSetCreatureTypes {
		t.Errorf("mods %+v, want add types then set creature types", mods)
	}
}
