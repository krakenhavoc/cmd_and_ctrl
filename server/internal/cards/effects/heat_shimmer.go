package effects

// Heat Shimmer — Sorcery {2}{R} (Special Guests, S82 BNS):
//
//	"Create a token that's a copy of target creature, except it has haste
//	 and 'At the beginning of the end step, exile this token.'"
//
// Electroduplicate's body (TokenCopyWithHasteAndGrant) on any creature, with
// an exile in place of the sacrifice. Haste and the quoted trigger are part
// of the copying process (CR 707.9a), so both are copiable values of the
// token; the trigger is the token's own, not a delayed trigger of the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c80d39e-f3af-4d24-bca0-19f0bd0f5082",
		Name:         "Heat Shimmer",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants:       []AbilityGrant{ExileThisTokenAtEndStepGrant(heatShimmerGrant, "Heat Shimmer — exile this token")},
		OnResolve:    TokenCopyWithHasteAndGrant(heatShimmerGrant),
	})
}

// heatShimmerGrant is the key of the granted exile trigger, an on-disk identity.
const heatShimmerGrant = "heat-shimmer/exile-this-token"
