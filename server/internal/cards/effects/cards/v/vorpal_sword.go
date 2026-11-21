package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vorpal Sword — Artifact — Equipment {B} (Adventures in the Forgotten
// Realms Commander):
//
//	"Equipped creature gets +2/+0 and has deathtouch.
//	 {5}{B}{B}{B}: Until end of turn, this Equipment gains "Whenever equipped
//	 creature deals combat damage to a player, that player loses the game."
//	 Equip {B}{B}"
//
// The granted trigger is a bundle (ADR 0093) the activated ability gives the
// Equipment itself until end of turn (CR 611.2a): a layer-6 grant on that
// permanent that ends in the cleanup step, and a new object (the Sword left
// and came back) has nothing. It is the Equipment's trigger, so it watches
// the creature the Sword is attached to when the damage is dealt, and the
// player who was dealt the combat damage loses the game (CR 104.3e).
//
// No simplification.
const vorpalSwordGrant = "vorpal-sword/loses-the-game"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "61bfab3f-d8f0-4ff5-9ddf-3fd605070c8f",
		Name:         "Vorpal Sword",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("deathtouch"),
		},
		Grants: []effects.AbilityGrant{{
			Key: vorpalSwordGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventDealDamage,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
					},
					"Vorpal Sword — that player loses the game",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil {
							return nil
						}
						return effects.LoseTheGame{Player: item.Trigger.Event.Target}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "Whenever equipped creature deals combat damage to a player, that player loses the game.",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label: "{5}{B}{B}{B}: Until end of turn, this Equipment gains \"Whenever equipped creature deals combat damage to a player, that player loses the game.\"",
				Cost:  effects.ManaCost("{5}{B}{B}{B}"),
				// Gives a trigger to the Sword; it answers no spell or ability.
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target:   item.SourceCardID,
						Mods:     []game.Mod{game.GrantAbilitiesMod(vorpalSwordGrant)},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Vorpal Sword — gains \"that player loses the game\" until end of turn",
					}.Apply(ctx)
				},
			},
			effects.EquipAbility("{B}{B}"),
		},
	})
}
