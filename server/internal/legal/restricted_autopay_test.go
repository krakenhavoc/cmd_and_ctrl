package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// restricted_autopay_test.go — #3874's enumerator half. Auto-pay now
// plans Cavern of Souls' and Delighted Halfling's coloured halves for a
// spell their restriction admits, so whether a cast is payable depends on
// the spell. The enumerator prices a cast under the same spend context the
// cast gate pays under, so it offers the cast exactly when the dispatched
// move goes through, and the bot and the client's offers agree with
// auto-pay. The engine half is game/autotap_restricted_test.go.

const (
	oracleCavernOfSouls     = "89ca686a-7c72-4d8f-9290-e89635624a83"
	oracleDelightedHalfling = "f9d3b046-0b95-4103-a630-4b3fb88bb60b"
)

// namedCavern puts a Cavern of Souls on the battlefield and names `tribe`.
func namedCavern(t *testing.T, g *game.Game, p *game.Player, tribe string) uuid.UUID {
	t.Helper()
	id := battlefieldCard(g, p, game.Card{Name: "Cavern of Souls", TypeLine: "Land", OracleID: oracleCavernOfSouls})
	var choice uuid.UUID
	g.WithWriteLock(func() { choice = g.QueueCreatureTypeChoiceForEffect(p.ID, id, "Cavern of Souls") })
	if err := g.ResolveCreatureTypeChoice(choice, p.ID, tribe); err != nil {
		t.Fatalf("name %s: %v", tribe, err)
	}
	return id
}

func TestRestrictedManaCastsAreOfferedExactlyWhenAutoPayCanPay(t *testing.T) {
	for _, tc := range []struct {
		name     string
		board    func(t *testing.T, g *game.Game, p *game.Player) uuid.UUID
		spell    game.Card
		castable bool
	}{
		{"Cavern named Elf, an Elf", func(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
			return namedCavern(t, g, p, "Elf")
		}, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"}, true},
		{"Cavern named Elf, a Goblin", func(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
			return namedCavern(t, g, p, "Elf")
		}, game.Card{Name: "Goblin Guide", TypeLine: "Creature — Goblin Scout", ManaCost: "{R}"}, false},
		{"Halfling, a legendary spell", func(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
			return battlefieldCard(g, p, game.Card{Name: "Delighted Halfling", TypeLine: "Creature — Halfling Citizen", OracleID: oracleDelightedHalfling})
		}, game.Card{Name: "Yeva, Nature's Herald", TypeLine: "Legendary Creature — Elf Shaman", ManaCost: "{G}"}, true},
		{"Halfling, a nonlegendary spell", func(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
			return battlefieldCard(g, p, game.Card{Name: "Delighted Halfling", TypeLine: "Creature — Halfling Citizen", OracleID: oracleDelightedHalfling})
		}, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{G}"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			active := g.Seats[g.Turn.ActiveSeat]
			clearHand(active)
			advanceTo(t, g, game.StepPrecombatMain)
			source := tc.board(t, g, active)
			spell := handCard(active, tc.spell)

			casts := castMovesOf(legal.EnumerateFor(g, active.ID), spell)
			if got := len(casts) > 0; got != tc.castable {
				t.Fatalf("cast offered = %v, want %v", got, tc.castable)
			}
			for _, m := range casts {
				clone := g.Clone()
				if err := actions.Dispatch(clone, actions.Action{
					Type: actions.Type(m.Type), Player: m.Player, Caller: active.ID, Params: m.Params,
				}); err != nil {
					t.Fatalf("offered move %q rejected: %v", m.Label, err)
				}
				tapped := false
				for _, c := range clone.Battlefield.Cards {
					if c.InstanceID == source {
						tapped = c.Tapped
					}
				}
				if !tapped {
					t.Errorf("%q: the restricted source was not tapped", m.Label)
				}
				if item := clone.StackMeta[spell]; item == nil || !item.SpellCantBeCounteredByMana() {
					t.Errorf("%q: the spell does not record the can't-be-countered rider", m.Label)
				}
			}
		})
	}
}
