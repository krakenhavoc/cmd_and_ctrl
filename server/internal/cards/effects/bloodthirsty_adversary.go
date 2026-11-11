package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodthirsty Adversary — Creature — Vampire {1}{R}, 2/2:
//
//	"Haste
//	 When this creature enters, you may pay {2}{R} any number of times.
//	 When you pay this cost one or more times, put that many +1/+1
//	 counters on this creature, then exile up to that many target instant
//	 and/or sorcery cards with mana value 3 or less from your graveyard and
//	 copy them. You may cast any number of the copies without paying their
//	 mana costs."
//
// "Any number of times" is MayPayX with a repeated {2}{R}: N is chosen as
// the trigger resolves, then paid through the ordinary "you may pay"
// prompt. Paying one or more times creates a CR 603.12 reflexive trigger,
// which goes on the stack and picks up to N targets then (CR 603.3d), N
// carried as Params.Amount. Its resolution puts the counters on, exiles the
// cards still legal targets through the CR 614 window, and copies the ones
// that really reached exile (CR 707.12) with a free-cast offer each, all
// during the resolution (CR 608.2g). A copy not cast ceases to exist.
func init() {
	Register(Spec{
		OracleID:        "3b943c3e-1f18-4997-b6d0-28ef75aed88e",
		Name:            "Bloodthirsty Adversary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bloodthirsty Adversary — pay {2}{R} any number of times; counters, then copy instants and sorceries from your graveyard", bloodthirstyAdversaryEnters),
		},
	})
}

// bloodthirstyAdversaryPaid is the paid half, keyed (#4289): the number of
// payments rides the prompt.
var bloodthirstyAdversaryPaid = PayUnlessBranches("pay-unless/bloodthirsty-adversary-times",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Bloodthirsty Adversary — " + strconv.Itoa(r.Number) + " +1/+1 counters, then copy up to that many instants and sorceries",
			Body:   cmdr2S42AdversaryBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func bloodthirstyAdversaryEnters(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Symbol: "{2}{R}",
		Label:  "Bloodthirsty Adversary",
		Buys:   "put X +1/+1 counters on it and copy up to X instants and sorceries from your graveyard",
		Goal:   XAsHighAsYouCan,
		Then:   bloodthirstyAdversaryPaid,
	}.Apply(NewContext(g, item))
}

// cmdr2S42AdversaryResolve is the reflexive trigger's resolution: the
// counters, then the exile of the legal picks with a copy-and-cast tail.
func cmdr2S42AdversaryResolve(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	if err := (CounterOnThis{Kind: game.CounterPlusOne, N: p.Amount}).Apply(ctx); err != nil {
		return err
	}
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(cmdr2S42AdversaryCopies))
}

// cmdr2S42AdversaryCopies copies what reached exile and offers every copy
// as a free cast. An on-disk key.
var cmdr2S42AdversaryCopies = CardTail("bloodthirsty-adversary/copy-and-cast", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return CopyAndCast{Cards: o.IDs, Free: true}.Apply(NewContext(ctx.Game, ctx.Item))
})
