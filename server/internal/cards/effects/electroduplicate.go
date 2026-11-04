package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Electroduplicate — Sorcery {2}{R}:
//
//	"Create a token that's a copy of target creature you control,
//	 except it has haste and 'At the beginning of the end step,
//	 sacrifice this token.'
//	 Flashback {2}{R}{R} (You may cast this card from your graveyard
//	 for its flashback cost. Then exile it.)"
//
// Reflection of Kiki-Jiki's shape (haste copy, scheduled sacrifice) on
// a sorcery instead of an activated ability: haste rides the token's
// printed Keywords, since a token has no catalog key to hang a Layer 6
// grant off, and the sacrifice is a real CR 603.7 delayed trigger
// scheduled against the token's own instance ID — read back off
// EventTokenCreated, because CreateTokenCopy mints the instance and
// returns nothing. "The end step" rather than "the next end step" is
// the same StepEnd default ScheduleDelayedTrigger already uses.
//
// The copy runs the copied card's entry too (#3172): the token goes through
// the ordinary battlefield-entry pipeline (#762), so its triggers fire and
// its "as enters" clause is asked.
func init() {
	Register(Spec{
		OracleID:         "112f2b3f-32e7-40b1-b80e-0b99184a840c",
		Name:             "Electroduplicate",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}{R}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve:        TokenCopyWithHasteSacrificedAtEndStep("Electroduplicate — sacrifice the token"),
	})
}
