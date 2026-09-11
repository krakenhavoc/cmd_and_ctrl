package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// surveil_lands.go — the Murders at Karlov Manor "surveil land"
// cycle:
//
//	"({T}: Add {X} or {Y}.)"
//	"This land enters tapped."
//	"When this land enters, surveil 1."
//
// Three of the ten, the three this deck plays. Like the shocklands
// these are nonbasic lands carrying two basic land types, so the
// mana ability is declared rather than left to the synthetic shape.
//
// # Declared sandbox simplification: NO SURVEIL
//
// The enters-tapped half is real (a CR 614 self-replacement) and the
// mana is real. The surveil is NOT IMPLEMENTED, and that is the
// whole reason these cards are worth playing over a Guildgate — a
// reanimator deck surveils to fill its graveyard, not to smooth a
// draw.
//
// Surveil needs a new PendingChoice kind: it is scry's structure
// with "bottom of library" replaced by "graveyard", which means a
// new prompt, a new resolver, a new clone leg and a new client
// dialog. Reusing PendingChoiceScry would be wrong in the direction
// that matters — a card put on the bottom is not a card in the
// graveyard, and this deck's whole plan is the difference.
//
// Shipping them anyway is a deliberate call: a tapped dual that taps
// for the right colours is most of what a land does on most turns,
// and the alternative is that the card stays out of the catalog and
// the player does the enters-tapped by hand as well. But the AUTO
// badge on these three is over-promising until surveil lands, so it
// is written down here, on the card, and in the decklist doc.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"ccfb8b4d-651c-418a-aa19-cb23105b3f2f", "Meticulous Archive", "W", "U"},
		{"216a2a92-9ca3-4ca3-8af7-686c13b04290", "Shadowy Backstreet", "W", "B"},
		{"08d80efc-9542-4ba2-824c-c8615d8d07f2", "Undercity Sewers", "U", "B"},
	} {
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          t.name,
			Completeness:  CompletenessCaveats,
			Caveats:       []string{"The land enters tapped and taps for both colors, but the \"surveil 1\" on entry never happens."},
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
		})
	}
}
