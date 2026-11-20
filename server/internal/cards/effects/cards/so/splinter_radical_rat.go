package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splinter, Radical Rat — Legendary Creature — Mutant Ninja Rat
// {1}{W/B}{W/B}, 2/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"If a triggered ability of a Ninja creature you control triggers,
//	 that ability triggers an additional time.
//	 {1}{U}: Target Ninja can't be blocked this turn."
//
// The first line is Delney, Streetwise Lookout's trigger doubler (#752)
// with a Ninja filter: the source must be a Ninja creature, read from
// its characteristics when the ability triggered, so a dies trigger is
// judged by what it last was on the battlefield (CR 603.10a). The
// default controller check keeps it to creatures you control. Splinter
// is a Ninja and would double his own triggers, but he has none.
//
// The activated ability is Access Tunnel's body
// (targetCantBeBlockedThisTurn): a "can't be blocked" rule on the
// target until end of turn. "Target Ninja" is any Ninja on the
// battlefield, whoever controls it.
//
// No simplification.
func init() {
	doubler := effects.DoublesAbilitiesOf(effects.And(effects.Creature(), effects.Subtype("Ninja")), effects.DoublesAbilitiesOfOptions{})
	doubler.Label = "Splinter, Radical Rat"
	doubler.RowLabel = "Triggered abilities of Ninja creatures you control trigger an additional time."
	effects.Register(effects.Spec{
		OracleID:        "fb2cd508-f5cd-4d5b-b8e7-3cae1f39fd58",
		Name:            "Splinter, Radical Rat",
		Completeness:    effects.CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Target Ninja can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetPermanent("target Ninja", effects.Subtype("Ninja")),
			Effect:  effects.TargetCantBeBlockedThisTurn("Splinter, Radical Rat"),
		}},
	})
}
