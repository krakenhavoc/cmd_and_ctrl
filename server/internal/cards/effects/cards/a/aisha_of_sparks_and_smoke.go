package cardsa

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aisha of Sparks and Smoke — Legendary Creature — Human Warrior {1}{R}{R},
// 4/2:
//
//	"Prowess
//	 {R/W}: Aisha of Sparks and Smoke gains first strike until end of turn.
//	 Whenever Aisha deals combat damage, you may cast a sorcery spell from your
//	 hand with mana value less than or equal to that damage without paying its
//	 mana cost."
//
// The trigger is Buster Sword's free cast, to anything Aisha hits (a creature
// or a player), not only a player: the damage is read off the trigger's event,
// so lifelink, a doubler or a prevented half counts what was actually dealt.
// The offer is a pick among the hand cards with a sorcery spell of mana value
// no greater than the damage, judged on the face being cast (CR 601.3e, 709.3a),
// or none; the pick opens the cast at once, at any timing (CR 608.2g), through
// the ordinary announce path with its targets, modes and additional costs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eae3e762-dacd-4bd2-923c-3abb5ceb729a",
		Name:            "Aisha of Sparks and Smoke",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R/W}: Aisha of Sparks and Smoke gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{R/W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{
					Target:   ctx.Source(),
					Keywords: []string{"first strike"},
					Label:    "Aisha of Sparks and Smoke — first strike",
				}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventDealDamage && ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID
			}, "Aisha of Sparks and Smoke — you may cast a sorcery spell from your hand with mana value less than or equal to that damage without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					damage := effects.ThatDamage(item)
					if damage < 0 {
						return nil
					}
					effects.OfferFreeHandCast(ctx, ctx.Controller(), aishaFreeCast,
						"a sorcery spell with mana value "+strconv.Itoa(damage)+" or less",
						effects.FreeCastSpell{Filter: game.PermissionFilter{CardTypesAny: "sorcery"}}.WithManaValueAtMost(damage))
					return nil
				}),
		},
	})
}

// aishaFreeCast is the offer's keyed continuation.
var aishaFreeCast = effects.RegisterFreeHandCastOffer("choose-cards/aisha-of-sparks-and-smoke-cast", "Aisha of Sparks and Smoke")
