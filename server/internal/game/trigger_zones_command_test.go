package game

import "testing"

// #2802: the command zone is a declared trigger zone, and the
// battlefield may be named only beside another zone (an eminence
// trigger's {battlefield, command}).
func TestTriggerZonesAcceptTheEminenceShape(t *testing.T) {
	for _, tc := range []struct {
		zones []ZoneKind
		ok    bool
	}{
		{[]ZoneKind{ZoneCommand}, true},
		{[]ZoneKind{ZoneBattlefield, ZoneCommand}, true},
		{[]ZoneKind{ZoneBattlefield}, false},
		{[]ZoneKind{ZoneBattlefield, ZoneLibrary}, false},
		{[]ZoneKind{ZoneStack}, false},
	} {
		zone, why := TriggerZonesUnsupported(tc.zones)
		if (why == "") != tc.ok {
			t.Errorf("TriggerZonesUnsupported(%v) = %q, %q; want ok=%v", tc.zones, zone, why, tc.ok)
		}
	}
}
