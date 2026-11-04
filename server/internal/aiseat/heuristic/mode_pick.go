package heuristic

import (
	"slices"
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mode_pick.go — #4674: a modal trigger's bullets, chosen as the ability
// is put on the stack (CR 603.3c).
//
// Before #4674 every answer to a mode_pick scored the same, so the bot
// took the enumerator's first. On "choose up to one" (Ertai Resurrected,
// Hullbreaker Horror, Ravager Wurm) and "choose up to X" (The Ruinous
// Wrecking Crew) the first answer is none (CR 700.2b), so the bot never
// chose a mode at all.
//
// The prompt now carries each offered bullet as a modal card's options
// are carried (`mode_bullets`): the purpose the catalog declares for it
// and, for a targeted bullet, its legal targets as the trigger would be
// judged now, with the clause's `harmful` flag. A bullet is priced as
// the rest of the policy prices the same things:
//
//   - its purpose by purposeValue, the price a cast's mode gets;
//   - each target clause by the best legal pick for it, priced as the
//     pick_target that follows will price it. A clause declared harmful
//     (game.TargetSpec.Harmful) is priced by opponentsChoiceValue, so a
//     removal bullet whose only legal target is the bot's own permanent
//     costs what losing it costs. Any other clause says nothing about
//     which way it cuts: it is priced by targetsValue only when it cannot
//     pick anything of the bot's own ("target opponent", "target spell
//     you don't control"), and at nothing when it can (Charming Prince's
//     blink of "another target creature you own", a pump on "target
//     creature"), so an undeclared bullet is neither chosen nor avoided
//     for its target.
//
// An answer is worth the sum of its bullets. The empty answer is worth
// nothing, and ties go to the enumerator's earlier answer, so the bot
// takes every bullet that helps, takes no bullet worth nothing, and
// answers none when every bullet hurts it. A bullet that declares no
// purpose and no target (The Ruinous Wrecking Crew's symmetric edict)
// is worth nothing and is left out of an "up to".
//
// What a bullet gives an opponent besides its target (Ertai's "its
// controller draws a card") is not declared, and so not priced; nor is
// whether a fight's own creature survives (Ravager Wurm).

// choiceModePick is the mode_pick kind on the wire.
const choiceModePick = "mode_pick"

// modePickValue scores one answer to a mode_pick: the sum of what each
// chosen bullet is worth. With PriceModePicks off, or a prompt that does
// not carry its bullets, every answer scores 0, as before #4674.
func (p *Policy) modePickValue(st *state, ch *protocol.PendingChoiceView, modes []int) (float64, string) {
	if !p.cfg.PriceModePicks || ch == nil || len(ch.ModeBullets) == 0 || len(ch.ModeBullets) != len(ch.ModeIndexes) {
		return 0, "unrecognised choice"
	}
	if len(modes) == 0 {
		return 0, "mode pick: choose none"
	}
	bullets := make(map[int]*protocol.ModeOptionView, len(ch.ModeIndexes))
	for j, i := range ch.ModeIndexes {
		bullets[i] = &ch.ModeBullets[j]
	}
	var v float64
	for _, m := range modes {
		if b := bullets[m]; b != nil {
			v += p.modeBulletValue(st, b)
		}
	}
	if v <= 0 {
		return v, "mode pick: no bullet here helps"
	}
	return v, "mode pick: the bullets that help"
}

// modeBulletValue is what one bullet does for the bot: its declared
// purpose, plus the best legal pick for each of its target clauses.
func (p *Policy) modeBulletValue(st *state, b *protocol.ModeOptionView) float64 {
	var v float64
	var ps purposeSet
	ps.add(b.Purpose)
	if p.purposePriced(ps) {
		v += p.purposeValue(st, ps, 0, nil, false)
	}
	clauses := b.Clauses
	if len(clauses) == 0 && b.LegalTargets != nil {
		clauses = []protocol.LegalTargetsView{*b.LegalTargets}
	}
	for i := range clauses {
		v += st.bestClausePicks(p.cfg, &clauses[i])
	}
	return v
}

// bestClausePicks is what the best legal picks for one target clause are
// worth: its Min required picks, the best first, then any more up to Max
// that are worth something. A clause the bot may leave empty never costs
// it anything, and one that is not declared harmful but can pick the
// bot's own is worth nothing (its direction is unknown).
func (st *state) bestClausePicks(cfg Config, c *protocol.LegalTargetsView) float64 {
	if !c.Harmful && st.clauseReachesMine(c) {
		return 0
	}
	price := func(t targetRef) float64 {
		if c.Harmful {
			return st.opponentsChoiceValue(cfg, st.me, []targetRef{t})
		}
		return st.targetsValue(cfg, []targetRef{t})
	}
	vals := make([]float64, 0, len(c.Players)+len(c.Cards))
	for _, id := range c.Players {
		vals = append(vals, price(targetRef{Kind: "player", ID: id}))
	}
	for _, id := range c.Cards {
		vals = append(vals, price(targetRef{Kind: "card", ID: id}))
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(vals)))
	hi := c.Max
	if hi <= 0 || hi > len(vals) {
		hi = len(vals)
	}
	var v float64
	for i := 0; i < hi; i++ {
		if i >= c.Min && vals[i] <= 0 {
			break
		}
		v += vals[i]
	}
	return v
}

// clauseReachesMine reports whether a target clause can pick the bot
// itself or anything of its own: a permanent or a stack object it
// controls, or a card it owns in a graveyard or exile.
func (st *state) clauseReachesMine(c *protocol.LegalTargetsView) bool {
	if slices.Contains(c.Players, st.me) {
		return true
	}
	for _, id := range c.Cards {
		if x := st.bf[id]; x != nil && x.Controller == st.me {
			return true
		}
		if x := st.stack[id]; x != nil && x.Controller == st.me {
			return true
		}
		if x := st.graveyard[id]; x != nil && x.Owner == st.me {
			return true
		}
		if x := st.exile[id]; x != nil && x.Owner == st.me {
			return true
		}
	}
	return false
}
