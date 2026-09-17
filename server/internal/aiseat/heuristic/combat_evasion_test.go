package heuristic

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestCouldBlockEvasion reads only the public, effective CardView fields that
// a policy may inspect. The legal-move enumerator remains authoritative; this
// estimate tells the attack planner whether a defender has an answer at all.
func TestCouldBlockEvasion(t *testing.T) {
	card := func(abilities []string, colors []string, typeLine string, power int) *protocol.CardView {
		return &protocol.CardView{
			TypeLine:  typeLine,
			Abilities: abilities,
			Colors:    colors,
			Power:     power,
		}
	}

	st := &state{view: &protocol.GameView{}}
	for _, tc := range []struct {
		name string
		atk  *protocol.CardView
		blk  *protocol.CardView
		want bool
	}{
		{
			name: "fear permits black blocker",
			atk:  card([]string{"fear"}, []string{"B"}, "Creature", 2),
			blk:  card(nil, []string{"B"}, "Creature", 2),
			want: true,
		},
		{
			name: "fear permits colorless artifact",
			atk:  card([]string{"fear"}, []string{"B"}, "Creature", 2),
			blk:  card(nil, nil, "Artifact Creature", 2),
			want: true,
		},
		{
			name: "fear rejects colorless nonartifact despite black mana cost",
			atk:  card([]string{"fear"}, []string{"B"}, "Creature", 2),
			blk:  &protocol.CardView{TypeLine: "Creature", ManaCost: "{B}", Power: 2},
			want: false,
		},
		{
			name: "intimidate permits one shared color with multicolor attacker",
			atk:  card([]string{"intimidate"}, []string{"U", "R"}, "Creature", 2),
			blk:  card(nil, []string{"U"}, "Creature", 2),
			want: true,
		},
		{
			name: "intimidate permits colorless artifact",
			atk:  card([]string{"intimidate"}, []string{"U", "R"}, "Creature", 2),
			blk:  card(nil, nil, "Artifact Creature", 2),
			want: true,
		},
		{
			name: "intimidate rejects unrelated color nonartifact",
			atk:  card([]string{"intimidate"}, []string{"U", "R"}, "Creature", 2),
			blk:  card(nil, []string{"G"}, "Creature", 2),
			want: false,
		},
		{
			name: "intimidate rejects colorless nonartifact",
			atk:  card([]string{"intimidate"}, []string{"U", "R"}, "Creature", 2),
			blk:  card(nil, nil, "Creature", 2),
			want: false,
		},
		{
			name: "shadow prevents shadow blocker from blocking ordinary attacker",
			atk:  card(nil, nil, "Creature", 2),
			blk:  card([]string{"shadow"}, nil, "Creature", 2),
			want: false,
		},
		{
			name: "shadow prevents ordinary blocker from blocking shadow attacker",
			atk:  card([]string{"shadow"}, nil, "Creature", 2),
			blk:  card(nil, nil, "Creature", 2),
			want: false,
		},
		{
			name: "shadow creatures block each other",
			atk:  card([]string{"shadow"}, nil, "Creature", 2),
			blk:  card([]string{"shadow"}, nil, "Creature", 2),
			want: true,
		},
		{
			name: "horsemanship only restricts attacker",
			atk:  card(nil, nil, "Creature", 2),
			blk:  card([]string{"horsemanship"}, nil, "Creature", 2),
			want: true,
		},
		{
			name: "horsemanship attacker needs horsemanship blocker",
			atk:  card([]string{"horsemanship"}, nil, "Creature", 2),
			blk:  card(nil, nil, "Creature", 2),
			want: false,
		},
		{
			name: "skulk permits equal current power",
			atk:  card([]string{"skulk"}, nil, "Creature", 2),
			blk:  card(nil, nil, "Creature", 2),
			want: true,
		},
		{
			name: "skulk rejects greater current power",
			atk:  card([]string{"skulk"}, nil, "Creature", 2),
			blk:  card(nil, nil, "Creature", 3),
			want: false,
		},
		{
			name: "cant block is a restriction rather than an ability string",
			atk:  card(nil, nil, "Creature", 2),
			blk:  &protocol.CardView{TypeLine: "Creature", Power: 2, Restrictions: []string{"cant_block"}},
			want: false,
		},
		{
			name: "cant be blocked is an attacker restriction",
			atk:  &protocol.CardView{TypeLine: "Creature", Power: 2, Restrictions: []string{"cant_be_blocked"}},
			blk:  card(nil, nil, "Creature", 2),
			want: false,
		},
		{
			name: "literal cant block ability is not a restriction",
			atk:  card(nil, nil, "Creature", 2),
			blk:  card([]string{"can't block"}, nil, "Creature", 2),
			want: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := couldBlock(st, "defender", tc.atk, tc.blk); got != tc.want {
				t.Fatalf("couldBlock() = %v, want %v", got, tc.want)
			}
		})
	}
}
