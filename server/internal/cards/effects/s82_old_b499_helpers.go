package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b499_helpers.go — shared pieces of the S82 Magic 2010 batch 499
// cards (#5585).

// OldB499PreventOneOfTheDamageToYou is "If a source an opponent controls would
// deal damage to you, prevent 1 of that damage" (Guardian Seraph): damage of
// 1 is prevented entirely, larger damage loses 1 (CR 615). It reads the
// source's last-known controller, so a spell's or a dead creature's damage
// counts as the opponent's.
func OldB499PreventOneOfTheDamageToYou(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true, // CR 615.1a
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget != src.Controller {
				return false
			}
			return ev.SourceLKI != nil && ev.SourceLKI.Controller != uuid.Nil && ev.SourceLKI.Controller != src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			if ev.DamageAmount <= 1 {
				ev.Cancel()
				return nil
			}
			ev.DamageAmount--
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      name + " — prevent 1 of that damage",
		RowLabel:   "Prevent 1 of the damage a source an opponent controls would deal to you.",
	}
}

// OldB499CantBlockTargets makes every card target still legal as the
// resolving spell finishes unable to block until end of turn (Panic Attack).
func OldB499CantBlockTargets(ctx *Context, name string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: name + " — can't block"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// OldB499ToughnessLessThanPower is "with toughness less than this
// creature's power" (Stone Giant): the candidate's toughness against the
// source's power, judged as the target is chosen and again as the ability
// resolves (CR 608.2b), against last-known power (CR 608.2h).
func OldB499ToughnessLessThanPower() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourceToughnessVsPower, Cmp: game.CmpLess}
}

// oldB499DestroyTheObjectBody is a delayed trigger that destroys the object it
// names, if it is still that object on the battlefield (CR 603.7c). An
// on-disk key: never renamed, never reused.
var oldB499DestroyTheObjectBody = game.DelayedBody("destroy/old-b499-the-object", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !OnBattlefield(g, p.Object.ID) {
		return nil
	}
	return DestroyTarget{Target: p.Object.ID}.Apply(NewContext(g, item))
})

// OldB499FlyThenDestroyAtEndStep is Stone Giant's resolution: the chosen
// creature gains flying until end of turn, and is destroyed at the beginning
// of the next end step.
func OldB499FlyThenDestroyAtEndStep(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: label + " — flying"}).Apply(ctx); err != nil {
			return err
		}
		return ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  label + " — destroy " + c.Name,
			Cards:  []uuid.UUID{id},
			Body:   oldB499DestroyTheObjectBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		}.Apply(ctx)
	}
}

// OldB499SphinxSearch is Sphinx Ambassador's combat-damage trigger: the
// controller searches the damaged player's library for a card (it stays
// where it is, unrevealed), that player names a card, and a creature card
// that doesn't have that name may be put onto the battlefield under the
// controller's control; then that player shuffles.
func OldB499SphinxSearch(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	owner := item.Trigger.Event.Target
	t := ctx.Tail(oldB499SphinxFound)
	t.Carry = []uuid.UUID{owner}
	return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:       ctx.Controller(),
		LibraryOwner: owner,
		Source:       ctx.Source(),
		Dest:         game.ZoneLibrary,
		HoldFound:    true,
		Limit:        1,
		Reason:       "Sphinx Ambassador — search that player's library for a card",
		ThenTail:     t,
	})
}

// oldB499SphinxFound: the search is done; the searched player names a card.
// Carry is [library owner, found card…]. An on-disk key.
var oldB499SphinxFound = game.RegisterTail("then/old-b499-sphinx-found", func(g *game.Game, o game.TailOutcome) error {
	owner := o.Tail.Carry[0]
	if len(o.IDs) == 0 {
		return g.ShuffleLibraryForEffect(owner)
	}
	next := o.Tail
	next.Ref = oldB499SphinxNamed
	next.Carry = []uuid.UUID{owner, o.IDs[0]}
	g.QueueCardNameChoiceKeyedForEffect(owner, o.Tail.Source, "Sphinx Ambassador — choose a card name", next)
	return nil
})

// oldB499SphinxNamed: a creature card without the chosen name may be put
// onto the battlefield under the searcher's control. An on-disk key.
var oldB499SphinxNamed = game.RegisterTail("then/old-b499-sphinx-named", func(g *game.Game, o game.TailOutcome) error {
	owner, found := o.Tail.Carry[0], o.Tail.Carry[1]
	c, ok := g.LookupCardForEffect(found)
	if !ok || !c.IsCreature() || game.CardNameMatches(c, o.Tail.Word) {
		return g.ShuffleLibraryForEffect(owner)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    o.Tail.Player,
		FromPlayer: owner,
		Source:     o.Tail.Source,
		Question:   "Sphinx Ambassador — you may put " + c.Name + " onto the battlefield under your control",
		Cards:      []uuid.UUID{found},
		Min:        0,
		Max:        1,
		ThenKey:    oldB499SphinxPicked,
		Carry:      []uuid.UUID{owner},
		LiveZone:   game.ZoneLibrary,
		Asked:      o.Tail.Asked,
	})
	return nil
})

// oldB499SphinxPicked puts the pick onto the battlefield under the chooser's
// control, then the searched player shuffles. An on-disk key.
var oldB499SphinxPicked = game.RegisterChooseCardsThen("choose-cards/old-b499-sphinx-picked", func(g *game.Game, r game.CardsChosen) error {
	owner := r.Carry[0]
	if len(r.Picked) == 0 {
		return g.ShuffleLibraryForEffect(owner)
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: r.Picked[0], From: game.ZoneLibrary}},
		game.ZoneEntryOptions{Controller: r.Chooser},
		game.KeyedTail{Ref: oldB499ShuffleCarried, Player: r.Chooser, Source: r.Source, Carry: []uuid.UUID{owner}})
})

// oldB499ShuffleCarried shuffles the library of the player in Carry[0].
// An on-disk key.
var oldB499ShuffleCarried = game.RegisterTail("then/old-b499-shuffle-carried", func(g *game.Game, o game.TailOutcome) error {
	return g.ShuffleLibraryForEffect(o.Tail.Carry[0])
})
