package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Legion Loyalty — Enchantment {6}{W}{W} (EDHREC rank 2905):
//
//	"Creatures you control have myriad. (Whenever a creature with
//	 myriad attacks, for each opponent other than defending player,
//	 you may create a token copy that's tapped and attacking that
//	 player or a planeswalker they control. Exile the tokens at end
//	 of combat.)"
//
// Every attacker becomes one attacker per opponent. Myriad (CR
// 702.116) is a triggered ability the keyword grants, so the grant
// is granted to each creature through layer 6 and fires when that
// creature is declared as an attacker. Each separate grant is another
// instance of myriad.
// The shared granted-myriad body asks per opponent, allows planeswalker
// attacks, and creates the copies simultaneously after count replacements.
func init() {
	Register(Spec{
		OracleID:     "24e39d26-6a2b-461e-b194-bda575bcf898",
		Name:         "Legion Loyalty",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{{Key: "legion-loyalty/myriad", Triggered: []game.TriggeredAbility{GrantedMyriad("Legion Loyalty")}, Text: "Myriad"}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && target.IsCreature()
		}, "legion-loyalty/myriad"), "Creatures you control have myriad.")},
	})
}
