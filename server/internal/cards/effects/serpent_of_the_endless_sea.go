package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serpent of the Endless Sea — Creature — Serpent {4}{U}, */*:
//
//	"Serpent of the Endless Sea's power and toughness are each equal to
//	 the number of Islands you control.
//	 This creature can't attack unless defending player controls an
//	 Island."
//
// The size is a characteristic-defining ability (CR 604.3) in layer 7a,
// Ulvenwald Hydra's mechanism counting Islands rather than lands. Islands
// are read by their effective subtype, so a land an effect made an Island
// counts. With no Islands it is a 0/0 and dies to the state-based action.
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c), with the defending
// player worked out per target (CR 508.5, 508.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b954d5f-3a93-4dd9-9d60-6097594d449c",
		Name:         "Serpent of the Endless Sea",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7A_CDA,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := b08LandsWithSubtypeControlled(g, source.Controller, "Island")
					c.Power = n
					c.Toughness = n
				},
			},
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
	})
}
