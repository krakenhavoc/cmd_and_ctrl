package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// return_you_control.go — "return a <permanent> you control to its
// owner's hand", the untargeted self-bounce the bounce lands (Azorius
// Chancery and the karoos), Whitemane Lion and Time Wipe print.
//
// The sentence names no target, so it is a choice made while the
// ability RESOLVES (CR 608.2) rather than when it is put on the stack
// (CR 601.2c / 603.3d): opponents do not see which permanent is coming
// back until it comes back, hexproof and shroud are irrelevant to it,
// and nothing can make it fizzle. That is ChoosePermanents' own-board
// shape with a floor and a ceiling of one — the engine's
// own_permanents prompt, the one Scapeshift asks with — plus a bounce
// of whatever was picked.
//
// It was written as a target clause across the catalog before #1214
// gave a resolution-time pick over one's own permanents a home, and
// every card that did so carries a caveat saying so.

// ReturnOneYouControl asks the resolving effect's controller to pick
// one permanent they control that passes Match, and returns it to its
// owner's hand. Mandatory: with at least one candidate the prompt has
// a floor of one. With none there is nothing to ask and nothing
// happens — "only as much as possible", CR 609.3.
type ReturnOneYouControl struct {
	// Match is the clause's noun — MatchLand for "a land you
	// control", MatchCreature for "a creature you control". Nil means
	// any permanent.
	Match func(game.Card) bool

	// Question is the prompt header: "<card> — return a land you
	// control to its owner's hand".
	Question string
}

func (r ReturnOneYouControl) Apply(ctx *Context) error {
	match := r.Match
	return ChoosePermanents{
		Question: r.Question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && (match == nil || match(c)) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 1, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			for _, id := range picked.Cards() {
				if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
}
