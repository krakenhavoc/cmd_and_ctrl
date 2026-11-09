package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dawnBlessedPennantOracle = "311eef28-21e9-44fa-96df-d281f0a9ba34"

func TestDawnBlessedPennantIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(dawnBlessedPennantOracle)
	if !ok || spec.Name != "Dawn-Blessed Pennant" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// dawnBlessedPennantEnter casts the Pennant and names `tribe` from the
// eight printed types.
func dawnBlessedPennantEnter(t *testing.T, g *game.Game, me *game.Player, tribe string) uuid.UUID {
	t.Helper()
	pennant := rfdCast(t, g, "Dawn-Blessed Pennant", "Artifact", dawnBlessedPennantOracle, 0, 0)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no type choice as the Pennant entered")
	}
	want := []string{"Elemental", "Elf", "Faerie", "Giant", "Goblin", "Kithkin", "Merfolk", "Treefolk"}
	if got := optionWords(pick); !slices.Equal(got, want) {
		t.Fatalf("offered %v, want %v", optionWords(pick), want)
	}
	answerOptionPick(t, g, me.ID, slices.Index(want, tribe))
	var named string
	g.WithWriteLock(func() { named = g.NamedTribeOf(pennant) })
	if named != tribe {
		t.Fatalf("chosen type = %q, want %q", named, tribe)
	}
	return pennant
}

// A permanent of the chosen type entering under your control gains 1
// life; a changeling counts; another type does not.
func TestDawnBlessedPennantGainsLifeForTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dawnBlessedPennantEnter(t, g, me, "Elf")

	life := me.Life
	rfdCast(t, g, "Elvish Mystic", "Creature — Elf Druid", "", 1, 1)
	if me.Life != life+1 {
		t.Fatalf("an Elf entering: life %d -> %d, want +1", life, me.Life)
	}
	rfdCast(t, g, "Goblin Guide", "Creature — Goblin Scout", "", 2, 2)
	if me.Life != life+1 {
		t.Fatalf("a Goblin entering changed life to %d", me.Life)
	}
	waryFarmerTokenFor(g, me.ID)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("a changeling entering: life %d, want %d", me.Life, life+2)
	}
}

// The return is limited to cards of the chosen type, judged against the
// Pennant as it last existed after the sacrifice; a changeling card is
// every type.
func TestDawnBlessedPennantReturnsACardOfTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pennant := dawnBlessedPennantEnter(t, g, me, "Kithkin")
	goblin := b17GraveyardCard(me, "Boggart Brute", "Creature — Goblin Warrior", "{2}{R}")
	kithkin := b17GraveyardCard(me, "Kinsbaile Cavalier", "Creature — Kithkin Knight", "{3}{W}")

	b06AddMana(me, "C", "C")
	err := g.ActivateCatalogAbility(me.ID, pennant, 0, game.ActivateAbilityParams{Targets: b16TargetCard(goblin), Strict: true})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting a Goblin with Kithkin chosen: err = %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(pennant) {
		t.Fatal("a refused activation sacrificed the Pennant")
	}
	b16Activate(t, g, me.ID, pennant, 0, game.ActivateAbilityParams{Targets: b16TargetCard(kithkin), Strict: true})
	if g.Battlefield.Contains(pennant) {
		t.Error("the Pennant is sacrificed as the cost")
	}
	if !me.Hand.Contains(kithkin) {
		t.Error("the Kithkin card did not return to hand")
	}
}

func TestDawnBlessedPennantReturnsAChangelingCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pennant := dawnBlessedPennantEnter(t, g, me, "Merfolk")
	shifter := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: shifter, Name: "Changeling Outcast", TypeLine: "Creature — Shapeshifter", ManaCost: "{B}",
		Keywords: []string{game.KeywordChangeling}, Owner: me.ID, Controller: me.ID,
	})
	b06AddMana(me, "C", "C")
	b16Activate(t, g, me.ID, pennant, 0, game.ActivateAbilityParams{Targets: b16TargetCard(shifter), Strict: true})
	if !me.Hand.Contains(shifter) {
		t.Error("a changeling card is a Merfolk card in the graveyard")
	}
}
