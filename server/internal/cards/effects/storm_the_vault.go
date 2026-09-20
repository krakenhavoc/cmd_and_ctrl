package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm the Vault — "Whenever one or more creatures you control deal
// combat damage to a player, create a Treasure token. // At the
// beginning of your end step, if you control five or more artifacts,
// transform Storm the Vault."
//
// The front half of the simplest possible proof of ADR 0079's in-place
// transform verb: a Legendary Enchantment that becomes a Legendary
// Land without going anywhere. Its back face is vault_of_catlacan.go,
// registered under "<oracle_id>#1".
//
// Two things about the trigger pair are easy to get wrong and are the
// reason this card was picked rather than a shorter one:
//
//   - "ONE OR MORE creatures" is a batch clause (CR 603.2c). Written
//     per-creature it would make one Treasure per attacker, which is
//     STRONGER than printed — a three-creature alpha strike would
//     produce three. WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer
//     is OncePerBatchPerPlayer, which is both halves of the rule: one
//     Treasure per player damaged, and one per combat-damage batch.
//   - The end-step trigger has an intervening if (CR 603.4), so the
//     artifact count is checked TWICE — once when the step begins, to
//     decide whether the ability triggers at all, and again as it
//     resolves. A Treasure sacrificed in response to the trigger stops
//     the transform, and that is the printed card.
//
// No simplification. The transform itself is the ADR 0079 verb: the
// permanent keeps its place, its timestamp and anything attached to
// it, and becomes a land in the same instant.
func init() {
	Register(Spec{
		OracleID:     stormTheVaultOracleID,
		Name:         "Storm the Vault",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Storm the Vault — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
			On(game.EventBeginEndStep, stormTheVaultHasFiveArtifacts,
				"Storm the Vault — transform it",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4's second check. The first one lives in
					// the AppliesTo above; this is the one that
					// notices a Treasure spent in response.
					if countControlled(g, item.Controller, MatchArtifact) < 5 {
						return nil
					}
					return TransformThis{}.Apply(NewContext(g, item))
				}),
		},
	})
}

// stormTheVaultHasFiveArtifacts is the intervening-if's first check
// (CR 603.4), asked as the end step begins: the ability does not
// trigger at all unless its controller already has the artifacts.
func stormTheVaultHasFiveArtifacts(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor == source.Controller &&
		countControlled(g, source.Controller, MatchArtifact) >= 5
}

// stormTheVaultOracleID is shared with the back face, which registers
// under it plus "#1" (game.CatalogKey). One constant so the two files
// cannot drift apart — the same shape invasion_of_karsus.go and
// refraction_elemental.go use.
const stormTheVaultOracleID = "72205fac-a94a-45cc-94c6-40ece2fdce0e"
