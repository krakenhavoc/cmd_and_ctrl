package effectstest

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top100_test.go — the play-rate batch catalogued in
// docs/decklists/top-100-commander-staples.md: the twenty
// conditional-dual lands and the eleven singles that needed no new
// machinery.
//
// The assertions aim at what each card is FOR, not at "it resolved".
// A checkland that always enters untapped would pass a "did it enter"
// test and be a strictly better card than the one printed, so every
// conditional-dual test pins BOTH sides of its condition.

const (
	DrownedCatacombOracle    = "819fc966-434e-470f-91e9-a38df974ad17"
	SunkenHollowOracle       = "cd2c90ac-2b04-461c-92f3-939871b6b6a3"
	MorphicPoolOracle        = "bd004c9d-771e-4e63-a97d-a2259c096af8"
	PongifyOracle            = "05849bd6-8f38-4031-be2b-e2aa03beb8cc"
	AbradeOracle             = "f9db72dc-9a5b-48a4-a86e-7464d9a2166a"
	FeedTheSwarmOracle       = "5825997b-10d7-4a36-972c-a80ddd90b8ed"
	AssassinsTrophyOracle    = "ac10d218-f9a6-4058-9cda-a15ca1b0b7b5"
	DeadlyRollickOracle      = "0456ec64-2c81-4763-a352-8ff64a4c3d6b"
	FierceGuardianshipOracle = "d09c9cba-fdd2-479b-ad5d-d05181c3e3f9"
	CommandersSphereOracle   = "0b67c4e2-f88b-4e01-85a1-9d5f5b8db13b"
	PathOfAncestryOracle     = "b473e293-59e3-4e04-acf2-622604aeb25f"
	GarruksUprisingOracle    = "3127ae9b-a7a7-43ec-89d7-688f8445b33d"
	FabledPassageOracle      = "0c85b8f7-0bd0-4680-9ec5-d4b110460a54"
	MyriadLandscapeOracle    = "2549bc57-9ffb-4053-9f10-f2a5f792b845"
)

// newDuelCatalogGame is newCatalogGame with two seats. The bond lands
// are the only cards in the catalog whose behaviour depends on the
// number of opponents, and "unless you have two or more opponents" is
// only observable when the table is small enough for it to be false.
func NewDuelCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// top100ActiveSeat returns the seat whose turn it is — the one
// playLandFromHand / castCatalogSpell act as.
func Top100ActiveSeat(g *game.Game) *game.Player {
	return g.Seats[g.Turn.ActiveSeat]
}

// top100Opponent returns some seat other than the active one.
func Top100Opponent(g *game.Game) *game.Player {
	me := Top100ActiveSeat(g)
	for _, p := range g.Seats {
		if p != nil && p.ID != me.ID {
			return p
		}
	}
	return nil
}

// top100AssertEnteredTapped pins the enters-tapped REPLACEMENT rather
// than merely a tapped permanent: an OnETB tap would leave the card
// tapped too, but it would emit a tap event because the permanent
// really was untapped on the battlefield for a beat.
func Top100AssertEnteredTapped(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	card, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s isn't on the battlefield", name)
	}
	if !card.Tapped {
		t.Errorf("%s entered untapped; its condition was not met, so it must enter tapped", name)
	}
	if n := TapEventsFor(g, id); n != 0 {
		t.Errorf("%s: %d tap events — it should have ENTERED tapped, not been tapped after", name, n)
	}
}

func Top100AssertEnteredUntapped(t *testing.T, g *game.Game, id uuid.UUID, name string) {
	t.Helper()
	card, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s isn't on the battlefield", name)
	}
	if card.Tapped {
		t.Errorf("%s entered tapped; its condition was met, so it must enter untapped", name)
	}
}

// --- Garruk's Uprising -------------------------------------------

// top100CastCreature casts a plain non-catalog creature from a
// player's hand through the real stack path, so the resolution emits
// EventETB and the trigger harvester sees it.
func Top100CastCreature(t *testing.T, g *game.Game, p *game.Player, name string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	// A creature spell needs sorcery speed, and the priority passes
	// that resolved the previous spell may have moved the step on.
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
