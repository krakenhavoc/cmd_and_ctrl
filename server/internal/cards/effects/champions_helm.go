package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Champion's Helm — Artifact — Equipment for {3} (EDHREC rank 748):
//
//	"Equipped creature gets +2/+2.
//	 As long as equipped creature is legendary, it has hexproof.
//	 Equip {1}"
//
// The commander-protection Equipment, and the first attachment static
// in the catalog with a CONDITION on it rather than a flat grant.
// That is the whole reason it is in this batch: GrantToAttachedWhile
// re-reads its condition on every layer recompute, so the hexproof
// appears and disappears as the host's legendary status changes,
// with nothing captured at equip time.
//
// The condition is read off the host card through Card.IsLegendary,
// which parses the effective type line's supertypes. A creature that
// is made legendary while wearing the Helm gains hexproof; a token
// copy that "isn't legendary" does not. The pump is unconditional and
// applies to any creature, which is exactly the split the card
// prints.
//
// Hexproof is honoured by the S23 targeting gate, and it is the
// asymmetric one of the pair: it stops your opponents' removal and
// leaves your own Auras and pump spells legal — which is why the Helm
// is a better commander shield than Whispersilk Cloak's shroud.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c01aafb8-3da5-4eb1-8731-2a223e747d63",
		Name:         "Champion's Helm",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttachedWhile(
				func(host *game.Card, _ *game.Game, _ *game.Card) bool {
					return host.IsLegendary()
				},
				"hexproof",
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}
