package game

import "testing"

// A one-spell permission (Locke) is spent only when it is the REASON the
// cast is legal. With the card's own condition true, its text casts it
// and the permission survives (the ruling TestALimitedGrantIsNotSpent
// ByTheCardsOwnPermission pins for an unconditional zone); with the
// condition false, the permission is what opened the zone and is spent.
func TestALimitedGrantIsSpentOnlyWhileAConditionalZoneIsClosed(t *testing.T) {
	for _, tc := range []struct {
		name      string
		speed     int
		wantSpent bool
	}{
		{"condition true", MaxSpeed, false},
		{"condition false", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			advanceTo(t, g, StepPrecombatMain)
			withConditionalCastZones(t, conditionalZoneOracle)
			id := seedGraveyard(me, "Test Lightwheel", "Sorcery", "{W}")
			g.WithWriteLock(func() {
				for i := range me.Graveyard.Cards {
					if me.Graveyard.Cards[i].InstanceID == id {
						me.Graveyard.Cards[i].OracleID = conditionalZoneOracle
					}
				}
				c, _ := g.LookupCardForEffect(id)
				g.GrantCastPermissionToCardsForEffect(CastPermission{
					Player: me.ID, Zone: ZoneGraveyard, CastOnly: true, CastsLeft: 1, Label: "Test Locke",
				}, []Card{c})
			})
			g.SetSpeedForTest(me.ID, tc.speed)
			if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard"}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			spent := len(me.CastPermissions) == 0
			if spent != tc.wantSpent {
				t.Errorf("permission spent = %v, want %v (permissions %+v)", spent, tc.wantSpent, me.CastPermissions)
			}
		})
	}
}
