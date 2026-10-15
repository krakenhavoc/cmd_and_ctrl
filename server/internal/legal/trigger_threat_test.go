package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_threat_test.go — #3882. #3828 read an opponent's activated
// ability off its catalog row; a triggered ability on the stack is read
// the same way, through the ref its stack item carries to the triggered
// row. A row that declares a sweep reaching something the seat controls,
// or an edict (AnswerRemove), threatens the seat; a trigger that only
// touches their own board, or the seat's own trigger, does not.

const (
	oracleFleshbagMarauder   = "4b1bf05e-753e-4350-a913-894cf3cecc0c"
	oracleButcherOfMalakir   = "a85197ab-dc94-4b72-9716-8dbdbbe90ff8"
	oraclePhyrexianScripture = "11173ad3-c007-478f-bce0-d756eac07ccb"
	oracleToxrill            = "c71b2325-bde6-4364-a93b-8477ffeb25d8"
	oraclePurphoros          = "4fdbbec2-e921-4b63-958d-f9ba1e417197"
	oracleMassacreWurm       = "93cf50cf-0ecc-4d3e-abea-778c1ebacec4"
)

// pushTriggered puts a triggered ability of `source` (a catalog card on
// the board) on the stack under `by`, naming the catalog triggered row
// whose key contains `label`, as the harvest does.
func (b *threatBoard) pushTriggered(t *testing.T, by *game.Player, source uuid.UUID, label string) {
	t.Helper()
	var card game.Card
	b.g.WithWriteLock(func() {
		for _, c := range b.g.Battlefield.Cards {
			if c.InstanceID == source {
				card = c
			}
		}
	})
	for _, row := range game.CatalogTriggers(game.CatalogKey(card)) {
		if !strings.Contains(row.Key, label) {
			continue
		}
		ref := game.TriggeredAbilityRef(row)
		if ref == nil {
			t.Fatalf("%s: trigger %q cannot be named by a ref", card.Name, row.Key)
		}
		id := uuid.New()
		b.g.WithWriteLock(func() {
			b.g.Stack.PushTop(game.Card{InstanceID: id, Name: "Stack item", Owner: by.ID, Controller: by.ID})
			if b.g.StackMeta == nil {
				b.g.StackMeta = make(map[uuid.UUID]*game.StackItem)
			}
			b.g.StackMeta[id] = &game.StackItem{
				ID: id, Kind: game.StackItemTriggered, Controller: by.ID, Owner: by.ID,
				SourceCardID: source, Label: row.Key, Params: game.EffectParams{Ability: ref},
			}
		})
		return
	}
	t.Fatalf("%s has no triggered row containing %q", card.Name, label)
}

func TestTriggeredWipesAndEdictsThreatenTheSeat(t *testing.T) {
	cases := []struct {
		name  string
		card  game.Card
		label string
	}{
		{"Fleshbag Marauder's edict", game.Card{Name: "Fleshbag Marauder", TypeLine: "Creature — Zombie Warrior", OracleID: oracleFleshbagMarauder, ManaCost: "{2}{B}", Power: 3, Toughness: 1}, "each player sacrifices"},
		{"Butcher of Malakir's dies edict", game.Card{Name: "Butcher of Malakir", TypeLine: "Creature — Vampire Warrior", OracleID: oracleButcherOfMalakir, ManaCost: "{5}{B}{B}", Power: 5, Toughness: 4}, "each opponent sacrifices"},
		{"Phyrexian Scriptures' chapter II", game.Card{Name: "Phyrexian Scriptures", TypeLine: "Enchantment — Saga", OracleID: oraclePhyrexianScripture, ManaCost: "{2}{B}"}, "II: destroy all"},
		{"Massacre Wurm's enters trigger, whose sweep is declared on the card", game.Card{Name: "Massacre Wurm", TypeLine: "Creature — Phyrexian Wurm", OracleID: oracleMassacreWurm, ManaCost: "{3}{B}{B}{B}", Power: 6, Toughness: 5}, "get -2/-2"},
		{"Toxrill's slime counters on creatures you don't control", game.Card{Name: "Toxrill, the Corrosive", TypeLine: "Legendary Creature — Phyrexian Slug Horror", OracleID: oracleToxrill, ManaCost: "{5}{B}{B}", Power: 7, Toughness: 7}, "slime counter"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			b.pushTriggered(t, b.opp, b.oppCard(tc.card), tc.label)
			b.checkGated(t, true)
		})
	}
}

func TestTriggerThatTouchesOnlyTheirBoardDoesNotThreaten(t *testing.T) {
	t.Run("a chapter that exiles graveyards", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		saga := b.oppCard(game.Card{Name: "Phyrexian Scriptures", TypeLine: "Enchantment — Saga", OracleID: oraclePhyrexianScripture, ManaCost: "{2}{B}"})
		b.pushTriggered(t, b.opp, saga, "III: exile")
		b.checkGated(t, false)
	})
	t.Run("Massacre Wurm's life-loss trigger", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		wurm := b.oppCard(game.Card{Name: "Massacre Wurm", TypeLine: "Creature — Phyrexian Wurm", OracleID: oracleMassacreWurm, ManaCost: "{3}{B}{B}{B}", Power: 6, Toughness: 5})
		b.pushTriggered(t, b.opp, wurm, "loses 2 life")
		b.checkGated(t, false)
	})
	t.Run("a damage-to-each-opponent trigger", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		god := b.oppCard(game.Card{Name: "Purphoros, God of the Forge", TypeLine: "Legendary Enchantment Creature — God", OracleID: oraclePurphoros, ManaCost: "{3}{R}", Power: 6, Toughness: 5})
		b.pushTriggered(t, b.opp, god, "")
		b.checkGated(t, false)
	})
	t.Run("the seat's own edict", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		marauder := battlefieldCard(b.g, b.seat, game.Card{Name: "Fleshbag Marauder", TypeLine: "Creature — Zombie Warrior", OracleID: oracleFleshbagMarauder, ManaCost: "{2}{B}", Power: 3, Toughness: 1})
		b.pushTriggered(t, b.seat, marauder, "each player sacrifices")
		b.checkGated(t, false)
	})
}
