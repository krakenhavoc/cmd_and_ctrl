package effects

import (
	"regexp"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// temporary_target_purpose_test.go — #4302: a change to a target that
// ends with the turn declares which way it cuts, so the bot gives a
// benefit to its own creature and a harm to an opponent's.

func TestUntilEndOfTurnGuard(t *testing.T) {
	entries := func(ts ...game.TargetPurpose) game.Purpose { return ForTargets(ts...) }
	creature := func() *game.TargetSpec { return TargetCreature("target creature") }
	cases := []struct {
		name string
		spec Spec
		want string // "" = accepted
	}{
		{"a keyword grant", Spec{Targets: creature(), Purpose: entries(BenefitUntilEOT(0, 0, 0, "lifelink"))}, ""},
		{"a pump", Spec{Targets: creature(), Purpose: entries(BenefitUntilEOT(0, 2, 2))}, ""},
		{"a direction alone", Spec{Targets: creature(), Purpose: entries(BenefitUntilEOT(0, 0, 0))}, ""},
		{"a shrink", Spec{Targets: creature(), Purpose: entries(HarmUntilEOT(0, -2, -2))}, ""},
		{"a harm alone", Spec{Targets: creature(), Purpose: entries(HarmUntilEOT(0, 0, 0))}, ""},
		{"on an activated row", Spec{Activated: []ActivatedAbility{
			{Label: "a", Targets: creature(), Purpose: entries(BenefitUntilEOT(0, 0, 0, "trample"))},
		}}, ""},
		{"an unknown direction", Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{UntilEndOfTurn: "both"})}, "neither benefit nor harm"},
		{"a pump with no direction", Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{Pump: &game.Pump{Power: 1}})}, "without UntilEndOfTurn"},
		{"a benefit that shrinks", Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Power: -1, Toughness: -1}})}, "that is a harm"},
		{"a harm that pumps", Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{Power: 2}})}, "adds power"},
		{"a harm that grants a keyword", Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{Keywords: []string{"flying"}}})}, "adds power"},
		{"a pump that gives nothing", Spec{Targets: creature(), Purpose: entries(game.TargetPurpose{
			UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{}})}, "gives nothing"},
		{"a capitalised keyword", Spec{Targets: creature(), Purpose: entries(BenefitUntilEOT(0, 0, 0, "Lifelink"))}, "not lowercase"},
		{"on a player", Spec{Targets: TargetPlayer("target player"), Purpose: entries(BenefitUntilEOT(0, 0, 0))}, "cannot target a permanent"},
		{"on a spell", Spec{Targets: TargetSpell("target spell"), Purpose: entries(HarmUntilEOT(0, 0, 0))}, "cannot target a permanent"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.spec.Name = "Test Card"
			var msg string
			func() {
				defer func() {
					if r := recover(); r != nil {
						msg, _ = r.(string)
					}
				}()
				checkSpecPurposes(c.spec)
			}()
			switch {
			case c.want == "" && msg != "":
				t.Errorf("refused: %s", msg)
			case c.want != "" && !strings.Contains(msg, c.want):
				t.Errorf("got %q, want a refusal containing %q", msg, c.want)
			}
		})
	}
}

// untilEndOfTurnRow finds an activated row whose printed text changes a
// target until end of turn: "target creature gains lifelink until end of
// turn", "gets -2/-2 until end of turn", "can't be blocked this turn".
var untilEndOfTurnRow = regexp.MustCompile(`(?i)\btarget\b[^.:]*\b(gains?|gets|loses|can't block|can't be blocked)\b[^.]*\b(until end of turn|this turn)`)

// untilEndOfTurnExempt are the rows the pattern finds that do not declare
// a direction, each for a reason a direction cannot carry.
var untilEndOfTurnExempt = map[string]string{
	// "+X/+X or -X/-X": the controller picks the sign at resolution.
	"Liliana of the Dark Realms": "the sign is chosen at resolution",
	// The third resolution in a turn adds {R}{R}{R}{R}: repeating the
	// grant is the point, and a direction would price the repeats at
	// nothing.
	"Soulbright Seeker": "its repeats add mana",
	// Moves Zack Fair's counters and Equipment onto the creature for good.
	"Zack Fair": "part of what it does lasts",
}

// TestUntilEndOfTurnRowsDeclareTheirDirection holds every catalog
// activated row that changes a target until end of turn to declaring
// which way the change cuts on that target's entry (#4302), and the
// direction to the printed words: "gets -" and "can't block" are harms,
// "gets +", "gains" and "can't be blocked" benefits. A row that declares
// nothing is priced as though its target were removed, which is how
// Heliod, Sun-Crowned gave an opponent's creature lifelink.
func TestUntilEndOfTurnRowsDeclareTheirDirection(t *testing.T) {
	found := 0
	for _, spec := range All() {
		for i, row := range spec.Activated {
			if row.Targets == nil || !untilEndOfTurnRow.MatchString(row.Label) {
				continue
			}
			if _, ok := untilEndOfTurnExempt[spec.Name]; ok {
				continue
			}
			found++
			var e *game.TargetPurpose
			for _, x := range row.Purpose.Targets.List() {
				if x.Slot == 0 {
					e = &x
				}
			}
			if e == nil || e.UntilEndOfTurn == "" {
				t.Errorf("%s ability %d %q declares no UntilEndOfTurn on its target (#4302)", spec.Name, i, row.Label)
				continue
			}
			harm := strings.Contains(row.Label, "gets -") || strings.Contains(row.Label, "can't block")
			if want := map[bool]game.TargetEffect{true: game.TargetHarm, false: game.TargetBenefit}[harm]; e.UntilEndOfTurn != want {
				t.Errorf("%s ability %d %q declares %q, want %q", spec.Name, i, row.Label, e.UntilEndOfTurn, want)
			}
			if pm := e.Pump; pm != nil {
				for _, k := range pm.Keywords {
					if !strings.Contains(strings.ToLower(row.Label), k) {
						t.Errorf("%s ability %d %q declares the keyword %q its text does not grant", spec.Name, i, row.Label, k)
					}
				}
			}
		}
	}
	if found < 70 {
		t.Errorf("found %d rows that change a target until end of turn, want at least the 74 #4302 declared", found)
	}
}

// TestHeliodDeclaresLifelinkForItsTarget: the card that started it.
func TestHeliodDeclaresLifelinkForItsTarget(t *testing.T) {
	s, ok := Lookup("63e596a2-9126-4af6-8782-c38687d664ad")
	if !ok || len(s.Activated) != 1 {
		t.Fatalf("Heliod, Sun-Crowned is not registered with its one activated ability")
	}
	got := s.Activated[0].Purpose.Targets.List()
	if len(got) != 1 || got[0].UntilEndOfTurn != game.TargetBenefit || got[0].Pump == nil ||
		len(got[0].Pump.Keywords) != 1 || got[0].Pump.Keywords[0] != "lifelink" {
		t.Errorf("Heliod's target entry = %+v, want a lifelink benefit until end of turn", got)
	}
}
