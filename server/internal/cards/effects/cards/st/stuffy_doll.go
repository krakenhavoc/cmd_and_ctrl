package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stuffy Doll — Artifact Creature — Construct {5}, 0/1 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Indestructible
//	 As this creature enters, choose a player.
//	 Whenever this creature is dealt damage, it deals that much damage to the
//	 chosen player.
//	 {T}: This creature deals 1 damage to itself."
//
// Saskia's chosen player (CR 614.12: stored on the Doll as it enters) with
// Brash Taunter's reflection of the damage the creature was dealt. The player
// is read as the ability triggers, so a Doll that has left still hits the
// player chosen for it; one who has left the game is skipped. Damage dealt
// together (several blockers) is one trigger for the total. The tap ability
// damages the Doll itself, and indestructibility keeps it on the battlefield.
//
// No simplification.
func init() {
	const label = "Stuffy Doll — it deals that much damage to the chosen player"
	trigger := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return effects.B35SelfWasDealtDamage(ev, source)
		},
		Key:        label,
		AtBatchEnd: effects.ThatMuchDamage(),
		Effect:     stuffyDollReflect,
	}
	trigger.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Player = effects.ChosenPlayerOf(g, source.InstanceID)
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "499dceab-1890-49c8-b35d-b059a1dc950f",
		Name:            "Stuffy Doll",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		AsEnters:        effects.ChoosePlayerAsEnters("Stuffy Doll", effects.Players),
		Triggered:       []game.TriggeredAbility{trigger},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to itself.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{Source: item.SourceCardID, Target: item.SourceCardID, Amount: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

func stuffyDollReflect(g *game.Game, item *game.StackItem) error {
	chosen := item.Params.Player
	if chosen == uuid.Nil || item.Trigger == nil {
		return nil
	}
	if p := g.PlayerByIDForEffect(chosen); p == nil || p.Eliminated {
		return nil
	}
	ev := item.Trigger.Event
	return effects.DealDamage{Source: ev.Target, Target: chosen, Amount: ev.Amount}.Apply(effects.NewContext(g, item))
}
