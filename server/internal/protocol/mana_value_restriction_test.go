package protocol

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4694: mana spendable only on a spell with mana value 4 or greater
// floats in the pool (Ashling, Rimebound adds it from a trigger), so the
// pool view carries its restrictions, and the bot's matcher reads the
// mana value of the face being cast with the X announced.

func TestManaPoolRunsKeepRestrictedManaApart(t *testing.T) {
	mv4 := []string{game.ManaRestrictManaValueAtLeast(4)}
	pool := game.ManaPool{
		{Color: "R"},
		{Color: "R", Restrictions: mv4},
		{Color: "R", Restrictions: mv4},
		{Color: "R"},
		{Color: "G"},
	}
	got := manaPoolRunsOf(pool)
	want := []ManaPoolRun{
		{Color: "R", Count: 1},
		{Color: "R", Count: 2, Restrictions: mv4},
		{Color: "R", Count: 1},
		{Color: "G", Count: 1},
	}
	if !slices.EqualFunc(got, want, func(a, b ManaPoolRun) bool {
		return a.Color == b.Color && a.Count == b.Count && slices.Equal(a.Restrictions, b.Restrictions)
	}) {
		t.Errorf("runs = %+v, want %+v", got, want)
	}
	if ManaPoolTotal(got) != 5 {
		t.Errorf("total = %d, want 5", ManaPoolTotal(got))
	}
	// The run owns its list: the engine token's slice is not shared.
	got[1].Restrictions[0] = "mutated"
	if pool[1].Restrictions[0] == "mutated" {
		t.Error("the run aliases the token's restriction list")
	}
}

func TestManaRestrictionsAdmitCastAsReadsTheFaceAndX(t *testing.T) {
	mv4 := []string{game.ManaRestrictManaValueAtLeast(4)}
	mdfc := CardView{
		TypeLine: "Sorcery // Land",
		Layout:   game.LayoutModalDFC,
		ManaCost: "{4}{R}",
		Faces:    []CardFaceView{{Name: "Spell", ManaCost: "{4}{R}"}, {Name: "Land"}},
	}
	split := CardView{
		TypeLine: "Instant // Sorcery",
		Layout:   "split",
		ManaCost: "{1}{R} // {3}{U}",
		Faces:    []CardFaceView{{Name: "Small", ManaCost: "{1}{R}"}, {Name: "Big", ManaCost: "{3}{U}"}},
	}
	transform := CardView{
		TypeLine: "Creature — Human // Creature — Werewolf",
		Layout:   game.LayoutTransform,
		ManaCost: "{3}{G}",
		Faces:    []CardFaceView{{Name: "Front", ManaCost: "{3}{G}"}, {Name: "Back"}},
	}
	fireball := CardView{TypeLine: "Sorcery", ManaCost: "{X}{R}"}
	cases := []struct {
		name  string
		card  CardView
		face  int
		x     int
		admit bool
	}{
		{"a modal DFC's spell face", mdfc, 0, 0, true},
		{"a modal DFC's other face, mana value 0", mdfc, 1, 0, false},
		{"a split card's small half", split, 0, 0, false},
		{"a split card's big half", split, 1, 0, true},
		{"a transform card's back face reads the front's cost (CR 712.8c)", transform, 1, 0, true},
		{"an X spell at X = 2", fireball, 0, 2, false},
		{"an X spell at X = 3", fireball, 0, 3, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ManaRestrictionsAdmitCastAs(mv4, &c.card, "", c.face, c.x); got != c.admit {
				t.Errorf("admits = %v, want %v", got, c.admit)
			}
		})
	}
	if ManaRestrictionsAdmitCast(mv4, &fireball, "") {
		t.Error("ManaRestrictionsAdmitCast counts {X} at 0: {X}{R} is mana value 1")
	}
}
