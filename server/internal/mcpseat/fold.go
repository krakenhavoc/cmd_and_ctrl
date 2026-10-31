package mcpseat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fold.go — #4737: one card's moves that differ only in which of several
// identical permanents they name are listed once.
//
// Once the binary fetches a cut card in full, a removal spell over a board
// of 296 Treasures is 296 moves, which the board itself prints as one
// entry, "Treasure x296". Listing each would put one window past any
// context budget and tell the agent nothing it can use: the board gives it
// no way to tell those Treasures apart. So the list folds them by the same
// label the board groups permanents under (boardtext.GroupLabel), plus the
// controller, the owner and whether it is a token, and the line says how
// many more there are and which numbers they have. Each is still a numbered
// move of the window that act takes, and legal_moves(match: …), which never
// folds, lists them one by one.

// uuidPattern finds the instance ids in a move's params and line.
var uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// folder computes fold keys against one view.
type folder struct {
	v  *protocol.GameView
	me string
	nw nameWrapper
	// same maps a battlefield permanent's id to the token every permanent
	// interchangeable with it shares. A face-down permanent has none: two
	// of them look alike, but they are not alike, and the seat may know
	// which is which.
	same map[string]string
}

func newFolder(v *protocol.GameView, me string, nw nameWrapper) *folder {
	f := &folder{v: v, me: me, nw: nw, same: map[string]string{}}
	opts := boardOptions(compactZoneCards)
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.InstanceID == "" || (c.FaceDown && !c.FaceVisible) {
			continue
		}
		f.same[c.InstanceID] = "\x01" + c.Controller + "|" + c.Owner + "|" + strconv.FormatBool(c.IsToken) +
			"|" + boardtext.GroupLabel(v, c, opts) + "\x01"
	}
	return f
}

// key is what a move has in common with every move it folds with: its
// type, kind, player, label, params and line, with each permanent it names
// replaced by that permanent's token. named is false for a move that names
// no foldable permanent, which never folds.
func (f *folder) key(m legal.Move) (key string, named bool) {
	src := m.Source.String()
	swap := func(s string) string {
		return uuidPattern.ReplaceAllStringFunc(s, func(id string) string {
			if tok, ok := f.same[id]; ok && id != src {
				named = true
				return tok
			}
			return id
		})
	}
	params := swap(string(m.Params))
	line := swap(moveLine(0, m, f.v, f.me, f.nw))
	return m.Type + "\x00" + string(m.Kind) + "\x00" + m.Player.String() + "\x00" + m.Label + "\x00" + params + "\x00" + line, named
}

// fold splits a card's moves (idx, indices into moves) into the ones the
// list prints and, for each printed one, the moves folded under it.
func (f *folder) fold(moves []legal.Move, idx []int) (shown []int, under map[int][]int) {
	under = map[int][]int{}
	first := map[string]int{}
	for _, i := range idx {
		k, named := f.key(moves[i])
		if named {
			if r, ok := first[k]; ok {
				under[r] = append(under[r], i)
				continue
			}
			first[k] = i
		}
		shown = append(shown, i)
	}
	return shown, under
}

// lines is how many lines a card's moves take once folded.
func (f *folder) lines(moves []legal.Move) int {
	idx := make([]int, len(moves))
	for i := range idx {
		idx[i] = i
	}
	shown, _ := f.fold(moves, idx)
	return len(shown)
}

// foldNote is the tail of a line that stands for more moves than its own.
func foldNote(more []int) string {
	return fmt.Sprintf(" [and %d more like it, each naming another identical permanent: %s]", len(more), moveRanges(more))
}

// moveRanges writes ascending move numbers as runs: "moves 14-308, 312".
func moveRanges(idx []int) string {
	var parts []string
	for i := 0; i < len(idx); {
		j := i
		for j+1 < len(idx) && idx[j+1] == idx[j]+1 {
			j++
		}
		if j > i {
			parts = append(parts, strconv.Itoa(idx[i])+"-"+strconv.Itoa(idx[j]))
		} else {
			parts = append(parts, strconv.Itoa(idx[i]))
		}
		i = j + 1
	}
	word := "moves "
	if len(idx) == 1 {
		word = "move "
	}
	return word + strings.Join(parts, ", ")
}
