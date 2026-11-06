package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avatar Kyoshi, Earthbender — Legendary Creature — Human Avatar
// {5}{G}{G}{G}, 6/6 (Avatar: The Last Airbender Eternal, #5075):
//
//	"During your turn, Avatar Kyoshi has hexproof.
//	 At the beginning of combat on your turn, earthbend 8, then untap
//	 that land. (Target land you control becomes a 0/0 creature with
//	 haste that's still a land. Put eight +1/+1 counters on it. When it
//	 dies or is exiled, return it to the battlefield tapped.)"
//
// The hexproof is a layer 6 grant on Kyoshi herself that holds only while
// it is her controller's turn. The earthbend is the engine verb; the
// untap rides its Then, so it happens once the counters have landed, and
// only to the land the earthbend acted on.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd751ed4-4002-4cae-9d1d-1fe0f1fd4965",
		Name:         "Avatar Kyoshi, Earthbender",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && IsYourTurn(g, source.Controller)
		}, "hexproof")},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Avatar Kyoshi, Earthbender — earthbend 8, then untap that land",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					land := FirstLegalBattlefieldTarget(ctx)
					if land == uuid.Nil {
						return nil
					}
					then := ctx.Tail(avatarKyoshiEarthbenderUntap)
					then.Carry = []uuid.UUID{land}
					return Earthbend{Target: land, N: 8, Then: then}.Apply(ctx)
				}), EarthbendTargets()),
		},
	})
}

// avatarKyoshiEarthbenderUntap is "then untap that land" once the
// earthbend's counters have landed: Carry is the land. An on-disk key:
// never renamed, never reused.
var avatarKyoshiEarthbenderUntap = CardTail("avatar-kyoshi-earthbender/untap-that-land", func(ctx *Context, o game.TailOutcome) error {
	var lands []uuid.UUID
	for _, id := range o.Tail.Carry {
		if onBattlefield(ctx.Game, id) {
			lands = append(lands, id)
		}
	}
	return untapEach(ctx, lands)
})
