package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 set NEO, batch 8 (#5493): Kamigawa: Neon Dynasty's "if you control an
// artifact and an enchantment" payoffs and its Samurai tokens.
//
// The clause has no rules meaning of its own: it asks for one artifact and one
// enchantment among the permanents the player controls, read through the layers
// (an Enchantment Artifact, a Shrine artifact, satisfies both halves).

// neoB08YouControlAnArtifactAndAnEnchantment is the clause for the player.
func NeoB08YouControlAnArtifactAndAnEnchantment(g *game.Game, player uuid.UUID) bool {
	var artifact, enchantment bool
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player {
			continue
		}
		for _, t := range c.Effective().Types {
			switch t {
			case "Artifact":
				artifact = true
			case "Enchantment":
				enchantment = true
			}
		}
		if artifact && enchantment {
			return true
		}
	}
	return false
}

// neoB08CondArtifactAndEnchantment is the intervening "if" of the clause,
// checked as the ability resolves (CR 603.4) for the ability's controller.
var NeoB08CondArtifactAndEnchantment = CondBoard{
	Name: "you control an artifact and an enchantment",
	Test: func(ctx *Context) bool {
		return NeoB08YouControlAnArtifactAndAnEnchantment(ctx.Game, ctx.Controller())
	},
}

// neoB08WhenArtifactAndEnchantment is the same clause as the event happens,
// read for the source's controller, for composing with AllOf (CR 603.4).
func NeoB08WhenArtifactAndEnchantment(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return NeoB08YouControlAnArtifactAndAnEnchantment(g, source.Controller)
}

// neoB08ModeCondition is Soul Transfer's "If you control an artifact and an
// enchantment as you cast this spell, you may choose both instead", read for
// the chooser as the modes are chosen (CR 601.2b).
var NeoB08ModeCondition = game.ModeCondition("you-control-an-artifact-and-an-enchantment", NeoB08YouControlAnArtifactAndAnEnchantment)

// neoB08SamuraiToken is the 2/2 white Samurai creature token with vigilance.
func NeoB08SamuraiToken() game.Card { return TokenCard("2/2 white Samurai with vigilance") }

// neoB08ReturnTheModesTargetToHand (shared with Dr. Leonard "Bones" McCoy) is "return target … card from your
// graveyard to your hand" as a modal bullet's body (Soul Transfer).
func NeoB08ReturnTheModesTargetToHand(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
}
