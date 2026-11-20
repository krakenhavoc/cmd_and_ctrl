package effectstest

import (
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_grant_test.go — the catalog half of ADR 0093's seam: the
// AbilityGrant bundle's mana slot and text, the GrantAbilities
// constructor, the boot-time refusals, and TestEveryGrantKeyResolves.

// grantKeyProblems walks an engine-facing def map and reports every
// layer-6 grant that names a bundle nobody registered, or a bundle
// with a static the layer pass cannot apply (ADR 0093 Decision 10, as
// amended 2026-10-10 and 2026-10-15: game.GrantedStaticProblem).
// Separate from the test so the test can prove it reports both on a
// fixture.
func GrantKeyProblems(all map[string]*game.CardDef) (problems []string, grants int) {
	for owner, d := range all {
		for _, s := range d.Static {
			for _, k := range s.GrantAbilities {
				grants++
				bundle := all[game.GrantKey(k)]
				switch {
				case bundle == nil:
					problems = append(problems, owner+" grants "+k+", which no card registers in Spec.Grants")
				case effects.LayerGrantStaticProblem(bundle) != "":
					problems = append(problems, owner+" grants "+k+", a bundle with a static a layer-6 grant cannot give: "+effects.LayerGrantStaticProblem(bundle))
				default:
					// ADR 0093 amendment 2026-10-13: a grant to cards in a
					// zone gives only activated abilities that function
					// from it; a grant to permanents, none that need
					// another zone.
					zone := s.GrantZone
					if zone == "" {
						zone = game.ZoneBattlefield
					}
					for _, a := range bundle.Activated {
						if !game.AbilityFunctionsFromZone(a, zone) {
							problems = append(problems, owner+" grants "+k+" to cards in "+string(zone)+", but the bundle's ability "+a.Label+" does not function from there")
						}
					}
					if zone != game.ZoneBattlefield && (len(bundle.Triggered) > 0 || len(bundle.Static) > 0 || len(bundle.ManaAbilities) > 0) {
						problems = append(problems, owner+" grants "+k+" to cards in "+string(zone)+", a bundle with more than activated abilities")
					}
				}
			}
		}
	}
	sort.Strings(problems)
	return problems, grants
}

func ExpectRegisterPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("no panic; want one mentioning %q", want)
		}
		if msg, _ := r.(string); !strings.Contains(msg, want) {
			t.Fatalf("panic %v; want one mentioning %q", r, want)
		}
	}()
	fn()
}
