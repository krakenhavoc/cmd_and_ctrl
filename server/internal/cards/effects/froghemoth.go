package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Froghemoth — Creature — Frog Horror {3}{G}{G}, 4/4 (Adventures in the
// Forgotten Realms):
//
//	"Trample, haste
//	 Whenever this creature deals combat damage to a player, exile up to
//	 that many target cards from their graveyard. Put a +1/+1 counter on
//	 this creature for each creature card exiled this way. You gain 1 life
//	 for each noncreature card exiled this way."
//
// The clause is built per trigger (TargetsFrom, handed the trigger's event):
// at most as many target cards as the damage dealt, all owned by the player
// it was dealt to, chosen as the trigger goes on the stack (CR 603.3d). A
// damage amount of zero leaves the trigger without a target clause. Each
// target still legal as it resolves is exiled (CR 608.2b) with the "if it was
// a creature card" tail (ExileThenIfItWas, judged on the card as it was in
// the graveyard): a creature card puts a +1/+1 counter on Froghemoth, any
// other card gains 1 life. The tail runs only for a card that reached exile,
// so "for each ... exiled this way" counts exactly those. A Froghemoth that
// has left puts no counter on (CR 400.7); the life is still gained.
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66e66543-1f56-4e62-9183-3804843db101",
		Name:            "Froghemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: ThisDealtCombatDamageToAPlayer,
			Key:       froghemothLabel,
			TargetsFrom: func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
				n, victim := tc.Event.Amount, tc.Event.Target
				if n <= 0 || victim == uuid.Nil {
					return nil
				}
				return TargetCardInGraveyard("up to that many target cards from that player's graveyard", OwnedBy(victim)).WithCount(0, n)
			},
			Effect: froghemothExile,
		}},
	})
}

const froghemothLabel = "Froghemoth — exile up to that many target cards from their graveyard; +1/+1 counter per creature card, 1 life per noncreature card"

var froghemothExiledCard = IfItWasTail("if-it-was/froghemoth-counter-or-life", func(ctx *Context) error {
	return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}, func(ctx *Context) error {
	return GainLife{Amount: 1}.Apply(ctx)
})

func froghemothExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ExileThenIfItWas{Target: t.ID, Was: WasCreatureCard, Keyed: froghemothExiledCard}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
