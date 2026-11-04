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
// The shared body is TokenCopyWithHasteSacrificedAtEndStep: haste and the
// quoted trigger are gained as part of the copying process, so both are
// copiable values of the token (CR 707.9a). The token is sacrificed by its
// own trigger, a copy of the token has the trigger too, and a token that
// loses its abilities stays (#4950). "The end step" is each end step, not
// only the next one.
//
// The copy runs the copied card's entry too (#3172): the token goes through
// the ordinary battlefield-entry pipeline (#762), so its triggers fire and
// its "as enters" clause is asked.
//
// No simplification.
const electroduplicateGrant = "electroduplicate/sacrifice-this-token"

func init() {
	Register(Spec{
		OracleID:         "112f2b3f-32e7-40b1-b80e-0b99184a840c",
		Name:             "Electroduplicate",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}{R}")},
		Targets:          TargetCreature("target creature you control", YouControl()),
		OnResolve:        TokenCopyWithHasteSacrificedAtEndStep(electroduplicateGrant),
		Grants:           []AbilityGrant{SacrificeThisTokenAtEndStepGrant(electroduplicateGrant, "Electroduplicate — sacrifice this token")},
	})
}
