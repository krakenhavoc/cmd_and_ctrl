package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const eclipsedRealmsOracle = "5715ed43-395c-4877-99a7-8e28e7bf9dce"

func TestEclipsedRealmsIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(eclipsedRealmsOracle)
	if !ok || spec.Name != "Eclipsed Realms" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// eclipsedRealmsPlay plays the land, names `tribe` from the printed
// eight, and taps it for one restricted white mana.
func eclipsedRealmsPlay(t *testing.T, g *game.Game, me *game.Player, tribe string) uuid.UUID {
	t.Helper()
	land := playLandFromHand(t, g, "Eclipsed Realms", eclipsedRealmsOracle)
	monSettle(t, g)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no type choice as the land entered")
	}
	want := []string{"Elemental", "Elf", "Faerie", "Giant", "Goblin", "Kithkin", "Merfolk", "Treefolk"}
	if got := optionWords(pick); !slices.Equal(got, want) {
		t.Fatalf("offered %v, want %v", got, want)
	}
	answerOptionPick(t, g, me.ID, slices.Index(want, tribe))
	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for coloured mana: %v", err)
	}
	ashlingPickColor(t, g, me.ID, "W")
	if got := me.ManaPool.Total(); got != 1 {
		t.Fatalf("pool holds %d mana, want 1", got)
	}
	return land
}

// The coloured mana casts a spell of the chosen type, of any card type,
// and nothing else.
func TestEclipsedRealmsManaCastsOnlySpellsOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eclipsedRealmsPlay(t, g, me, "Kithkin")

	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Goblin Instant", TypeLine: "Kindred Instant — Goblin", ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	var short *game.InsufficientManaError
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a Goblin spell with the Realms' mana: err = %v, want InsufficientManaError", err)
	}
	kith := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: kith, Name: "Kithkin Instant", TypeLine: "Kindred Instant — Kithkin", ManaCost: "{W}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, kith, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a Kithkin spell with the Realms' mana: %v", err)
	}
}

// The coloured mana activates an ability of a source of the chosen type,
// and not of another source.
func TestEclipsedRealmsManaActivatesOnlyAbilitiesOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eclipsedRealmsPlay(t, g, me, "Elf")

	goblin := b27Push(g, me.ID, "Gingerbrute", "Artifact Creature — Goblin", gingerbruteOracle, "{1}", 1, 1)
	var short *game.InsufficientManaError
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{Strict: true}); !errors.As(err, &short) {
		t.Fatalf("a Goblin's ability with the Realms' mana: err = %v, want InsufficientManaError", err)
	}
	elf := b27Push(g, me.ID, "Gingerbrute", "Artifact Creature — Elf", gingerbruteOracle, "{1}", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("an Elf's ability with the Realms' mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool holds %d mana after the activation, want 0", got)
	}
}
