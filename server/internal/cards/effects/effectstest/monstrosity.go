package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monstrosity_test.go — #1700, CR 701.37 (ADR 0071 amendment
// 2026-09-28). The keyword action's rules first, through real
// activations on real cards, then one test per card.

const (
	PolukranosOracle       = "ecbf5560-853a-43a9-bf70-ac8d403b0ce8"
	StormbreathOracle      = "7a770865-9383-45aa-883f-21ee649d1ea3"
	HundredHandedOneOracle = "954aaaa7-c3c1-4696-9ace-bcb5359b1709"
	EmberSwallowerOracle   = "0cdc86c4-8c68-4ba6-8165-511f32156a4c"
	ArborColossusOracle    = "bec585af-0c59-46a7-838f-d747c80a9e97"
	FleecemaneLionOracle   = "e3c8cdf7-a26a-45eb-9498-99b618cfba99"
	DomesticatedHydraOrcl  = "b6d65fa3-26a9-44b2-a93d-a295bd080729"
	HydraBroodmasterOracle = "9e4fb4b1-9b28-4a49-b4ac-d5549c08594d"
)

func PushMonster(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Monster", OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// activateMonstrosity announces the card's one activated ability
// (every card here prints exactly one) with X = x.
func ActivateMonstrosity(t *testing.T, g *game.Game, who, id uuid.UUID, x int) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, id, 0, game.ActivateAbilityParams{XValue: x}); err != nil {
		t.Fatalf("activate monstrosity: %v", err)
	}
}

func MonstrousNow(g *game.Game, id uuid.UUID) bool {
	var m bool
	g.ReadSnapshot(func() { m = g.IsMonstrous(id) })
	return m
}

// becameMonstrousEvents returns the Amount of every EventBecameMonstrous
// the game emitted for `id`.
func BecameMonstrousEvents(g *game.Game, id uuid.UUID) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventBecameMonstrous && ev.CardID == id {
				out = append(out, ev.Amount)
			}
		}
	})
	return out
}

func GiveHandCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}
