package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// surveil_lands.go — the Murders at Karlov Manor "surveil land"
// cycle:
//
//	"({T}: Add {X} or {Y}.)"
//	"This land enters tapped."
//	"When this land enters, surveil 1."
//
// Eight of the ten, across the decks that reach this file. Like the shocklands
// these are nonbasic lands carrying two basic land types, so the
// mana ability is declared rather than left to the synthetic shape.
//
// S22: the surveil is real. These shipped in an earlier sprint with
// the enters-tapped half and the mana only, plus a long comment
// explaining that the surveil — the whole reason to play one of
// these over a Guildgate — was not implemented, because it needed a
// PendingChoice kind that did not exist. The Surveil primitive
// closes that: the ETB trigger queues the real look-at-one prompt,
// and a reanimator deck gets the graveyard card it was promised.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"ccfb8b4d-651c-418a-aa19-cb23105b3f2f", "Meticulous Archive", "W", "U"},
		{"216a2a92-9ca3-4ca3-8af7-686c13b04290", "Shadowy Backstreet", "W", "B"},
		{"08d80efc-9542-4ba2-824c-c8615d8d07f2", "Undercity Sewers", "U", "B"},
		{"d51831b1-7394-456e-a1de-6787a59f5932", "Lush Portico", "G", "W"},
		{"840119bf-e60f-4ff7-9c9b-d420d09df545", "Underground Mortuary", "B", "G"},
		{"04e5e84f-8fd4-43ab-8f9d-5b24646f7ae5", "Raucous Theater", "B", "R"},
		{"d2bcff58-7a8a-46ef-b6b3-39501d4c8e6e", "Thundering Falls", "U", "R"},
		{"b33656ae-3473-4223-845f-f9147f87678b", "Commercial District", "R", "G"},
	} {
		// Bind per iteration: the closures below outlive the loop.
		name := t.name
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          name,
			Completeness:  CompletenessFull,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
			Triggered: []game.TriggeredAbility{
				WhenThisEnters(name+" — surveil 1", Do(Surveil{N: 1})),
			},
		})
	}
}
