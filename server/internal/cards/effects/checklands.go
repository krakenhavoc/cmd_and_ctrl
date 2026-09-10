package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// checklands.go — the M10/Ixalan "buddy land" cycle:
//
//	"This land enters tapped unless you control a <A> or a <B>."
//	"{T}: Add {X} or {Y}."
//
// Three of the ten, the three this deck plays. The cycle is data
// and the machinery is one shared condition, so adding the rest
// later is three lines each.
//
// The condition names LAND TYPES, not colours, which is why it reads
// "a Plains or an Island" rather than "a white or blue source": a
// Hallowed Fountain switches on a Glacial Fortress and a Mox Amber
// does not.
//
// Sandbox simplification inherited from controlsLandSubtype: the
// check reads printed type lines, so a continuous type-changing
// effect (Urborg making every land a Swamp) does not turn a
// checkland on. See the note there.
func init() {
	// `look` is the land types the condition names, lowercase —
	// containsFoldASCII folds the haystack and not the needle. `a`
	// and `b` are the colours the land taps for.
	for _, t := range []struct {
		oracleID string
		name     string
		look     []string
		a, b     string
	}{
		{"819fc966-434e-470f-91e9-a38df974ad17", "Drowned Catacomb", []string{"island", "swamp"}, "U", "B"},
		{"027dd013-baa7-4111-b3c9-f4d1414e9c45", "Glacial Fortress", []string{"plains", "island"}, "W", "U"},
		{"7e5d9efe-48a9-434b-bb09-056e0e09cc9a", "Isolated Chapel", []string{"plains", "swamp"}, "W", "B"},
	} {
		// Capture per iteration: the closures below outlive the loop
		// body.
		cond := controlsLandSubtype(t.look...)
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Replacements:  []game.ReplacementEffect{EntersTappedUnless(cond)},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
