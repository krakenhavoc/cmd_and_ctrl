package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Labyrinth Raptor — Creature — Nightmare Dinosaur {B}{R}, 2/2 (Ikoria:
// Lair of Behemoths, #5512):
//
//	"Menace
//	 Whenever a creature you control with menace becomes blocked, defending
//	 player sacrifices a creature of their choice that's blocking it.
//	 {B}{R}: Creatures you control with menace get +1/+0 until end of
//	 turn."
//
// The trigger is Berserk Murlodont's EventBecomesBlocked, one per blocked
// attacker, and asks whether the creature has menace through the layers at
// that moment (CR 603.2). At resolution the defending player (the one the
// blocked creature attacks) picks one of THEIR creatures that is blocking
// it, through the sacrifice prompt; a blocker that has gone is not offered,
// and with none left nothing is sacrificed (CR 701.21a). The pump is
// "creatures you control with menace" read once at resolution (CR 611.2c).
//
// No simplification.
func init() {
	hasMenace := effects.And(effects.Creature(), effects.YouControl(), effects.HasKeyword("menace"))
	effects.Register(effects.Spec{
		OracleID:        "a29b2e93-d872-4be8-8b15-57f57282ed45",
		Name:            "Labyrinth Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventBecomesBlocked {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.IsCreature() && c.Controller == source.Controller && game.HasKeyword(&c, "menace")
			}, "Labyrinth Raptor — defending player sacrifices a creature of their choice that's blocking it",
				labyrinthRaptorSacrifice),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{R}: Creatures you control with menace get +1/+0 until end of turn.",
			Cost:    effects.ManaCost("{B}{R}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: hasMenace, Power: 1, Label: "Labyrinth Raptor — +1/+0"}.Apply(ctx)
			})),
		}},
	})
}

// labyrinthRaptorSacrifice makes the defending player sacrifice one of
// their creatures that is blocking the blocked creature.
func labyrinthRaptorSacrifice(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	attacker := item.Trigger.Event.Target
	defender := g.DefendingPlayerForAttackerForEffect(attacker)
	var candidates []uuid.UUID
	for _, id := range effects.BlockersOf(g, attacker) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsCreature() {
			continue
		}
		if defender == uuid.Nil {
			defender = c.Controller
		}
		if c.Controller == defender {
			candidates = append(candidates, id)
		}
	}
	if defender == uuid.Nil {
		return nil
	}
	return effects.SacrificeChoice{
		Player:     defender,
		Candidates: candidates,
		Question:   "Labyrinth Raptor — sacrifice a creature that's blocking it",
	}.Apply(ctx)
}
