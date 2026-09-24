package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// sacrifice_only_source_test.go — the enumerator half of #1242 and
// #1283. #544's invariant, twice: offer the casts and activations the
// engine accepts, and none it refuses.
//
//   - A board of Eldrazi Spawn pays for a cast now, so a bot is offered
//     it — before #1242 canPayExcluding asked the auto-tapper, which
//     refused every sacrifice-only source, and no bot ever cracked a
//     Spawn to cast anything.
//   - A Spawn a cast NAMES to its additional cost is not also the
//     cast's mana (game.CastAutoTapExclusions), so a cast whose
//     affordability leaned on it is not offered.
//   - Cadaverous Bloom's "Exile a card from your hand" is one payment
//     out of the engine's own walk, on its own wire field.

const (
	oracleDeadlyDispute   = "457af74a-02b3-4659-846d-63e482667f34"
	oracleCadaverousBloom = "fbb0f73b-5e30-4632-99c1-e49582e41f8d"
)

func spawnCard() game.Card {
	return game.Card{
		Name:      "Eldrazi Spawn",
		TypeLine:  "Token Creature — Eldrazi Spawn",
		Toughness: 1,
		ManaAbilities: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{C}",
			Label:         "Sacrifice this creature: Add {C}",
		}},
	}
}

func TestABoardOfEldraziSpawnMakesACastEnumerable(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, spawnCard())
	}
	golem := handCard(active, game.Card{Name: "Three-Drop Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{3}", Power: 3, Toughness: 3})

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, golem, legal.KindCast); len(got) == 0 {
		t.Fatalf("three Eldrazi Spawn pay {3}, but no cast was offered: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// Deadly Dispute ({1}{B}, "As an additional cost, sacrifice an artifact
// or creature") off a Swamp and one Eldrazi Spawn: the Spawn is the
// only thing to sacrifice AND the only source of the {1}. The engine
// will not spend it twice, so the cast is not a move.
func TestACastIsNotOfferedWhenItsSacrificeIsAlsoItsMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, spawnCard())
	dispute := handCard(active, game.Card{Name: "Deadly Dispute", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: oracleDeadlyDispute})

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, dispute, legal.KindCast); len(got) != 0 {
		t.Errorf("offered %v — the Spawn cannot be both the sacrifice and the {1}", labels(got))
	}
	dispatchAll(t, g, active.ID, moves)
}

// …and with a second Spawn out the same cast IS a move: one Spawn is
// the sacrifice, the other the {1}.
func TestACastIsOfferedWhenAnotherSourceCoversItsSacrifice(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, spawnCard())
	battlefieldCard(g, active, spawnCard())
	dispute := handCard(active, game.Card{Name: "Deadly Dispute", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: oracleDeadlyDispute})

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, dispute, legal.KindCast); len(got) == 0 {
		t.Fatalf("Swamp + two Spawn cast Deadly Dispute, but no cast was offered: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// The activation twin: "{1}, Sacrifice a creature: You gain 1 life"
// with one Eldrazi Spawn out. The Spawn is the only creature to
// sacrifice AND the only source of the {1}, and ActivateCatalogAbility's
// auto-tap will not spend what the activation names
// (game.AbilityAutoTapExclusions), so the activation is not a move.
// With a Mountain beside it, it is.
func TestAnActivationIsNotOfferedWhenItsSacrificeIsAlsoItsMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, active, spawnCard())
	outlet := battlefieldCard(g, active, game.Card{
		Name:     "Paid Altar",
		TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{1}, Sacrifice a creature: You gain 1 life",
			Cost: game.AbilityCost{
				Mana: "{1}",
				SacrificeOther: &game.TargetSpec{
					Mode:  "permanent",
					Label: "a creature",
					Zones: []game.ZoneKind{game.ZoneBattlefield},
					CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
						return c.IsCreature()
					},
					Min: 1,
					Max: 1,
				},
			},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, outlet, legal.KindActivate); len(got) != 0 {
		t.Errorf("offered %v — the Spawn cannot be both the sacrifice and the {1}", labels(got))
	}
	dispatchAll(t, g, active.ID, moves)

	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	moves = legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, outlet, legal.KindActivate); len(got) == 0 {
		t.Fatalf("a Mountain pays the {1}, but no activation was offered: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// Cadaverous Bloom: one mana move per activation, paying with a card
// out of hand on `exile_ids` (never `discard_ids`), and none at all
// with an empty hand.
func TestCadaverousBloomIsEnumeratedWithAnExilePayment(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)
	bloom := battlefieldCard(g, active, game.Card{Name: "Cadaverous Bloom", TypeLine: "Enchantment", OracleID: oracleCadaverousBloom})

	if got := movesFrom(legal.EnumerateFor(g, active.ID), bloom, legal.KindMana); len(got) != 0 {
		t.Fatalf("an empty hand offered %v — the Bloom has nothing to exile", labels(got))
	}

	pitch := handCard(active, creature("Spare Beast", "{5}{G}", 5, 5))
	moves := legal.EnumerateFor(g, active.ID)
	got := movesFrom(moves, bloom, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("want one Bloom activation, got %v", labels(got))
	}
	var params struct {
		ExileIDs   []string `json:"exile_ids"`
		DiscardIDs []string `json:"discard_ids"`
	}
	if err := json.Unmarshal(got[0].Params, &params); err != nil {
		t.Fatalf("params: %v", err)
	}
	if len(params.ExileIDs) != 1 || params.ExileIDs[0] != pitch.String() || len(params.DiscardIDs) != 0 {
		t.Errorf("params = %s, want exile_ids [%s] and no discard_ids", string(got[0].Params), pitch)
	}
	dispatchAll(t, g, active.ID, moves)
}
