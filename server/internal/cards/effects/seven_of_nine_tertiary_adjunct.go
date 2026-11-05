package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seven of Nine, Tertiary Adjunct — {2}{U} Legendary Artifact Creature
// — Borg 1/4 (Star Trek Commander, #4974):
//
//	"That Is Irrelevant — {T}: Draw a card, then discard a card. When
//	 you discard a nonland card this way, put two charge counters on up
//	 to one target artifact."
//
// "That Is Irrelevant" is a flavor word with no rules meaning
// (CR 207.2d). The loot is Glacial Dragonhunt's draw-then-discard with a
// mandatory discard, and only a NONLAND discard creates the CR 603.12
// reflexive trigger, whose "up to one" target is chosen as it goes on
// the stack. Its {T} is a cost, so it cannot be activated the turn Seven
// comes under your control (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22c3b484-0f96-4051-807c-fc0fcb7cd1a8",
		Name:         "Seven of Nine, Tertiary Adjunct",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "That Is Irrelevant — {T}: Draw a card, then discard a card. When you discard a nonland card this way, put two charge counters on up to one target artifact.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
					Ref: sevenOfNineDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
				})
			},
		}},
	})
}

var (
	// The loot's discard half, run once the draw has landed. The key is
	// an on-disk identity.
	sevenOfNineDiscard = game.RegisterDrawThen("seven-of-nine-tertiary-adjunct/discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			Question: "Seven of Nine, Tertiary Adjunct — discard a card",
		}, sevenOfNineDiscarded, nil)
	})

	// The discard's answer, keyed: a nonland discard creates the
	// reflexive trigger. The key is an on-disk identity.
	sevenOfNineDiscarded = game.RegisterDiscardThen("choose-cards/seven-of-nine-tertiary-adjunct-discarded", func(g *game.Game, r game.DiscardsDone) error {
		nonland := false
		for _, id := range r.Discarded.Cards() {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				nonland = true
			}
		}
		if !nonland {
			return nil
		}
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Seven of Nine, Tertiary Adjunct"}
		return WhenYouDo("Seven of Nine, Tertiary Adjunct — put two charge counters on up to one target artifact",
			sevenOfNineChargeBody).Apply(NewContext(g, parent))
	})

	// "Put two charge counters on up to one target artifact."
	sevenOfNineChargeBody = game.ReflexiveBody("seven-of-nine-tertiary-adjunct/charge", simpleBody(sevenOfNineCharge),
		constTargets(func() *game.TargetSpec {
			return TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1)
		}))
)

// sevenOfNineCharge puts two charge counters on the target if it is
// still legal; none chosen is not an error.
func sevenOfNineCharge(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return AddCounter{Target: t.ID, Kind: game.CounterCharge, N: 2}.Apply(ctx)
		}
	}
	return nil
}
