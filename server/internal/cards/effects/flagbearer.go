package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// FlagbearerTargetRule is "While an opponent is choosing targets as part
// of casting a spell they control or activating an ability they control,
// that player must choose at least one Flagbearer on the battlefield if
// able" (Coalition Flag, Standard Bearer; #6236, CR 601.2c): a layer-6
// self static that marks the permanent with the rule
// (game.Characteristic.FlagbearerTargetRule). The engine binds each
// opponent of its controller at the cast and activation gates and in the
// legal-move enumerator (game/flagbearer.go). Coalition Flag's "enchanted
// creature is a Flagbearer" is an ordinary layer-4 subtype grant.
func FlagbearerTargetRule() game.StaticAbility {
	return game.StaticAbility{
		Label: "While an opponent is choosing targets as part of casting a spell they control or activating an ability they control, that player must choose at least one Flagbearer on the battlefield if able.",
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.FlagbearerTargetRule = true
		},
	}
}
