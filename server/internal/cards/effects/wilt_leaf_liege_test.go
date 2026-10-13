package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wilt_leaf_liege_test.go — ADR 0153 PR 2 (#3168).

const wiltLeafLiegeOracle = "2441696b-a9ba-4813-ba2e-e71f85281d05"

func TestWiltLeafLiegeEntersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	liege := d.handCard(d.victim, "Wilt-Leaf Liege", "Creature — Elf Knight", "{1}{G/W}{G/W}{G/W}", wiltLeafLiegeOracle, 4, 4)
	filler := d.filler(d.victim)

	d.theyDiscard(t, liege, filler)

	if _, ok := d.onBattlefield(liege); !ok {
		t.Fatalf("the Liege is in %q, want the battlefield", d.zoneOf(liege))
	}
}

func TestWiltLeafLiegeStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := newDBOTable(t)
	liege := d.handCard(d.caster, "Wilt-Leaf Liege", "Creature — Elf Knight", "{1}{G/W}{G/W}{G/W}", wiltLeafLiegeOracle, 4, 4)
	filler := d.filler(d.caster)

	d.iDiscard(t, liege, filler)

	if !d.caster.Graveyard.Contains(liege) {
		t.Fatalf("the Liege is in %q, want the graveyard", d.zoneOf(liege))
	}
}

// Both anthems apply on their own: other green creatures and other white
// creatures you control get +1/+1, a green-white one gets +2/+2, and a
// creature of another colour or an opponent's gets nothing.
func TestWiltLeafLiegeAnthems(t *testing.T) {
	d := newDBOTable(t)
	liege := pushBattlefieldCardWithTimestamp(d.g, game.Card{
		InstanceID: uuid.New(), Name: "Wilt-Leaf Liege", TypeLine: "Creature — Elf Knight",
		OracleID: wiltLeafLiegeOracle, Power: 4, Toughness: 4, Colors: []string{"G", "W"},
		Owner: d.caster.ID, Controller: d.caster.ID,
	})
	mk := func(owner *game.Player, colors ...string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(d.g, game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Colors: colors, Owner: owner.ID, Controller: owner.ID,
		})
	}
	green, white, both := mk(d.caster, "G"), mk(d.caster, "W"), mk(d.caster, "G", "W")
	red, theirs := mk(d.caster, "R"), mk(d.victim, "G")

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"green", green, 3}, {"white", white, 3}, {"green-white", both, 4},
		{"red", red, 2}, {"an opponent's green", theirs, 2}, {"the Liege itself", liege, 4},
	} {
		if got := effectivePower(t, d.g, tc.id); got != tc.want {
			t.Errorf("%s creature has power %d, want %d", tc.name, got, tc.want)
		}
	}
}
