package deck

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// companion_s82_rpr_b02_test.go — the conditions of Lurrus, Kaheera, Keruga,
// Lutri and Obosh (S82 reprint batch 2, #5908), judged on printed data.

func rprCard(name, typeLine, cost string) cards.Card {
	c := basicLegal(name, typeLine)
	c.OracleID = uuid.New()
	c.ManaCost = cost
	return c
}

func offendingNames(oracle string, deck ...cards.Card) []string {
	return companionConditions[oracle].Offenders(deck)
}

func TestRprB02EveryConditionIsRegistered(t *testing.T) {
	for _, id := range []string{
		"3bc757c1-3adb-4321-8832-8e1cc9e687f7", "35428c77-9731-4d4d-8f75-e939a491cf86",
		"ca5121b2-e6d9-40bc-a2d2-05852c4efbe8", "158a6225-a246-4fd6-aa57-0df8067b4383",
		"6985c8a9-9299-461b-8bef-bd2b2c739c11",
	} {
		if c, ok := companionConditions[id]; !ok || c.Offenders == nil || c.Text == "" {
			t.Errorf("%s is not registered with an Offenders function and a text", id)
		}
	}
}

func TestRprB02ManaValueOfPrintedCosts(t *testing.T) {
	for _, tc := range []struct {
		cost string
		want int
	}{
		{"", 0}, {"{X}{R}", 1}, {"{2}{G/W}{G/W}", 4}, {"{2/W}{2/W}", 4}, {"{R/P}", 1}, {"{1}{W}{U}{B}{R}{G}", 6}, {"{0}", 0},
	} {
		if got := companionManaValue(rprCard("c", "Instant", tc.cost)); got != tc.want {
			t.Errorf("mana value of %q = %d, want %d", tc.cost, got, tc.want)
		}
	}
	split := rprCard("Fire // Ice", "Instant // Instant", "{1}{R} // {1}{U}")
	split.Layout = "split"
	split.CardFaces = []cards.CardFace{{ManaCost: "{1}{R}"}, {ManaCost: "{1}{U}"}}
	if got := companionManaValue(split); got != 4 {
		t.Errorf("a split card's mana value = %d, want 4", got)
	}
	dfc := rprCard("Front // Back", "Creature — Human // Creature — Wolf", "")
	dfc.Layout = "transform"
	dfc.CardFaces = []cards.CardFace{{ManaCost: "{1}{G}"}, {ManaCost: ""}}
	if got := companionManaValue(dfc); got != 2 {
		t.Errorf("a double-faced card's mana value = %d, want its front's 2", got)
	}
}

func TestRprB02LurrusFlagsPermanentCardsOverTwo(t *testing.T) {
	const id = "3bc757c1-3adb-4321-8832-8e1cc9e687f7"
	got := offendingNames(id,
		rprCard("Cheap Bear", "Creature — Bear", "{1}{G}"),
		rprCard("Big Bear", "Creature — Bear", "{3}{G}"),
		rprCard("Big Spell", "Sorcery", "{5}{R}"),
		rprCard("Big Gear", "Artifact", "{4}"),
		rprCard("Forest", "Basic Land — Forest", ""),
	)
	if len(got) != 2 || got[0] != "Big Bear" || got[1] != "Big Gear" {
		t.Errorf("offenders %v, want [Big Bear Big Gear]", got)
	}
}

func TestRprB02KaheeraFlagsCreaturesOutsideItsTypes(t *testing.T) {
	const id = "35428c77-9731-4d4d-8f75-e939a491cf86"
	changeling := rprCard("Chameleon", "Creature — Shapeshifter", "{2}")
	changeling.Keywords = []string{"Changeling"}
	got := offendingNames(id,
		rprCard("Lion", "Creature — Cat", "{2}"),
		rprCard("Dino Knight", "Artifact Creature — Dinosaur Knight", "{2}"),
		rprCard("Bear", "Creature — Bear", "{2}"),
		rprCard("Goblin Spell", "Instant", "{R}"),
		rprCard("Cat Instant", "Kindred Instant — Cat", "{R}"),
		changeling,
	)
	if len(got) != 1 || got[0] != "Bear" {
		t.Errorf("offenders %v, want [Bear]", got)
	}
}

func TestRprB02KerugaAndOboshFlagByManaValue(t *testing.T) {
	deck := []cards.Card{
		rprCard("Two", "Creature — Bear", "{1}{G}"),
		rprCard("Three", "Creature — Bear", "{2}{G}"),
		rprCard("Four", "Sorcery", "{3}{R}"),
		rprCard("Zero", "Artifact", ""),
		rprCard("Forest", "Basic Land — Forest", ""),
	}
	if got := companionConditions["ca5121b2-e6d9-40bc-a2d2-05852c4efbe8"].Offenders(deck); len(got) != 2 || got[0] != "Two" || got[1] != "Zero" {
		t.Errorf("Keruga offenders %v, want [Two Zero]", got)
	}
	if got := companionConditions["6985c8a9-9299-461b-8bef-bd2b2c739c11"].Offenders(deck); len(got) != 3 || got[0] != "Two" || got[1] != "Four" || got[2] != "Zero" {
		t.Errorf("Obosh offenders %v, want [Two Four Zero]", got)
	}
}

func TestRprB02LutriFlagsRepeatedNonlandNamesOnly(t *testing.T) {
	const id = "158a6225-a246-4fd6-aa57-0df8067b4383"
	got := offendingNames(id,
		rprCard("Bolt", "Instant", "{R}"), rprCard("Bolt", "Instant", "{R}"), rprCard("Bolt", "Instant", "{R}"),
		rprCard("Once", "Instant", "{R}"),
		rprCard("Forest", "Basic Land — Forest", ""), rprCard("Forest", "Basic Land — Forest", ""),
	)
	if len(got) != 1 || got[0] != "Bolt" {
		t.Errorf("offenders %v, want [Bolt] once", got)
	}
}
