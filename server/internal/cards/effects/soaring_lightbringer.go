package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soaring Lightbringer — Enchantment Creature — Bird Glimmer {4}{W}, 4/5
// (Duskmourn Commander):
//
//	"Flying
//	 Other enchantment creatures you control have flying.
//	 Whenever you attack a player, create a 1/1 white Glimmer enchantment
//	 creature token that's tapped and attacking that player."
//
// The grant reads effective types (an animated enchantment counts). The
// trigger is once per player attacked in a declaration (Long-Range
// Sensor's shape; an attack at a planeswalker or battle is no attack at a
// player, CR 508.3e). The Glimmer enters tapped and attacking that player
// without being declared (CR 508.4); being an enchantment creature it has
// flying from this card.
//
// No simplification.
func init() {
	otherEnchantmentCreatures := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return OtherCreaturesYouControl(target, g, source) && target.IsEnchantment()
	}
	Register(Spec{
		OracleID:        "65ea7eb1-751b-4ecd-83cb-e9b6015ec194",
		Name:            "Soaring Lightbringer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(otherEnchantmentCreatures, "flying"), "Other enchantment creatures you control have flying."),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Soaring Lightbringer — create a 1/1 white Glimmer enchantment creature token tapped and attacking that player",
				func(g *game.Game, item *game.StackItem) error {
					tmpl := NewContext(g, item).TextToken(TokenCard("1/1 white Glimmer enchantment"))
					tmpl.Tapped = true
					tmpl.AttackingTarget = triggerAttackTarget(item)
					return CreateToken{Template: tmpl, N: 1}.Apply(NewContext(g, item))
				})),
		},
	})
}
