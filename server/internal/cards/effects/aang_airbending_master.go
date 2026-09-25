package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aang, Airbending Master — Legendary Creature — Human Avatar Ally
// {4}{W}, 4/4:
//
//	"When Aang enters, airbend another target creature. (Exile it.
//	 While it's exiled, its owner may cast it for {2} rather than its
//	 mana cost.)
//	 Whenever one or more creatures you control leave the battlefield
//	 without dying, you get an experience counter.
//	 At the beginning of your upkeep, create a 1/1 white Ally creature
//	 token for each experience counter you have."
//
// Three abilities, each an existing shape:
//
//   - The enters trigger is MANDATORY and says "another", unlike Aang,
//     the Last Airbender's optional "up to one other". So the clause is
//     built per trigger with TargetsFrom, which is handed the source
//     and can exclude it by instance (NotSelf) — Aang is never offered
//     as his own target, and with no other creature on the battlefield
//     the trigger is removed outright (CR 603.3d) instead of going on
//     the stack aimed at himself. Airbend is the shared primitive
//     (airbend.go).
//   - "Leave the battlefield without dying" is CR 700.4 read off the
//     LTB event's destination — Dour Port-Mage's condition without the
//     word "other": Aang counts himself, so a bounced or flickered Aang
//     gets a counter too, off the CR 603.10a look-back the harvester
//     already does for a source that left. "One or more" is
//     OncePerBatch (#587 / #829): a mass bounce of your board is one
//     counter, and a second, later batch is another (CR 603.2c).
//     Airbending your own creature with the enters trigger is a
//     departure without dying, and counts — as printed.
//   - "You get an experience counter" is a player counter (CR 122.1),
//     placed through the CR 614 window with the controller named as
//     the placer, so Lae'zel and Vorinclex see "you" putting it.
//   - The upkeep count is read at RESOLUTION, so a counter gained in
//     response to the trigger is included.
//
// No simplification.
func init() {
	const (
		etbLabel    = "Aang, Airbending Master — airbend another target creature"
		expLabel    = "Aang, Airbending Master — you get an experience counter"
		upkeepLabel = "Aang, Airbending Master — create a 1/1 Ally for each experience counter"
	)
	Register(Spec{
		OracleID:     "f3176779-74f7-4136-a701-438feeead7a0",
		Name:         "Aang, Airbending Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       etbLabel,
				TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
					return TargetCreature("another target creature", NotSelf(source.InstanceID))
				},
				Effect: AirbendOtherTarget,
			},
			OncePerBatch(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureYouControlLeftWithoutDying(ev, source, g)
			}, expLabel, youGetAnExperienceCounter)),
			AtYourUpkeep(upkeepLabel, func(g *game.Game, item *game.StackItem) error {
				n := experienceCounters(g, item.Controller)
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Ally"), N: n}.
					Apply(NewContext(g, item))
			}),
		},
	})
}

// youGetAnExperienceCounter is "you get an experience counter": one
// experience counter on the resolving item's controller, placed by
// that player (CR 122.1, ADR 0056 Decision 5).
//
// Caller holds g.mu.
func youGetAnExperienceCounter(g *game.Game, item *game.StackItem) error {
	return g.AddPlayerCounterByForEffect(item.Controller, item.Controller, game.CounterExperience, 1)
}

// experienceCounters is how many experience counters a player has
// right now — zero for a player who has none or has left the game.
func experienceCounters(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return 0
	}
	return p.Counters[game.CounterExperience]
}
