package effects

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// team_pump_declared_test.go — #4320: an untargeted row whose whole
// effect is "creatures you control get +N/+M and/or gain [keywords] until
// end of turn" declares it as Purpose.TeamPump, so the bot prices it by
// the attack it adds this turn instead of the flat ActivateBase, which
// had Purphoros, God of the Forge pumping a team that never attacked.

func TestTeamPumpGuard(t *testing.T) {
	pump := &game.Pump{Power: 1}
	row := func(p game.Purpose) []ActivatedAbility {
		return []ActivatedAbility{{Label: "a", Cost: ManaCost("{1}"), Purpose: p}}
	}
	cases := []struct {
		name string
		spec Spec
		want string // "" = accepted
	}{
		{"on an activated row", Spec{Activated: row(game.Purpose{TeamPump: pump})}, ""},
		{"a keyword grant", Spec{Activated: row(game.Purpose{TeamPump: &game.Pump{Keywords: []string{"haste"}}})}, ""},
		{"on the card", Spec{Purpose: game.Purpose{TeamPump: pump}}, "off an activated row"},
		{"on a triggered row", Spec{Triggered: []game.TriggeredAbility{{Purpose: game.Purpose{TeamPump: pump}}}}, "off an activated row"},
		{"on an any-player row", Spec{Activated: []ActivatedAbility{{Label: "a", Cost: ManaCost("{1}"), AnyPlayer: true,
			Purpose: game.Purpose{TeamPump: pump}}}}, "off an activated row"},
		{"a shrink", Spec{Activated: row(game.Purpose{TeamPump: &game.Pump{Power: -1}})}, "takes power or toughness away"},
		{"a pump that gives nothing", Spec{Activated: row(game.Purpose{TeamPump: &game.Pump{}})}, "gives nothing"},
		{"a capitalised keyword", Spec{Activated: row(game.Purpose{TeamPump: &game.Pump{Keywords: []string{"Haste"}}})}, "not lowercase"},
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

// teamPumpRow is a row whose first sentence pumps or grants keywords to
// every creature its controller controls until end of turn, printed
// amounts only: "{2}{R}: Creatures you control get +1/+0 until end of
// turn." The tail is what the label says after that sentence.
var teamPumpRow = regexp.MustCompile(`^[^:]+: Creatures you control (?:get \+(\d+)/\+(\d+)(?: and gain ([a-z ,]+))?|gain ([a-z ,]+)) until end of turn\.(.*)$`)

// teamPumpUndeclared are rows the pattern finds that declare no TeamPump,
// each for a reason a team pump cannot carry.
var teamPumpUndeclared = map[string]string{
	// "Each opponent loses 2 life." besides: priced by the pump alone,
	// the drain would be worth nothing outside combat.
	"Sauron, the Lidless Eye": "it does something else besides",
}

// teamPumpTails are the declared rows whose label goes on after the team
// pump, and why that does not change the declaration.
var teamPumpTails = map[string]string{
	// "This ability costs {2} less to activate if you control a
	// legendary creature.": the cost, not the effect.
	"Esquire of the King": "a cost reduction",
	// "Soldiers you control gain vigilance until end of turn.": a grant
	// to a subtype, left out, so the declaration is a floor.
	"The Howling Commandos": "a subtype's grant left out",
	// "Activate only as a sorcery.": a timing restriction, SorcerySpeed
	// on the row, not an effect.
	"Shefet Dunes": "a timing restriction",
}

// TestTeamPumpRowsDeclareTheirPump holds every catalog activated row of
// its controller's own whose text is a team pump until end of turn to
// declaring it, with the printed amounts and keywords (#4320), and every
// declared TeamPump to such a text. Loyalty abilities are left out: a
// planeswalker's +1 is worth its loyalty too, which a pump priced at
// nothing outside combat would lose. So are rows that declare survive or
// protect, which are answers (#4318 holds them for something to answer).
func TestTeamPumpRowsDeclareTheirPump(t *testing.T) {
	found := 0
	check := func(spec Spec, where string, row ActivatedAbility) {
		m := teamPumpRow.FindStringSubmatch(row.Label)
		declared := row.Purpose.TeamPump
		if m == nil {
			if declared != nil {
				t.Errorf("%s %q declares TeamPump, and its text is not a team pump until end of turn", where, row.Label)
			}
			return
		}
		if row.AnyPlayer || row.Cost.Loyalty != nil || row.Purpose.Answers.HasAny(game.AnswerSurvive|game.AnswerProtect) {
			return
		}
		if _, ok := teamPumpUndeclared[spec.Name]; ok {
			if declared != nil {
				t.Errorf("%s %q declares TeamPump and is listed as undeclared", where, row.Label)
			}
			return
		}
		found++
		if declared == nil {
			t.Errorf("%s %q declares no TeamPump (#4320)", where, row.Label)
			return
		}
		if tail := strings.TrimSpace(m[5]); tail != "" {
			if _, ok := teamPumpTails[spec.Name]; !ok {
				t.Errorf("%s %q goes on after its team pump (%q): list it in teamPumpTails or teamPumpUndeclared", where, row.Label, tail)
			}
		}
		power, _ := strconv.Atoi(m[1])
		toughness, _ := strconv.Atoi(m[2])
		keywords := teamPumpKeywords(m[3] + m[4])
		if declared.Power != power || declared.Toughness != toughness || !slices.Equal(declared.Keywords, keywords) || declared.CantBeBlocked {
			t.Errorf("%s %q declares TeamPump %+v, want +%d/+%d %v", where, row.Label, *declared, power, toughness, keywords)
		}
	}
	for _, spec := range All() {
		for i, row := range spec.Activated {
			check(spec, fmt.Sprintf("%s ability %d", spec.Name, i), row)
		}
		for _, g := range spec.Grants {
			for i, row := range g.Activated {
				check(spec, fmt.Sprintf("%s grant %q ability %d", spec.Name, g.Key, i), row)
			}
		}
	}
	if found < 21 {
		t.Errorf("found %d team-pump rows, want at least the 21 #4320 declared", found)
	}
}

// teamPumpKeywords splits "menace, trample, and lifelink" into its
// keywords, in printed order.
func teamPumpKeywords(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(part), "and "))
		for _, k := range strings.Split(part, " and ") {
			if k = strings.TrimSpace(k); k != "" {
				out = append(out, k)
			}
		}
	}
	return out
}

// TestPurphorosDeclaresItsTeamPump: the card that started it.
func TestPurphorosDeclaresItsTeamPump(t *testing.T) {
	s, ok := Lookup("4fdbbec2-e921-4b63-958d-f9ba1e417197")
	if !ok || len(s.Activated) != 1 {
		t.Fatalf("Purphoros, God of the Forge is not registered with its one activated ability")
	}
	if got := s.Activated[0].Purpose.TeamPump; got == nil || got.Power != 1 || got.Toughness != 0 || len(got.Keywords) != 0 {
		t.Errorf("Purphoros's row declares TeamPump %+v, want +1/+0", got)
	}
}
