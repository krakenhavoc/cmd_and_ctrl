package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarlet Spider, Ben Reilly — Legendary Creature — Spider Human Hero
// {1}{R}{G}, 4/3 (Marvel's Spider-Man promos, #4460):
//
//	"Web-slinging {R}{G} (You may cast this spell for {R}{G} if you also
//	 return a tapped creature you control to its owner's hand.)
//	 Trample
//	 Sensational Save — If Scarlet Spider was cast using web-slinging, he
//	 enters with X +1/+1 counters on him, where X is the mana value of the
//	 returned creature."
//
// Web-slinging is Spider-Sense's alternative cost: the returned creature is
// a cost, chosen and returned as the spell is cast, and the cast records it
// (PaidCost.AltCostObjects, ADR 0135 §4).
//
// Sensational Save is an "enters with" replacement (CR 614.1c) on his own
// entry: it reads the resolving spell's alternative cost through the entry
// event (game.EntryAltCostForEffect) and X off the returned creature as it
// last existed on the battlefield (CR 608.2h), so a creature with {X} in
// its cost counts 0 for it (CR 202.3e) and a token that isn't a copy counts
// 0. Cast for his mana cost, or put onto the battlefield without being
// cast, he enters with none. "Cast using web-slinging" is the card's own
// offer or one a permanent grants it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13ab2f3f-f65e-42b5-b743-39b30a4c4751",
		Name:            "Scarlet Spider, Ben Reilly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{{
			Key:          scarletSpiderWebSlinging,
			Label:        "Web-slinging {R}{G}",
			ManaCost:     "{R}{G}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		Replacements: []game.ReplacementEffect{{
			RowLabel:  "Sensational Save — If cast using web-slinging, enters with X +1/+1 counters, where X is the mana value of the returned creature.",
			EntryOnly: true,
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID != src.InstanceID {
					return false
				}
				key, _ := g.EntryAltCostForEffect(ev)
				return scarletSpiderCastUsingWebSlinging(key)
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				_, refs := g.EntryAltCostForEffect(ev)
				if returned := g.AltCostPermanentsForEffect(refs); len(returned) > 0 && returned[0].ManaValue > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, returned[0].ManaValue)
				}
				return nil
			},
			Label: "Scarlet Spider, Ben Reilly: Sensational Save",
		}},
	})
}

// scarletSpiderWebSlinging is web-slinging's offer key, the one every
// printed web-slinging card uses. An on-disk identity.
const scarletSpiderWebSlinging = "web_slinging"

// scarletSpiderCastUsingWebSlinging is "was cast using web-slinging": the
// card's own offer, or web-slinging a permanent grants it ("granted-web-
// slinging", Peter Parker // Amazing Spider-Man's key, which grants it to
// legendary spells such as this one).
func scarletSpiderCastUsingWebSlinging(key string) bool {
	return key == scarletSpiderWebSlinging || key == "granted-web-slinging"
}
