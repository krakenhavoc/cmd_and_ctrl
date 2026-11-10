package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b1_helpers.go — shared pieces for the Standard-set leftovers, batches
// 1 and 2 (Edge of Eternities, #4461).

// YouDrawThatManyPlusOneWhileHandAtMost is Quantum Riddler's "As long as
// you have <max> or fewer cards in hand, if you would draw one or more
// cards, you draw that many cards plus one instead". It asks the whole
// draw instruction (game.ReplacementEffect.DrawInstructionExtra), not the
// single draws, so "draw three" is four draws whatever else is in play.
func YouDrawThatManyPlusOneWhileHandAtMost(maxHand int, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:              []game.EventKind{game.EventDrawCard},
		DrawInstructionExtra: 1,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || !ev.DrawInstruction || ev.DrawCount < 1 || src == nil {
				return false
			}
			if ev.DrawPlayer != src.Controller {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DrawPlayer)
			return p != nil && !p.Eliminated && p.Hand.Size() <= maxHand
		},
		Label: label,
	}
}

// PutItsCountersOnTargetCreature is "put its counters on target creature you
// control" for an ability whose source has left (Broodguard Elite,
// Dockworker Drone): every counter the source had as it last existed
// (CR 122.8, 603.10), on the first target still legal as the ability
// resolves, and nothing when there is none.
func PutItsCountersOnTargetCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	left, ok := ctx.TriggeringPermanent()
	if !ok || len(left.Counters) == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return b14PutCounters(ctx, t.ID, left.Counters)
	}
	return nil
}

// GrantedAltCostFirstSpellFree is the key of Weftwalking's offer. An
// on-disk identity (a claim lands on StackItem.AltCost): never renamed,
// never reused.
const GrantedAltCostFirstSpellFree = "granted-first-spell-each-turn-free"

// FirstSpellEachPlayerCastsOnTheirTurnFree is Weftwalking's static: "The
// first spell each player casts during each of their turns may be cast
// without paying its mana cost." An offer with no mana cost is free
// (CR 118.9's second phrasing), X is 0 (CR 107.3b), and the spell keeps its
// own timing. It is made to every player (EveryPlayer), only on their own
// turn and only until they have cast a spell that turn
// (FirstSpellOnOwnTurn), from any zone the printed cost could be paid.
func FirstSpellEachPlayerCastsOnTheirTurnFree() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostFirstSpellFree,
			Label: "Cast it without paying its mana cost",
		},
		EveryPlayer:         true,
		FirstSpellOnOwnTurn: true,
	}
}
