package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sokka and Suki — Legendary Creature — Human Warrior Ally {U}{R}{W},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever Sokka and Suki or another Ally you control enters, attach
//	 up to one target Equipment you control to that creature.
//	 Whenever an Equipment you control enters, create a 1/1 white Ally
//	 creature token."
//
// The first trigger's Equipment is chosen as it goes on the stack ("up
// to one", so none is a legal choice) and is attached to the creature
// that entered as it resolves — nothing happens if that creature has
// left the battlefield or the Equipment is no longer yours by then.
// The second is Barret, Avalanche Leader's Equipment trigger with the
// set's Ally token, and the token it makes is an Ally entering, so it
// triggers the first ability in turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e84becc0-6e15-4a74-aec0-9746b86ab5cc",
		Name:         "Sokka and Suki",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && (c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Ally")))
				}, "Sokka and Suki — attach up to one target Equipment you control to that creature",
					sokkaAndSukiAttachToTheEnteringCreature),
				TargetPermanent("up to one target Equipment you control", HasSubtype("Equipment"), YouControl()).WithCount(0, 1)),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.HasSubtype("Equipment")
			}, "Sokka and Suki — create a 1/1 white Ally creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Ally"), N: 1})),
		},
	})
}

// sokkaAndSukiAttachToTheEnteringCreature attaches the still-legal
// Equipment target to the creature whose entry triggered the ability.
func sokkaAndSukiAttachToTheEnteringCreature(g *game.Game, item *game.StackItem) error {
	creature := item.Trigger.Event.CardID
	if creature == uuid.Nil || !onBattlefield(g, creature) {
		return nil
	}
	if c, ok := g.LookupCardForEffect(creature); !ok || !c.IsCreature() {
		return nil
	}
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.AttachForEffect(t.ID, game.TargetRef{Kind: game.TargetCard, ID: creature})
	}
	return nil
}
