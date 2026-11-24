package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stardew Valley — Land (Secret Lair Drop, #5462):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Tap an untapped creature you control: Create a Food token.
//	 {3}, {T}: Choose target permanent you control. Draw a card, then
//	 another player of your choice may gain control of that permanent.
//	 Activate only as a sorcery."
//
// The Food ability is The Shire's (a tap-another cost). The gift ability
// targets a permanent you control; it draws, then the activator names another
// player (an opponent: no teams here) and THAT player is asked whether to
// take it, so a refusal leaves the permanent where it was. The control change
// is indefinite (CR 611.2a), through the layer-2 effect, and it is skipped if
// the permanent has left the battlefield or become a new object since it was
// chosen (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "6a4ee425-b3b8-487d-866c-9e2d73682466",
		Name:          "Stardew Valley",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.ColorlessTapRow()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Tap an untapped creature you control: Create a Food token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.TapAnotherUntapped("an untapped creature you control", effects.Creature())),
			Effect:  effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1}),
		}, {
			Label:        "{3}, {T}: Choose target permanent you control. Draw a card, then another player of your choice may gain control of that permanent. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target permanent you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				return effects.ChoosePlayer{
					Among:    effects.Opponents,
					Question: "Stardew Valley — choose another player to offer " + c.Name + " to",
					ThenKey:  stardewValleyChosen,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

// stardewValleyChosen asks the chosen player whether to take the permanent.
// An on-disk key: never renamed, never reused.
var stardewValleyChosen = effects.ChoosePlayerThen("option-pick/stardew-valley-player", func(ctx *effects.Context, chosen uuid.UUID, carry []uuid.UUID) error {
	if chosen == uuid.Nil || len(carry) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(carry[len(carry)-1])
	if !ok || !effects.OnBattlefield(ctx.Game, c.InstanceID) {
		return nil
	}
	return effects.MayChoice{
		Player:   chosen,
		Question: "Stardew Valley — gain control of " + c.Name + "?",
		YesLabel: "Gain control",
		NoLabel:  "Don't",
		Then:     stardewValleyTake,
		Objects:  []game.ObjectRef{{ID: c.InstanceID, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

// stardewValleyTake is the chosen player's "yes". An on-disk key: never
// renamed, never reused.
var stardewValleyTake = effects.MayYes("confirm/stardew-valley-gain-control", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Objects) == 0 {
		return nil
	}
	ref := r.Objects[0]
	c, ok := ctx.Game.LookupCardForEffect(ref.ID)
	if !ok || c.ObjectEpoch != ref.Epoch || !effects.OnBattlefield(ctx.Game, ref.ID) {
		return nil
	}
	return effects.GainControl{
		Target:     ref.ID,
		Controller: r.Chooser,
		Duration:   game.IndefiniteDuration(),
		Label:      "Stardew Valley — another player gains control of it",
	}.Apply(ctx)
})
