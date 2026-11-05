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
// The body is TokenCopyWithHasteAndGrant. Haste and the quoted trigger are gained as
// part of the copying process (CR 707.9a), so both are copiable values of
// the token: haste on its printed Keywords, the trigger as a named bundle
// on its GrantedAbilities (SacrificeThisTokenAtEndStepGrant). A copy of the
// token has the trigger too and sacrifices itself, and the trigger is the
// token's own, so it fires at the beginning of every end step it sees, not
// only the next one. Unlike Reflection of Kiki-Jiki's "sacrifice it at the
// beginning of the next end step", this is not a CR 603.7 delayed trigger.
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
		Grants:           []AbilityGrant{SacrificeThisTokenAtEndStepGrant(electroduplicateGrant, "Electroduplicate — sacrifice this token")},
		OnResolve:        TokenCopyWithHasteAndGrant(electroduplicateGrant),
	})
}

// electroduplicateGrant is the key of the granted sacrifice trigger, an
// on-disk identity.
const electroduplicateGrant = "electroduplicate/sacrifice-this-token"
