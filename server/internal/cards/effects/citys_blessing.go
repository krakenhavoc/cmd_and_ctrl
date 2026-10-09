package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// citys_blessing.go — the card vocabulary for ascend and the city's
// blessing (CR 702.131, #2696, ADR 0096's 2026-10-08 amendment).
//
// The designation lives on the player (game/citys_blessing.go) and the
// engine grants it, so a card file declares only what it does WITH it:
//
//	PrintedKeywords: []string{game.KeywordAscend}               // the badge; the engine reads it
//	Static: []game.StaticAbility{SelfPumpWhileCitysBlessing(3, 0)} // Snubhorn Sentry
//	Condition: YouHaveTheCitysBlessingCondition()               // Arch of Orazca
//
// `ascend` is a canonical keyword, so a deck-imported card gets it from
// Scryfall with no catalog entry; a catalog card that ships the rest of
// its text lists it in PrintedKeywords like any other printed keyword.
//
// Where a condition is read follows the card, as the monarch's does
// (monarch.go): a static is read by the layer pass (the engine bumps it
// when the designation is granted), an activation condition when the
// ability is offered and again as it is activated (CR 602.1b), an
// intervening "if" at the trigger and again at resolution (CR 603.4),
// and "if you have the city's blessing, instead" as the clause resolves.

// YouHaveTheCitysBlessing is "if you have the city's blessing" /
// "as long as you have the city's blessing". Caller holds g.mu.
func YouHaveTheCitysBlessing(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.CitysBlessingForEffect(you)
}

// YouHaveTheCitysBlessingCondition is "Activate only if you have the
// city's blessing" (Arch of Orazca, Orazca Relic, Timestream
// Navigator), as an ActivationCondition.
func YouHaveTheCitysBlessingCondition() ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		return YouHaveTheCitysBlessing(g, controller)
	}
}

// YouHaveTheCitysBlessingNow is the intervening "if you have the city's
// blessing" as a trigger condition (Twilight Prophet, Resplendent
// Griffin). Pair it with the event test through AllOf, and re-check
// with YouHaveTheCitysBlessing inside the effect: CR 603.4 asks again
// as the trigger resolves.
func YouHaveTheCitysBlessingNow(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return YouHaveTheCitysBlessing(g, source.Controller)
}

// WhileCitysBlessing is the AppliesTo of "as long as you have the
// city's blessing, <these permanents> …": it holds when the permanent
// under test passes `inner` and the SOURCE's controller has the
// blessing.
func WhileCitysBlessing(inner func(target *game.Card, g *game.Game, source *game.Card) bool) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return inner(target, g, source) && YouHaveTheCitysBlessing(g, source.Controller)
	}
}

// SelfWhileCitysBlessing is "as long as you have the city's blessing,
// THIS creature …".
func SelfWhileCitysBlessing() func(target *game.Card, g *game.Game, source *game.Card) bool {
	return WhileCitysBlessing(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && selfOnly(target, g, source)
	})
}

// SelfPumpWhileCitysBlessing is "This creature gets +p/+t as long as
// you have the city's blessing" (Snubhorn Sentry, Dusk Charger, Spire
// Winder): a layer 7c modification.
func SelfPumpWhileCitysBlessing(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: SelfWhileCitysBlessing(),
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// SelfKeywordWhileCitysBlessing is "This creature has <keyword> as long
// as you have the city's blessing" (Skymarcher Aspirant, Storm Fleet
// Swashbuckler, Slippery Scoundrel's hexproof): a layer 6 grant.
func SelfKeywordWhileCitysBlessing(keyword string) game.StaticAbility {
	return KeywordGrant(SelfWhileCitysBlessing(), keyword)
}

// CantAttackUnlessYouHaveTheCitysBlessing is "This creature can't
// attack unless you have the city's blessing" (Wayward Swordtooth):
// an ordinary layer-6 self restriction (attack_target_restrictions.go)
// whose one clause is about the attacking creature's controller.
func CantAttackUnlessYouHaveTheCitysBlessing() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{ControllerMustHaveCitysBlessing: true})
}

// CantBlockUnlessYouHaveTheCitysBlessing is the block half of the same
// sentence: "This creature can't … block unless you have the city's
// blessing", a CR 509.1b rule on the creature itself.
func CantBlockUnlessYouHaveTheCitysBlessing() game.BlockRule {
	return CantBlockUnless(YouHaveTheCitysBlessing,
		"it can't block unless its controller has the city's blessing")
}
