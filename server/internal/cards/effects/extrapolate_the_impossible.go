package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extrapolate the Impossible — Sorcery {1}{B}:
//
//	"You may reveal exactly two cards you own with different names from
//	 outside the game. An opponent chooses one of them. You put that
//	 card into your hand."
//
// Scryfall rulings (2026-08-21): "The other card stays outside the
// game." and "In a casual game, a card you choose from outside the game
// comes from your personal collection. In a tournament event, a card you
// choose from outside the game must come from your sideboard." A deck's
// sideboard is that pool here (ADR 0191), so the card works whenever the
// deck brought one.
//
// Three beats, each a prompt the engine already knows how to ask:
//
//   - the caster chooses exactly two cards with different names from
//     their sideboard, or none ("you MAY reveal"): a choose_cards over
//     the container with a set rule, so a single card or a pair of the
//     same name is refused, and the enumerator offers only the pairs the
//     engine would accept. The pick is keyed (#4264, extrapolatePick):
//     its set rule is registered with its continuation, so a table
//     waiting on it is a restore point;
//   - both are revealed to the table, then the caster names the opponent
//     who chooses (the engine skips that question at two seats), and
//     that opponent picks one of the two they were just shown (a
//     reveal_pick, whose candidates are public because they were
//     revealed). Both prompts are keyed (#4264, extrapolateOpponent and
//     extrapolateTake): the two revealed cards and the caster ride the
//     prompts as data, so a table waiting on either is a restore point;
//   - the chosen card goes to the caster's hand. The other stays in the
//     sideboard, and stays known to the table, which saw it.
//
// A sideboard with fewer than two differently named cards offers no
// prompt and does nothing. Nothing is shipped stronger than printed.
func init() {
	Register(Spec{
		OracleID:     "1f47c869-f4a9-43b0-aec0-8da4e3a3fec5",
		Name:         "Extrapolate the Impossible",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return extrapolateResolve(ctx)
		},
	})
}

// extrapolateResolve asks the caster for the two cards, if the sideboard
// can supply two with different names.
//
// Caller holds g.mu.
func extrapolateResolve(ctx *Context) error {
	// A deck with no sideboard chooses from the card database (ADR
	// 0198), which always holds two cards with different names; the
	// prompt's own set rule is what keeps the answer to two.
	if ctx.Game.HasSideboardPool(ctx.Controller()) {
		var pool []game.Card
		for _, id := range ctx.Game.OutsideCandidatesForEffect(ctx.Controller(), nil) {
			if c, ok := ctx.Game.LookupOutsideCardForEffect(id); ok {
				pool = append(pool, c)
			}
		}
		if !atLeastTwoNames(pool) {
			return nil
		}
	}
	ctx.Game.QueueOutsidePickForEffect(game.OutsidePickPrompt{
		Player:   ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Extrapolate the Impossible — reveal exactly two cards with different names from outside the game, or none",
		Min:      0,
		Max:      2,
		Keyed:    extrapolatePick,
	})
	return nil
}

// extrapolatePick is the first prompt, keyed (#4264): no filter, the set
// rule "exactly two with different names" (never asked of an empty
// pick), and the reveal that follows. The continuation reads the
// controller and the card off the prompt (a stand-in for the spell, which
// has long resolved) and checks the pair is still outside the game.
var extrapolatePick = game.RegisterOutsidePickRule("choose-cards/extrapolate-the-impossible-reveal",
	nil, exactlyTwoDifferentNames, func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered {
			return nil
		}
		picked := g.StillOutsideForEffect(r.Chooser, r.Picked)
		if len(picked) != 2 {
			return nil
		}
		return extrapolateReveal(g, NewContext(g, &game.StackItem{
			Kind:         game.StackItemSpell,
			Controller:   r.Chooser,
			Owner:        r.Chooser,
			SourceCardID: r.Source,
		}), picked)
	})

// extrapolateReveal reveals the pair and has the caster name the
// opponent who picks one.
func extrapolateReveal(g *game.Game, ctx *Context, picked []uuid.UUID) error {
	g.RevealOutsideForEffect(game.RevealSpec{
		Player: ctx.Controller(),
		Source: ctx.Source(),
		Reason: "Extrapolate the Impossible — reveal two cards from outside the game",
		Cards:  picked,
	})
	// The caster rides first on the carry (OptionPickThen's contract), the
	// revealed pair after.
	g.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
		Chooser:  ctx.Controller(),
		Among:    Opponents.seats(ctx, ctx.Controller()),
		Source:   ctx.Source(),
		Question: "Extrapolate the Impossible — choose an opponent to pick one of the revealed cards",
		ThenKey:  extrapolateOpponent,
		Carry:    append([]uuid.UUID{ctx.Controller()}, picked...),
	})
	return nil
}

// extrapolateOpponent is the second prompt's answer, keyed (#4264): "an
// opponent chooses one of them". The opponent named is the answer, and
// the revealed pair is the carry. Nobody named (every opponent has left,
// CR 800.4a) leaves the card outside the game.
var extrapolateOpponent = OptionPickThen("option-pick/extrapolate-the-impossible-opponent",
	func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil || r.Option.Player == uuid.Nil {
			return nil
		}
		_, err := ctx.Game.RevealPickThenForEffect(game.RevealPickPrompt{
			Chooser:  r.Option.Player,
			Owner:    ctx.Controller(),
			Source:   r.Source,
			Question: "Extrapolate the Impossible — choose one of these cards; it goes into their hand and the other stays outside the game",
			Cards:    r.Carry,
			Min:      1,
			Max:      1,
			ThenKey:  extrapolateTake,
		}, nil)
		return err
	})

// extrapolateTake is the third prompt's answer, keyed (#4264): "you put
// that card into your hand". The caster owns the revealed cards
// (CardsChosen.From); the card not chosen stays where it was.
var extrapolateTake = game.RegisterChooseCardsThen("choose-cards/extrapolate-the-impossible-take",
	func(g *game.Game, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := g.BringFromOutsideForEffect(r.From, id); err != nil {
				return err
			}
		}
		return nil
	})

// exactlyTwoDifferentNames is the set-level rule of the first prompt:
// "exactly two cards you own with different names".
func exactlyTwoDifferentNames(picked []game.Card) bool {
	return len(picked) == 2 && DifferentNames(picked)
}

// atLeastTwoNames reports whether a pool holds two cards with different
// names, which is the least that can answer the first prompt.
func atLeastTwoNames(pool []game.Card) bool {
	for i := range pool {
		for j := i + 1; j < len(pool); j++ {
			if DifferentNames([]game.Card{pool[i], pool[j]}) {
				return true
			}
		}
	}
	return false
}
