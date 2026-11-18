package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	afrB18DragonsFire   = "7237d607-2613-4786-96d6-55792c2e02b9"
	afrB18Disciple      = "3e5d3574-cc92-434a-a210-fcda906391c8"
	afrB18BlueDragon    = "1d00e54a-5897-43dc-9c38-a0e2ce6a7433"
	afrB18GreenDragon   = "c2a4c6c7-e5b4-4cc8-91d8-5de8cc95ed00"
	afrB18Iymrith       = "d5efba91-8400-41a3-91c6-cb5e81f07976"
	afrB18Inferno       = "cd54438d-7f9d-4d80-b15b-0f0bf4becc15"
	afrB18Regent        = "2df30a75-4957-4eab-8b0b-4c973f00642d"
	afrB18Tarrasque     = "bfd66214-e576-4a7e-8559-4e4e51d56b98"
	afrB18Storvald      = "5d7802a7-df6e-47fb-9065-c7461ea24a19"
	afrB18Klauth        = "8cefa5d6-5151-4e75-8b72-f210a4c751ea"
	afrB18KlauthsWill   = "639926c1-c4b4-4a49-b125-0e1b584c5bb8"
	afrB18DragonType    = "Creature — Dragon"
	afrB18LegendDragon  = "Legendary Creature — Dragon"
	afrB18BeholdOfferAt = 0
)

func TestAfrB18CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		afrB18DragonsFire: "Dragon's Fire", afrB18Disciple: "Dragon's Disciple",
		afrB18BlueDragon: "Blue Dragon", afrB18GreenDragon: "Green Dragon",
		afrB18Iymrith: "Iymrith, Desert Doom", afrB18Inferno: "Inferno of the Star Mounts",
		afrB18Regent: "Westgate Regent", afrB18Tarrasque: "The Tarrasque",
		afrB18Storvald: "Storvald, Frost Giant Jarl", afrB18Klauth: "Klauth, Unrivaled Ancient",
		afrB18KlauthsWill: "Klauth's Will",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func afrB18CastFire(t *testing.T, g *game.Game, params game.CastSpellParams) error {
	t.Helper()
	_, err := castWithTapParams(t, g, "Dragon's Fire", "Instant", "{1}{R}", afrB18DragonsFire, params)
	return err
}

// Declined, Dragon's Fire deals 3.
func TestAfrB18DragonsFireDealsThreeWithoutBehold(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	_ = me
	if err := afrB18CastFire(t, g, game.CastSpellParams{Targets: slotted(foe)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 3 {
		t.Errorf("foe took %d, want 3", got)
	}
}

// Choosing a Dragon you control: damage equals that Dragon's power as it
// resolves, a pump in response included.
func TestAfrB18DragonsFireUsesTheChosenDragonsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	dragon := apaPush(g, me.ID, me.ID, game.Card{Name: "Big Dragon", TypeLine: afrB18DragonType, Power: 6, Toughness: 6})
	if err := afrB18CastFire(t, g, game.CastSpellParams{
		OptionalCosts: []int{afrB18BeholdOfferAt},
		RevealIDs:     []uuid.UUID{dragon},
		Targets:       slotted(foe),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 6 {
		t.Errorf("foe took %d, want 6 (the Dragon's power)", got)
	}
}

// Revealing a Dragon card from hand: the card's printed power.
func TestAfrB18DragonsFireUsesTheRevealedCardsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	card := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: card, Name: "Hand Dragon", TypeLine: afrB18DragonType, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	if err := afrB18CastFire(t, g, game.CastSpellParams{
		OptionalCosts: []int{afrB18BeholdOfferAt},
		RevealIDs:     []uuid.UUID{card},
		Targets:       slotted(foe),
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Hand.Contains(card) {
		t.Error("revealing moved the card out of the hand")
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 5 {
		t.Errorf("foe took %d, want 5", got)
	}
}

// A non-Dragon is not a legal pick.
func TestAfrB18DragonsFireRefusesANonDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Foe", TypeLine: "Creature — Test", Power: 1, Toughness: 9})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := afrB18CastFire(t, g, game.CastSpellParams{
		Strict:        true,
		OptionalCosts: []int{afrB18BeholdOfferAt},
		RevealIDs:     []uuid.UUID{bear},
		Targets:       slotted(foe),
	}); err == nil {
		t.Fatal("a Bear was accepted as the Dragon to choose")
	}
}
