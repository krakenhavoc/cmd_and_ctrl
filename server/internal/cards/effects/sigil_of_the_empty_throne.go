package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigil of the Empty Throne — Enchantment {3}{W}{W} (EDHREC rank
// 1364):
//
//	"Whenever you cast an enchantment spell, create a 4/4 white Angel
//	 creature token with flying."
//
// The enchantress deck's win condition: every enchantment cast after
// it is a Serra Angel without vigilance. Cast, not resolve — the
// Angel comes even if the enchantment is countered — and the spell's
// type is read off the stack where its type line is intact. The
// Sigil does not trigger on itself (it is not on the battlefield
// when it is cast).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b9b5b22-5a7d-4e37-a870-ca0f0efa4f36",
		Name:         "Sigil of the Empty Throne",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b12EnchantmentSpellCastByYou(ev, source, g)
			}, "Sigil of the Empty Throne — a 4/4 Angel with flying", Do(CreateToken{Template: b12WhiteAngelFlyingToken(), N: 1})),
		},
	})
}
