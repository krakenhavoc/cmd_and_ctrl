package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Cavern of Souls
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_restriction_admits_test.go is the view half of #3976: a mana row
// whose restriction is computed (Cavern of Souls' chosen type) carries
// the list its mint would stamp, and ManaRestrictionsAdmitCast answers
// for a projected card what the pool solver answers for the real one.

const cavernOfSoulsOracle = "89ca686a-7c72-4d8f-9290-e89635624a83"

func TestCavernsRowCarriesItsChosenTypeRestriction(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Cavern of Souls", TypeLine: "Land", OracleID: cavernOfSoulsOracle,
		Owner: owner, Controller: owner, NamedTribe: "Elf",
	})
	var cv *CardView
	v := ViewOfGame(g)
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == id.String() {
			cv = &v.Battlefield.Cards[i]
		}
	}
	if cv == nil || len(cv.ManaAbilities) != 2 {
		t.Fatalf("Cavern's view = %+v, want two mana rows", cv)
	}
	if len(cv.ManaAbilities[0].Restrictions) != 0 {
		t.Errorf("the {C} row carries %v, want no restriction", cv.ManaAbilities[0].Restrictions)
	}
	want := []string{game.ManaRestrictCast, game.ManaRestrictType("Creature"), game.ManaRestrictSubtype("Elf")}
	if got := cv.ManaAbilities[1].Restrictions; !slices.Equal(got, want) {
		t.Fatalf("the coloured row carries %v, want %v", got, want)
	}
	elf := &CardView{TypeLine: "Creature — Elf Druid", Colors: []string{"G"}}
	goblin := &CardView{TypeLine: "Creature — Goblin", Colors: []string{"R"}}
	if !ManaRestrictionsAdmitCast(cv.ManaAbilities[1].Restrictions, elf, "") {
		t.Error("an Elf creature spell is refused Cavern's mana named for Elf")
	}
	if ManaRestrictionsAdmitCast(cv.ManaAbilities[1].Restrictions, goblin, "") {
		t.Error("a Goblin is admitted by Cavern's mana named for Elf")
	}
}

func TestManaRestrictionsAdmitCastAsksTheEnginesMatcher(t *testing.T) {
	legendary := []string{game.ManaRestrictCast, game.ManaRestrictSupertype("Legendary")}
	elf := []string{game.ManaRestrictCast, game.ManaRestrictType("Creature"), game.ManaRestrictSubtype("Elf")}
	cases := []struct {
		name  string
		rs    []string
		card  CardView
		from  string
		admit bool
	}{
		{"no restriction admits anything", nil, CardView{TypeLine: "Sorcery"}, "", true},
		{"a legendary creature for Delighted Halfling", legendary, CardView{TypeLine: "Legendary Creature — Elf Druid"}, "command", true},
		{"a nonlegendary one is refused", legendary, CardView{TypeLine: "Creature — Elf Druid"}, "", false},
		{"a changeling is every creature type (CR 702.73a)", elf, CardView{TypeLine: "Creature — Shapeshifter", Abilities: []string{"changeling"}}, "", true},
		{"a noncreature Elf is refused", elf, CardView{TypeLine: "Kindred Sorcery — Elf"}, "", false},
		{"a symbol-level tag is not answered", []string{game.ManaRestrictNoGeneric}, CardView{TypeLine: "Creature — Elf"}, "", false},
		{"not from hand: a graveyard cast is admitted", []string{game.ManaRestrictNotFromHand}, CardView{TypeLine: "Sorcery"}, "graveyard", true},
		{"not from hand: a hand cast is refused", []string{game.ManaRestrictNotFromHand}, CardView{TypeLine: "Sorcery"}, "", false},
		{"an unknown zone is refused", legendary, CardView{TypeLine: "Legendary Creature — Elf"}, "sideboard", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ManaRestrictionsAdmitCast(c.rs, &c.card, c.from); got != c.admit {
				t.Errorf("admits = %v, want %v", got, c.admit)
			}
		})
	}
}
