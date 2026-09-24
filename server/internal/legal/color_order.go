package legal

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// color_order.go — #986. One ordering for a "choose a color" prompt's
// five answers, read by the two places that put those answers in front
// of somebody.
//
// CR 105.4 makes every one of the five a LEGAL answer, so the option
// list says nothing at all about which one the card wants. #780 gave
// the prompt the card's own declaration of what it will DO with the
// colour (game.ColorPurpose) and PR #985 taught the heuristic policy to
// switch on it — but the ORDER the answers are offered in was still the
// pre-#780 one for every purpose: "whatever this seat has most of on
// the battlefield". Wash Out offered a mono-green bot green first.
//
// Two readers, one function, and that is the point:
//
//   - `enumerator.colorAnswers` (choices.go), which is the move list a
//     bot seat sees. The `random` tier picks uniformly and the
//     `decideChoice` tie-break takes the FIRST offered answer, so the
//     order is the whole decision on any prompt the policy scores flat;
//   - `protocol.ViewOfGame`'s `choose_color` projection, which is the
//     order a HUMAN's colour buttons render in. The client has
//     rendered `color_options` in the order sent since the ADR 0040
//     addendum settled it for a mana pick ("Clients render the buttons
//     in the order sent"), so extending the same contract to
//     choose_color puts the sensible answer under the cursor without
//     the client learning a second ranking rule.
//
// A HOOK would be the wrong shape here, and that is the difference
// from #1014's `Options.OrderTargets`. Ranking a board by what a
// SPELL is worth against it is a policy question, which is why the
// target ordering is injected by the seat; this is a reading of the
// card's own printed text against public counts, it has to be the same
// for a bot and for a human because both get it off the same prompt,
// and `legal` is the layer both already read. There is no second
// scorer: the heuristic still prices each answer with its own Weights
// (`colorChoiceValue`), and this only decides what order it sees them
// in — and therefore what it does when it scores two of them the same.
//
// PUBLIC INFORMATION ONLY. Every term below is a count of permanents on
// the battlefield, which every seat can see. Nothing reads a hand, a
// library or a graveyard — ADR 0033 §3 — so the order can ride the wire
// to every viewer without being a leak, and a human reading a bot's
// buttons learns nothing they could not count themselves.

// colorBoard is the one walk of the battlefield every purpose scores
// from. Built once per prompt.
type colorBoard struct {
	// mine counts the permanents the chooser controls of each colour.
	// A multicoloured permanent counts once per colour, which is what
	// EffectiveColors says it is.
	mine map[string]int
	// theirs is the same for every OTHER seat's permanents.
	theirs map[string]int
	// threat is the greatest power among the CREATURES other seats
	// control of each colour — "the biggest thing pointed this way",
	// which is the question a protection colour actually asks.
	threat map[string]int
}

// colorBoardOf walks the battlefield once. Post-layer characteristics
// throughout (EffectiveColors, CurrentPower — #1281: Effective().Power
// excludes +1/+1 / -1/-1 counters, which is not what "the biggest
// thing pointed this way" means), because a Kenrith's Transformation'd
// creature is green to the player looking at it.
//
// Caller must hold g's read lock.
func colorBoardOf(g *game.Game, chooser uuid.UUID) colorBoard {
	b := colorBoard{
		mine:   map[string]int{},
		theirs: map[string]int{},
		threat: map[string]int{},
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		colors := c.EffectiveColors()
		if len(colors) == 0 {
			continue
		}
		if c.Controller == chooser {
			for _, col := range colors {
				b.mine[col]++
			}
			continue
		}
		for _, col := range colors {
			b.theirs[col]++
		}
		if !c.IsCreature() {
			continue
		}
		power := c.CurrentPower()
		for _, col := range colors {
			if power > b.threat[col] {
				b.threat[col] = power
			}
		}
	}
	return b
}

// OrderColorOptionsLocked ranks `options` for the seat that has to
// answer, highest first. The sort is STABLE, so colours that score
// equal keep the order they arrived in — WUBRG for a full prompt, the
// card's own narrowed list for "a color other than blue" — and two
// projections of one board always produce the same buttons.
//
// It never adds, removes or rewrites an option: every answer the engine
// would accept is still offered, in some order. A nil game, an unseated
// chooser or fewer than two options short-circuits to a copy.
//
// Caller must hold g's read lock (both callers build inside one
// ReadSnapshot; see EnumerateLocked's comment for why taking a second
// one would deadlock).
func OrderColorOptionsLocked(g *game.Game, chooser uuid.UUID, options []string, purpose game.ColorPurpose) []string {
	out := append([]string(nil), options...)
	if g == nil || chooser == uuid.Nil || len(out) < 2 {
		return out
	}
	b := colorBoardOf(g, chooser)
	score := colorScorerFor(b, purpose)
	sort.SliceStable(out, func(i, j int) bool { return score(out[i]) > score(out[j]) })
	return out
}

// colorScorerFor is the switch: one arm per game.ColorPurpose, and a
// default that is the pre-#986 rule unchanged.
//
// The arms are deliberately the SAME QUESTIONS the heuristic's
// `colorChoiceValue` asks (aiseat/heuristic/choices.go), answered with
// the crudest public proxy there is. They are not meant to beat the
// policy — the policy re-ranks — they are meant to be a sane answer for
// the seats that have no policy: the `random` tier, a tie-broken
// heuristic, and the human reading the buttons.
func colorScorerFor(b colorBoard, purpose game.ColorPurpose) func(string) int {
	switch purpose {
	case game.ColorForHarm:
		// Wash Out: everything of the named colour is punished, the
		// chooser's own permanents included. What the opposition loses
		// NET of what the chooser does — which is the difference
		// between bouncing their board and bouncing your own.
		return func(c string) int { return b.theirs[c] - b.mine[c] }

	case game.ColorForFilter:
		// Oona: the colour picks which of somebody ELSE's cards the
		// effect acts on and nothing of the chooser's is at stake, so
		// the chooser's own board is not a term at all. What the
		// opponents have SHOWN is the only public proxy for what they
		// are made of.
		return func(c string) int { return b.theirs[c] }

	case game.ColorForProtection:
		// Mother of Runes, Story Circle: the colour is the one being
		// defended AGAINST, so it is the colour of the biggest thing
		// somebody else could point at you. Power rather than a count,
		// because one 8/8 is the reason you hold the ability up and
		// four 1/1s are not.
		return func(c string) int { return b.threat[c] }
	}
	// ColorForMana, ColorForBenefit, and a prompt that declares
	// nothing: the colour this seat already has most of. Right for
	// Coldsteel Heart (the mana you will actually spend) and for
	// Heraldic Banner (the creatures the anthem will actually pump),
	// and it is exactly what colorAnswers did before this file existed,
	// so an unannotated prompt is ordered no worse than yesterday.
	return func(c string) int { return b.mine[c] }
}
