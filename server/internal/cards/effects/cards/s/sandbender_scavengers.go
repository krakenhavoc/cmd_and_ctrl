package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sandbender Scavengers — Creature — Human Rogue {W}{B}, 1/1 (Avatar: The
// Last Airbender, #4459):
//
//	"Whenever you sacrifice another permanent, put a +1/+1 counter on this
//	 creature.
//	 When this creature dies, you may exile it. When you do, return target
//	 creature card with mana value less than or equal to this creature's
//	 power from your graveyard to the battlefield."
//
// The counter trigger is Pirate Peddlers'. The dies trigger is Undead
// Butler's "you may exile it. When you do, …" (exileItFromTheGraveyardThen):
// the CR 603.12 reflexive trigger exists only once the Scavengers is
// really in exile, and its target is chosen as it goes on the stack, so
// the Scavengers can never return itself. The mana-value cap is the
// Scavengers' power as it last existed on the battlefield (the ruling,
// CR 608.2h), read before the exile and carried on the reflexive
// trigger's Params, from which the target clause is built.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0feacc7c-b93e-4322-bc8b-b3c19006128e",
		Name:         "Sandbender Scavengers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.AvB18YouSacrificedAnother,
				"Sandbender Scavengers — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.Optional(
				effects.WhenThisDies("Sandbender Scavengers — exile it", sandbenderScavengersExileIt),
				"Sandbender Scavengers — exile it to return a creature card from your graveyard to the battlefield?"),
		},
	})
}

// sandbenderScavengersExileIt reads the Scavengers' last-known power,
// then exiles it from the graveyard and, once it is in exile, creates the
// reflexive return carrying that power.
func sandbenderScavengersExileIt(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	power := 0
	if info, ok := ctx.SourcePermanent(); ok {
		power = info.Power
	}
	if err := (effects.ExileTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return effects.ReflexiveTrigger{
		Label:  "Sandbender Scavengers — return target creature card with mana value less than or equal to its power from your graveyard to the battlefield",
		Body:   sandbenderScavengersReturnBody,
		Params: game.EffectParams{Amount: power},
	}.Apply(ctx)
}

// sandbenderScavengersReturnBody is the reflexive half. An on-disk key.
var sandbenderScavengersReturnBody = game.ReflexiveBody("sandbender-scavengers/return-to-battlefield",
	effects.SimpleBody(effects.ReturnFirstLegalGraveyardTargetToBattlefield),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature card with mana value less than or equal to Sandbender Scavengers' power in your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(p.Amount))
	})
