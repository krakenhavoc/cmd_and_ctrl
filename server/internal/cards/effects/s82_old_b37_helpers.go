package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b37_helpers.go — shared bodies for the S82 older-sets batch 37
// (Modern Horizons 2, #5501).

// S82OldB37YouCastASpellOtherThanYourFirst is "whenever you cast a spell
// other than your first spell each turn" (Arcbound Tracker). The cast
// path bumps the per-turn tally before it emits EventCast, so a total of
// two or more means this is not the first (compare
// YouCastYourSecondSpellEachTurn).
func S82OldB37YouCastASpellOtherThanYourFirst(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	return g.CastTallyFor(ev.Actor).Total >= 2
}

// S82OldB37ReinforceX is "Reinforce X—{X}{G}{G}": a hand ability whose
// cost carries {X} and which puts X +1/+1 counters on target creature
// (CR 702.77a with a variable amount).
func S82OldB37ReinforceX(cost, label string) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    Plus(ManaCost(cost), DiscardThis()),
		Zones:   []game.ZoneKind{game.ZoneHand},
		Targets: TargetCreature("target creature"),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			x := ctx.X()
			if x <= 0 || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return AddCounter{Target: item.Targets[0].ID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
		},
	}
}
