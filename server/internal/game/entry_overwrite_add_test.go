package game

import "testing"

// #4834: an entry overwrite may add types instead of setting them
// (Excava, the Risen Past), and Problem refuses the mixtures.
func TestEntersAsAddFormRegistersAddModsAndRefusesMixtures(t *testing.T) {
	add := &EntersAs{AddTypes: []string{"Creature"}, AddSubtypes: []string{"Spirit"}, Keywords: []string{"flying"},
		BasePT: &EntersAsPT{Power: 1, Toughness: 1}, Label: "add"}
	if p := add.Problem(); p != "" {
		t.Fatalf("the add form: %s", p)
	}
	kinds := map[ModKind]bool{}
	for _, m := range add.mods() {
		kinds[m.Kind] = true
	}
	for _, k := range []ModKind{ModAddTypes, ModAddSubtypes, ModAddKeywords, ModSetBasePower, ModSetBaseToughness} {
		if !kinds[k] {
			t.Errorf("missing mod %s", k)
		}
	}
	if kinds[ModSetTypes] {
		t.Error("the add form set the types")
	}
	for name, bad := range map[string]*EntersAs{
		"sets and adds":            {Types: []string{"Artifact"}, AddTypes: []string{"Creature"}},
		"no type":                  {Keywords: []string{"flying"}},
		"set subtypes with an add": {AddTypes: []string{"Creature"}, Subtypes: []string{"Spirit"}},
	} {
		if bad.Problem() == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	cp := copyEntersAs(add)
	cp.BasePT.Power = 5
	cp.AddTypes[0] = "Land"
	if add.BasePT.Power != 1 || add.AddTypes[0] != "Creature" {
		t.Error("copyEntersAs shares storage")
	}
}
