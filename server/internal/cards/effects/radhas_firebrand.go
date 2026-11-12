package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Radha's Firebrand — Creature — Human Warrior {1}{R}, 3/1:
//
//	"Whenever this creature attacks, target creature defending player
//	 controls with power less than this creature's power can't block this
//	 turn.
//	 Domain — {5}{R}: This creature gets +2/+2 until end of turn. This
//	 ability costs {1} less to activate for each basic land type among
//	 lands you control. Activate only once each turn."
//
// The attack trigger is Ahn-Crop Crasher's can't-block body on a target
// read off this attack's defending player (TargetCreatureDefendingPlayerControls),
// narrowed by the source-relative "lesser power" comparison, which the
// engine checks at announce and again at resolution against the
// attacker's effective power (CR 608.2b). The pump's price is the shared
// domain count (DomainCostReduction, #5666); its limit is the
// announcement tally (OncePerTurnActivation, CR 602.1b).
//
// No simplification.
func init() {
	const trig = "Radha's Firebrand — target creature defending player controls with power less than this creature's power can't block this turn"
	const pump = "Domain — {5}{R}: This creature gets +2/+2 until end of turn. This ability costs {1} less to activate for each basic land type among lands you control. Activate only once each turn."
	attack := WheneverThisAttacks(trig, targetCreatureCantBlockThisTurn(trig))
	attack.TargetsFrom = func(ctx game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		spec := TargetCreatureDefendingPlayerControls(ctx, source, g)
		if spec == nil {
			return nil
		}
		spec.Label = "target creature defending player controls with power less than this creature's power"
		return RelativeToSource(spec, LesserPower())
	}
	Register(Spec{
		OracleID:     "0d6d5509-41e5-4099-98a0-18ae7531df7d",
		Name:         "Radha's Firebrand",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{attack},
		Activated: []ActivatedAbility{{
			Label:     pump,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:      ManaCost("{5}{R}"),
			Condition: OncePerTurnActivation(pump),
			CostModifiers: []game.CostModifier{
				DomainCostReduction("This ability costs {1} less to activate for each basic land type among lands you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: pump}.Apply(NewContext(g, item))
			},
		}},
	})
}
