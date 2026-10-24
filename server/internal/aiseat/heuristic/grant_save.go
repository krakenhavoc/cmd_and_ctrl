package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// grant_save.go — #4312: a benefit until end of turn aimed at the bot's own
// permanent, priced as the save it is when the stack would remove that
// permanent.
//
// answers.go prices an UNTARGETED row on the threatened creature by what
// it saves (ADR 0142 decision 6): a regeneration shield, "this creature
// gains hexproof". A targeted grant cannot declare answers (a row whose
// every announcement must name a target is refused them, ADR 0142 §2),
// so Mother of Runes, Liliana the Faultless, Plaza of Heroes and every
// Blossoming Defense were priced by the attack they add (#4302), which in
// answer to a Murder is nothing. Their target entry already says what
// they grant (TargetPurpose.Pump, Keywords and Toughness), and that is
// enough to read the save off:
//
//   - hexproof, shroud or protection keeps the permanent from what an
//     opponent aims at it: the spell or ability has an illegal target as
//     it resolves (CR 608.2b; CR 702.11b, 702.18a, 702.16b). Not from a
//     sweep, which targets nothing. Protection from "the color of your
//     choice" keeps it only from coloured threats, and the bot names
//     that colour (choices.go, biggestThreatOfColor);
//   - indestructible keeps it from destruction and lethal damage
//     (CR 702.12b), not from exile, a bounce, a sacrifice or -N/-N. A
//     targeted removal that declares no damage is presumed a Murder half
//     the time (destroyShare), as answerSaveOthers presumes;
//   - undying and persist bring it back after it dies, priced as
//     indestructible is, because what dies is what they return;
//   - toughness keeps it from declared damage and -N/-N it outgrows,
//     answerSave's pump test (dies).
//
// The save is worth what answerSave pays: the chance the bot loses the
// permanent times its value. One the bot has already made, an item of its
// own above the threats that targets the same permanent, is not made
// again.

// grantSave is what benefit `pump` saves on the bot's own permanent c from
// the stack. ok is false when it saves nothing there.
func (p *Policy) grantSave(st *state, c *protocol.CardView, pump *protocol.PumpView) (float64, bool) {
	if !p.cfg.PriceGrantSaves || !p.cfg.PriceAnswers || c == nil || pump == nil ||
		c.Controller != st.me || len(st.view.StackItems) == 0 {
		return 0, false
	}
	t, ok := st.threatsTo(c, p.cfg.RemovalConfidence)
	if !ok || st.answeredFor(c.InstanceID, t.top) {
		return 0, false
	}
	share := st.grantShare(c, t, pump)
	if share <= 0 {
		return 0, false
	}
	return share * t.chance * st.permanentValue(c), true
}

// grantShare is the share of threats t that grant `pump` answers on c: 1
// when it answers them all, destroyShare when indestructible meets a
// targeted removal that may not destroy, 0 when it answers none.
func (st *state) grantShare(c *protocol.CardView, t threats, pump *protocol.PumpView) float64 {
	share := 0.0
	for _, k := range pump.Keywords {
		switch k {
		case "hexproof", "shroud":
			if !t.sweep {
				share = 1
			}
		case "protection":
			if !t.sweep && st.threatsShareAColor(c, t.top) {
				share = 1
			}
		case "indestructible", "undying", "persist":
			if t.beyondDestroy || t.minus > 0 {
				continue
			}
			s := 1.0
			if t.removal && !t.sweep {
				s = destroyShare
			}
			share = max(share, s)
		}
	}
	if share < 1 && pump.Toughness > 0 && !t.removal && !dies(c, t, pump) {
		share = 1
	}
	return share
}

// threatsShareAColor reports whether every opposing item on the stack at
// or below index top that targets c comes from a source with a colour
// they all share: the one colour a protection grant can name.
func (st *state) threatsShareAColor(c *protocol.CardView, top int) bool {
	var common []string
	first := true
	for i := 0; i <= top && i < len(st.view.StackItems); i++ {
		it := &st.view.StackItems[i]
		if it.Controller == st.me || !targetsCard(it, c.InstanceID) {
			continue
		}
		src := st.stack[it.SourceCardID]
		if src == nil {
			src = st.bf[it.SourceCardID]
		}
		if src == nil || len(src.Colors) == 0 {
			return false
		}
		if first {
			common, first = slices.Clone(src.Colors), false
			continue
		}
		common = slices.DeleteFunc(common, func(col string) bool { return !slices.Contains(src.Colors, col) })
	}
	return !first && len(common) > 0
}

// answeredFor reports whether the bot already has an item on the stack
// above index top that targets permanent id: the save it made.
func (st *state) answeredFor(id string, top int) bool {
	for i := top + 1; i < len(st.view.StackItems); i++ {
		it := &st.view.StackItems[i]
		if it.Controller == st.me && targetsCard(it, id) {
			return true
		}
	}
	return false
}

// targetsCard reports whether stack item it targets card id.
func targetsCard(it *protocol.StackItemView, id string) bool {
	for _, ref := range it.Targets {
		if ref.ID == id {
			return true
		}
	}
	return false
}
