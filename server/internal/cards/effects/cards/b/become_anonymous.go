package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Become Anonymous — Instant {2}{U}{U}:
//
//	"Exile target nontoken creature you own and the top two cards of
//	 your library in a face-down pile, shuffle that pile, then cloak
//	 those cards. They enter tapped."
//
// The target is exiled first, then the top two cards of the controller's
// library are exiled face down, each by a keyed step so a replacement
// prompt in the middle resumes the rest. The pile (the target and the two
// cards) is shuffled and cloaked one at a time, each entering tapped
// (Game.CloakShuffledTappedFromExileForEffect). A target that is no longer
// legal makes the whole spell do nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bac0fcee-9c1a-46b7-86c8-ffbfcc1e96de",
		Name:         "Become Anonymous",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target nontoken creature you own", effects.YouOwn(), effects.Not(effects.IsTokenPredicate())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.ExileCardKeyedForEffect(id, ctx.Tail(becomeAnonymousTargetExiled))
		},
	})
}

// becomeAnonymousTargetExiled exiles the top two cards face down once the
// target is in exile, carrying the target on. Keyed: an on-disk identity,
// never renamed or reused.
var becomeAnonymousTargetExiled = effects.CardTail("then/become-anonymous-target-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	next := ctx.Tail(becomeAnonymousCloak)
	next.Carry = append([]uuid.UUID(nil), o.IDs...)
	return ctx.Game.ExileTopFaceDownKeyedForEffect(ctx.Controller(), 2, next)
})

// becomeAnonymousCloak shuffles the target and the two cards exiled face
// down together and cloaks them tapped. Keyed: an on-disk identity, never
// renamed or reused.
var becomeAnonymousCloak = effects.CardTail("then/become-anonymous-cloak", func(ctx *effects.Context, o game.TailOutcome) error {
	pile := append(append([]uuid.UUID(nil), o.Tail.Carry...), o.IDs...)
	if len(pile) == 0 {
		return nil
	}
	_, err := ctx.Game.CloakShuffledTappedFromExileForEffect(ctx.Controller(), pile)
	return err
})
