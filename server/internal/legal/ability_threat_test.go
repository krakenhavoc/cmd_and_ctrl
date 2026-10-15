package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_threat_test.go — #3825. #3821 gated the stack answers on an
// opponent's untargeted instant or sorcery, which left a wipe that comes
// from an ability (a planeswalker's minus, Nevinyrral's Disk, Pestilence,
// Pernicious Deed) unable to stop you. An opponent's activated ability now
// threatens the seat when its catalog row declares a sweep or an edict
// (Purpose.Sweep, else AnswerRemove) whose class holds something the seat
// controls. An ability that only touches their own board does not.
// Pernicious Deed has no catalog row yet; Nevinyrral's Disk, Pestilence
// and two planeswalkers stand in for it.

const (
	oracleNevinyrralsDisk = "96230edf-568a-47dd-b877-9d92aa58fac8"
	oracleEternalWanderer = "20a1671d-e8a4-4cf1-87a7-f2f6319f4b9e"
	oracleLilianaVeil     = "0ba134d8-ee7d-48ec-8dc6-57942b8e9261"
)

// pushActivated puts an activated ability of `source` (a catalog card on
// the board) on the stack under `by`, naming the catalog row whose label
// contains `label`, as the activation path does.
func (b *threatBoard) pushActivated(t *testing.T, by *game.Player, source uuid.UUID, label string) {
	t.Helper()
	var card game.Card
	b.g.WithWriteLock(func() {
		for _, c := range b.g.Battlefield.Cards {
			if c.InstanceID == source {
				card = c
			}
		}
	})
	key := game.CatalogKey(card)
	row := -1
	var name string
	for i, ab := range game.CatalogActivatedAbilities(key) {
		if strings.Contains(ab.Label, label) {
			row, name = i, ab.Label
		}
	}
	if row < 0 {
		t.Fatalf("%s has no activated row containing %q", card.Name, label)
	}
	b.pushItem(by, source, name, &game.AbilityRef{
		Key: key, Slot: game.AbilitySlotActivated, Ref: game.OwnAbilityRef(row), Name: name,
	})
}

// pushItem puts an activated-ability item on the stack. A nil ref is a
// sandbox activation, which names no catalog row.
func (b *threatBoard) pushItem(by *game.Player, source uuid.UUID, label string, ref *game.AbilityRef) {
	id := uuid.New()
	b.g.WithWriteLock(func() {
		b.g.Stack.PushTop(game.Card{InstanceID: id, Name: "Stack item", Owner: by.ID, Controller: by.ID})
		if b.g.StackMeta == nil {
			b.g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		b.g.StackMeta[id] = &game.StackItem{
			ID: id, Kind: game.StackItemActivated, Controller: by.ID, Owner: by.ID,
			SourceCardID: source, Label: label, Params: game.EffectParams{Ability: ref},
		}
	})
}

func (b *threatBoard) oppCard(c game.Card) uuid.UUID { return battlefieldCard(b.g, b.opp, c) }

func TestAbilityWipesThreatenTheSeat(t *testing.T) {
	cases := []struct {
		name  string
		card  game.Card
		label string
	}{
		{"Nevinyrral's Disk", game.Card{Name: "Nevinyrral's Disk", TypeLine: "Artifact", OracleID: oracleNevinyrralsDisk, ManaCost: "{4}"}, "Destroy all"},
		{"Pestilence", game.Card{Name: "Pestilence", TypeLine: "Enchantment", OracleID: oraclePestilenceTG, ManaCost: "{2}{B}{B}"}, "1 damage to each creature"},
		{"a planeswalker's minus", game.Card{Name: "The Eternal Wanderer", TypeLine: "Legendary Planeswalker — Wanderer", OracleID: oracleEternalWanderer, ManaCost: "{4}{W}{W}"}, "−4"},
		{"a planeswalker's ultimate that hits only the opponents' board", game.Card{Name: "Liliana of the Veil", TypeLine: "Legendary Planeswalker — Liliana", OracleID: oracleLilianaVeil, ManaCost: "{1}{B}{B}"}, "−6"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := newThreatBoard(t, game.StepPrecombatMain)
			b.pushActivated(t, b.opp, b.oppCard(tc.card), tc.label)
			b.checkGated(t, true)
		})
	}
}

func TestAbilityThatTouchesOnlyTheirBoardDoesNotThreaten(t *testing.T) {
	t.Run("a pump ability", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		shade := b.oppCard(game.Card{
			Name: "Evernight Shade", TypeLine: "Creature — Shade", OracleID: oracleEvernightShadeHT,
			ManaCost: "{3}{B}", Power: 1, Toughness: 1,
		})
		b.pushActivated(t, b.opp, shade, "")
		b.checkGated(t, false)
	})
	t.Run("an ability with no catalog row", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.pushItem(b.opp, b.oppBears, "Sandbox ability", nil)
		b.checkGated(t, false)
	})
	t.Run("the seat's own wipe", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.pushActivated(t, b.seat, b.pestilence, "1 damage to each creature")
		// Pestilence itself interacts whatever is on the stack, so the
		// gated answers are the probe.
		b.checkGated(t, false)
	})
}
