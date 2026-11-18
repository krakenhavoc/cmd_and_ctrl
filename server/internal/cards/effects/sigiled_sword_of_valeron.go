package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigiled Sword of Valeron — Artifact — Equipment {3}:
//
//	"Equipped creature gets +2/+0, has vigilance, and is a Knight in
//	 addition to its other types.
//	 Whenever equipped creature attacks, create a 2/2 white Knight
//	 creature token with vigilance that's attacking.
//	 Equip {3}"
//
// The token is put onto the battlefield attacking (CR 508.4): the
// attacker chooses which player or planeswalker it attacks, preferring
// what the equipped creature attacks. It never "attacked", so it
// triggers nothing that watches attack declarations.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1c4269a-1d99-4f3b-8107-a005df3468d5",
		Name:         "Sigiled Sword of Valeron",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("vigilance"),
			{
				Label:     "Equipped creature is a Knight in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					// "Knight" is a word of the Equipment's text (CR 612.2, #3984).
					knight := source.TextSubtype("Knight")
					if !eotHasType(c.Subtypes, knight) {
						c.Subtypes = append(c.Subtypes, knight)
					}
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Sigiled Sword of Valeron — create a 2/2 white Knight token with vigilance that's attacking",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: TokenCard("2/2 white Knight with vigilance"),
						N:        1,
						Prefer:   triggerAttackTarget(item),
						Label:    "Sigiled Sword of Valeron",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
