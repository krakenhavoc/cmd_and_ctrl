package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_spells_test.go — S28 sub-PR 4: the alternative costs that
// charge something other than mana.
//
// What a bug here would hide, worst first:
//
//  1. The cost not being charged at all. A Force of Will that pitches
//     nothing is a free counterspell, which is the worst thing this
//     sprint can ship.
//  2. The condition not gating. A Fierce Guardianship castable for
//     nothing with no commander out is the same bug wearing a hat.
//  3. The cost being charged at the wrong TIME. It is paid at
//     announce with the spell on the stack, so countering the spell
//     does not refund it.

const (
	ForceOfWillOracle    = "956381ba-6d37-4a8a-846c-bad79222dbee"
	SnuffOutOracle       = "324824cb-f938-401c-b9b5-d8908b431ef0"
	DazeOracle           = "70486bee-6ee7-41ea-b834-8caf4699302b"
	SolitudeOracle       = "dcb9c2a7-ae54-4ddc-a567-640bf4bf4366"
	PactOfNegationOracle = "f3e213a4-ba5a-468a-93b3-c0a34e1bd725"
	KambalOracle         = "4987c458-604a-4727-b360-170616e91e67"
)

// handCardFull seeds a fully-specified card into a hand. The oracle
// ID matters here in a way it does not for most fixtures: every cost
// clause in this file is looked up by it, so a card seeded without
// one silently offers no alternative cost at all.
func HandCardFull(p *game.Player, name, typeLine, manaCost, oracle string, colors []string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		OracleID: oracle, Colors: colors, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// freeSpellPermanent drops a permanent onto the battlefield.
func FreeSpellPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	return id
}

func ToMainForCost(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 20 && g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}
