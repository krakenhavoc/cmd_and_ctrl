package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cyclonus, the Saboteur // Cyclonus, Cybertronian Fighter (#3195, ADR
// 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {2}{U}{B}, 2/5:
//
//	"More Than Meets the Eye {5}{U}{B}
//	 Flying
//	 Whenever Cyclonus deals combat damage to a player, it connives.
//	 Then if Cyclonus's power is 5 or greater, convert it."
//
// Back face, Legendary Artifact — Vehicle, 5/5:
//
//	"Living metal
//	 Flying
//	 Whenever Cyclonus deals combat damage to a player, convert it. If
//	 you do, there is an additional beginning phase after this phase."
//
// Connive is draw, discard, and a +1/+1 counter for a nonland discard
// (CR 701.50a). The discard is a prompt, so "Then if Cyclonus's power is
// 5 or greater" runs from the discard's own continuation, after the
// counter lands and with the layers current. A Cyclonus that has left
// the battlefield still draws and discards (CR 701.50c) and then has
// nothing to convert. The back face's "if you do" is judged on the
// permanent actually turning over (CR 701.28e): a Cyclonus that already
// turned over since the trigger went on the stack adds no phase.
//
// No simplification.
const cyclonusOracleID = "7ead01b9-def3-4f9d-a860-af64e94cb45e"

func init() {
	Register(Spec{
		OracleID:         cyclonusOracleID,
		Name:             "Cyclonus, the Saboteur",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{5}{U}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Cyclonus, the Saboteur — it connives, then convert it if its power is 5 or greater",
				cyclonusConnive),
		},
	})
	Register(Spec{
		OracleID:        cyclonusOracleID + "#1",
		Name:            "Cyclonus, Cybertronian Fighter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Cyclonus, Cybertronian Fighter — convert it; if you do, an additional beginning phase",
				cyclonusConvertForPhase),
		},
	})
}

func cyclonusConnive(g *game.Game, item *game.StackItem) error {
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(NewContext(g, item)); err != nil {
		return err
	}
	source := item.SourceCardID
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   source,
		N:        1,
		Question: "Cyclonus, the Saboteur — connive: discard a card",
		Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
			if !onBattlefield(g, source) {
				return nil
			}
			if len(discarded) > 0 {
				if card, ok := g.LookupCardForEffect(discarded[0]); ok && !card.IsLand() {
					if err := g.AddCounterForEffect(source, game.CounterPlusOne, 1); err != nil {
						return err
					}
				}
			}
			g.RecomputeLayersIfStaleLocked()
			cyc, ok := g.LookupCardForEffect(source)
			if !ok || cyc.CurrentPower() < 5 {
				return nil
			}
			return ConvertThis{}.Apply(NewContext(g, item))
		},
	})
	return nil
}

func cyclonusConvertForPhase(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	converted, err := convertThisReporting(ctx)
	if err != nil || !converted {
		return err
	}
	return AddPhases{
		Anchor: game.PhaseAnchor{Kind: game.AnchorThisPhase},
		Kinds:  []game.PhaseKind{game.PhaseKindBeginning},
	}.Apply(ctx)
}
