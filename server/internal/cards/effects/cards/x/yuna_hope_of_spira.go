package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Yuna, Hope of Spira — Legendary Creature — Human Cleric {3}{G}{W}, 3/5
// (Final Fantasy, #4462):
//
//	"During your turn, Yuna and enchantment creatures you control have
//	 trample, lifelink, and ward {2}.
//	 At the beginning of your end step, return up to one target enchantment
//	 card from your graveyard to the battlefield with a finality counter on
//	 it."
//
// The first ability is three layer-6 grants over one predicate (Yuna, and
// each enchantment creature you control, read through the layers) that
// hold only while it is your turn, and the ward as the granted trigger
// over the same predicate (WardGranted, Hexing Squelcher's shape, ward
// {2} the pay-unless prompt). On an opponent's turn nothing is granted
// and the ward does not trigger. The end-step ability's target is chosen
// on the stack ("up to one" may choose none); the card returns with the
// finality counter as it enters (CR 614.1c, Grim Reaper's Scythe's
// return), so it is exiled instead of going to a graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3dee768a-612c-4960-bad2-2fe5b83a70c3",
		Name:         "Yuna, Hope of Spira",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(yunaGrants, "trample"),
			effects.KeywordGrant(yunaGrants, "lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{2}"), "Yuna, Hope of Spira — during your turn, Yuna and enchantment creatures you control have ward {2}", yunaGrants),
			effects.Targeting(effects.AtYourEndStep("Yuna, Hope of Spira — return up to one target enchantment card from your graveyard to the battlefield with a finality counter on it",
				effects.GrimReapersScytheReturn),
				effects.TargetCardInGraveyard("up to one target enchantment card from your graveyard", effects.YouOwn(), effects.Enchantment()).WithCount(0, 1)),
		},
	})
}

// yunaGrants is "During your turn, Yuna and enchantment creatures you
// control": Yuna herself, or an enchantment creature her controller
// controls, while her controller is the active player.
func yunaGrants(target *game.Card, g *game.Game, source *game.Card) bool {
	if !effects.IsActivePlayer(g, source.Controller) {
		return false
	}
	if target.InstanceID == source.InstanceID {
		return true
	}
	return target.Controller == source.Controller && target.IsCreature() && target.HasCardType("enchantment")
}
