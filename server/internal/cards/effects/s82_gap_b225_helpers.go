package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b225_helpers.go — shared shapes of the GAP batch 225 cards
// (#5948), in a file of their own so concurrent batches never collide.

// opponentWithMoreLands is "an opponent controls more lands than you" as
// the controller of a trigger or ability reads it.
func opponentWithMoreLands(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == you || p.Eliminated {
			continue
		}
		if B11OpponentControlsMoreThanYou(g, you, p.ID, game.Card.IsLand) {
			return true
		}
	}
	return false
}

// IfAnOpponentControlsMoreLandsThanYou narrows a trigger to the
// intervening "if an opponent controls more lands than you" (CR 603.4):
// it triggers only then, and does nothing if that has stopped being true
// by the time it resolves (Scouting Hawk).
func IfAnOpponentControlsMoreLandsThanYou(t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && opponentWithMoreLands(g, source.Controller) && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: "an opponent controls more lands than you", Test: func(ctx *Context) bool {
		return opponentWithMoreLands(ctx.Game, ctx.Controller())
	}}, t)
}

// ChooseAnotherPermanentYouControl is the resolution-time pick of "you may
// return another permanent you control to its owner's hand" (Rescuer
// Chwinga): nothing targets, zero is the "you may" decline, and the source
// itself is never offered.
func ChooseAnotherPermanentYouControl(g *game.Game, item *game.StackItem, question string, keyed PermanentsThen) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.InstanceID != self {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 1
		},
		Keyed: keyed,
	}.Apply(NewContext(g, item))
}

// GapB225ReturnPicked bounces the permanent Rescuer Chwinga's controller
// picked; one gone from the battlefield (or no pick) does nothing.
var GapB225ReturnPicked = RegisterPermanentsPick("prompt-run/rescuer-chwinga-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !OnBattlefield(ctx.Game, id) {
				continue
			}
			return BounceToHand{Target: id}.Apply(ctx)
		}
		return nil
	},
})

// balanceStep is one of Magus of the Balance's three "same way" instructions
// as a Do step: every player's count of what the instruction is about is read
// as the step begins; each player who has more than the fewest sacrifices or
// discards the excess of their choice (choosing which to keep is choosing
// which to lose, CR 101.4, 608.2d).
type balanceStep struct {
	label   string
	count   func(g *game.Game, player uuid.UUID) int
	discard bool
	spec    *game.TargetSpec
}

func (s balanceStep) Apply(ctx *Context) error {
	var seats []uuid.UUID
	counts := map[uuid.UUID]int{}
	fewest := -1
	for _, id := range ApnapPlayers(ctx.Game) {
		if p := ctx.PlayerByID(id); p == nil || p.Eliminated {
			continue
		}
		seats = append(seats, id)
		n := s.count(ctx.Game, id)
		counts[id] = n
		if fewest < 0 || n < fewest {
			fewest = n
		}
	}
	var asks []uuid.UUID
	for _, id := range seats {
		for i := counts[id]; i > fewest; i-- {
			asks = append(asks, id)
		}
	}
	if len(asks) == 0 {
		return nil
	}
	if s.discard {
		ctx.Game.PlayersDiscardForEffect(asks, game.DiscardPrompt{Source: ctx.Source(), N: 1, Question: s.label})
		return nil
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks, s.spec, s.label)
	return nil
}

// MagusOfTheBalanceSteps are the lands, the cards in hand and the creatures,
// in that order, each balanced to the player who has the fewest.
func MagusOfTheBalanceSteps() []Applier {
	controlled := func(match func(game.Card) bool) func(*game.Game, uuid.UUID) int {
		return func(g *game.Game, player uuid.UUID) int { return CountControlled(g, player, match) }
	}
	return []Applier{
		balanceStep{label: "Magus of the Balance — sacrifice lands down to the fewest any player controls",
			count: controlled(game.Card.IsLand), spec: SacrificeSpec("a land", Land())},
		balanceStep{label: "Magus of the Balance — discard cards down to the fewest any player holds", discard: true,
			count: CardsInHand},
		balanceStep{label: "Magus of the Balance — sacrifice creatures down to the fewest any player controls",
			count: controlled(game.Card.IsCreature), spec: SacrificeSpec("a creature", Creature())},
	}
}
