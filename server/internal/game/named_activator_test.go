package game

import (
	"testing"

	"github.com/google/uuid"
)

// named_activator_test.go — ADR 0106 §1's 2026-10-07 amendment (#1947):
// the predicate's opponents-only and owner-only arms.
func TestMayActivateNamedActivators(t *testing.T) {
	owner, thief, third := uuid.New(), uuid.New(), uuid.New()
	mine := Card{Owner: owner, Controller: owner}
	stolen := Card{Owner: owner, Controller: thief}
	opponents := ActivatedAbilityShape{OpponentsOnly: true}
	ownerOnly := ActivatedAbilityShape{OwnerOnly: true}
	for _, tc := range []struct {
		name   string
		player uuid.UUID
		c      Card
		zone   ZoneKind
		ab     ActivatedAbilityShape
		want   bool
	}{
		{"opponents-only: its controller", owner, mine, ZoneBattlefield, opponents, false},
		{"opponents-only: an opponent", thief, mine, ZoneBattlefield, opponents, true},
		{"opponents-only: a third player", third, mine, ZoneBattlefield, opponents, true},
		{"opponents-only, stolen: the owner is the thief's opponent", owner, stolen, ZoneBattlefield, opponents, true},
		{"opponents-only, stolen: the thief", thief, stolen, ZoneBattlefield, opponents, false},
		{"owner-only: the owner", owner, mine, ZoneBattlefield, ownerOnly, true},
		{"owner-only: another player", thief, mine, ZoneBattlefield, ownerOnly, false},
		{"owner-only, stolen: the controller", thief, stolen, ZoneBattlefield, ownerOnly, false},
		{"owner-only, stolen: the owner", owner, stolen, ZoneBattlefield, ownerOnly, true},
		{"plain row, stolen: the owner", owner, stolen, ZoneBattlefield, ActivatedAbilityShape{}, false},
		// Off the battlefield CR 108.4a stays the owner's, whatever the row says.
		{"opponents-only in a graveyard: an opponent", thief, mine, ZoneGraveyard, opponents, false},
	} {
		if got := MayActivate(tc.player, tc.c, tc.zone, tc.ab); got != tc.want {
			t.Errorf("%s: MayActivate = %v, want %v", tc.name, got, tc.want)
		}
	}
	if !opponents.ReachesAcross() || !ownerOnly.ReachesAcross() || (ActivatedAbilityShape{}).ReachesAcross() {
		t.Error("ReachesAcross disagrees with the flags")
	}
}
