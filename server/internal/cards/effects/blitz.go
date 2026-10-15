package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blitz.go — blitz (CR 702.152, #3772):
//
//	"Blitz {cost} (If you cast this spell for its blitz cost, it gains
//	 haste and "When this creature dies, draw a card." Sacrifice it at
//	 the beginning of the next end step.)"
//
// An alternative cost (CR 118.9) and dash's sibling. The haste, the
// granted draw and the delayed sacrifice are the game-side clause
// AlternativeCost.Blitz (game/blitz.go); use these constructors, never a
// hand-built offer.
//
// The draw is a granted triggered ability (ADR 0093), and a grant is a
// catalog bundle. Every blitz creature gains the SAME ability, so the
// bundle is filed once, here, under game.BlitzDrawGrant, rather than once
// per card: a per-card key would be the same ability under seventeen
// names. Jaxis, the Troublemaker's token gains it too, through
// game.ApplyBlitzRidersForEffect.

func init() {
	registerSharedGrant("blitz", AbilityGrant{
		Key: game.BlitzDrawGrant,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Blitz — draw a card", func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
		},
		Text: "When this creature dies, draw a card.",
	})
}

// registerSharedGrant files a grant bundle no single card owns — one a
// keyword's rules give (blitz's draw). Same checks and same map as a
// Spec.Grants bundle; `owner` names the keyword in a panic.
func registerSharedGrant(owner string, gr AbilityGrant) {
	checkGrants(owner, []AbilityGrant{gr})
	fileDef(game.GrantKey(gr.Key), buildGrantDef(gr))
}

// Blitz is "Blitz {cost}" — CR 702.152a, cast from hand.
//
//	AlternativeCosts: []game.AlternativeCost{Blitz("{2}{R}")},   // Riveteers Requisitioner
func Blitz(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:      "blitz",
		Label:    "Blitz " + cost,
		ManaCost: cost,
		Blitz:    true,
	}
}

// BlitzDiscard is "Blitz—{cost}, Discard a card" (Sabin, Master Monk).
// The discard is a cost paid as the spell is cast, from the caster's own
// hand (CR 701.9a), so a discard payoff sees it and a countered spell
// does not give the card back.
func BlitzDiscard(cost string) game.AlternativeCost {
	ac := Blitz(cost)
	ac.Label = "Blitz—" + cost + ", Discard a card"
	ac.DiscardFromHand = CardInYourHand("a card")
	ac.PayLabel = "a card"
	return ac
}

// BlitzPayLife is "Blitz—{cost}, Pay N life" (Tenacious Underdog). The
// life is validated before anything is paid (CR 119.4): a player with
// less life is not offered it.
func BlitzPayLife(cost string, life int) game.AlternativeCost {
	ac := Blitz(cost)
	ac.Label = "Blitz—" + cost + ", Pay " + strconv.Itoa(life) + " life"
	ac.Life = life
	return ac
}

// BlitzFromHandOrGraveyard is a blitz offer plus "You may cast this card
// from your graveyard using its blitz ability" (Sabin, Master Monk;
// Tenacious Underdog): two offers with two keys, the shape Timeline
// Culler's warp uses. "blitz" is bound to the hand and "blitz-graveyard"
// to the graveyard, same price and same label (the picker only shows the
// offer bound to the zone being cast from). The card also lists
// game.ZoneGraveyard in CastableZones.
//
// The binding cuts both ways, and both halves matter: the graveyard
// opens only through its own offer, so the card cannot be cast from
// there for its mana cost, and the hand offer — bound explicitly, since
// an unbound offer is claimable from any zone the card may be cast
// from — cannot be claimed from the graveyard.
func BlitzFromHandOrGraveyard(offer game.AlternativeCost) []game.AlternativeCost {
	hand := offer
	hand.FromZone = game.ZoneHand
	grave := offer
	grave.Key = "blitz-graveyard"
	grave.FromZone = game.ZoneGraveyard
	return []game.AlternativeCost{hand, grave}
}
