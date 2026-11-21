package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b15_helpers.go — Adventures in the Forgotten Realms, S82 batch 15
// (#5499): the planeswalkers and token-making legends.

// afrB15SpiderToken is the printed 2/1 black Spider with menace and reach
// (Drider, Lolth's −3), as a tokens_table.go key.
const AfrB15SpiderToken = "2/1 black Spider with menace and reach"

// afrB15BooToken is Minsc's Boo, a legendary 1/1 red Hamster with trample
// and haste, as a tokens_table.go key.
const AfrB15BooToken = "1/1 red Legendary Hamster named Boo with trample and haste"

// afrB15GuenhwyvarToken is Drizzt's Guenhwyvar, a legendary 4/1 green Cat
// with trample, as a tokens_table.go key.
const AfrB15GuenhwyvarToken = "4/1 green Legendary Cat named Guenhwyvar with trample"

// printedAfrB15DogIllusionToken is Mordenkainen's blue Dog Illusion: its
// power and toughness are each twice the number of cards in its controller's
// hand, a characteristic-defining ability (CR 604.3) the token always
// prints. Its slug is an on-disk identity: never renamed, never reused.
func printedAfrB15DogIllusionToken() TokenTemplate {
	const text = "This creature's power and toughness are each equal to twice the number of cards in your hand."
	return TokenTemplate{
		Slug: "dog-illusion",
		Card: game.Card{
			Name:              "Dog Illusion",
			TypeLine:          "Token Creature — Dog Illusion",
			Colors:            []string{"U"},
			VariableToughness: true, // printed */*
		},
		Static: []game.StaticAbility{{
			Label:             text,
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7A_CDA,
			DependsOnHandSize: true,
			AppliesTo:         SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				if p := g.PlayerByIDForEffect(source.Controller); p != nil {
					n = 2 * p.Hand.Size()
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Text: text,
	}
}

// AfrB15DogIllusionToken is that Dog Illusion's card value for CreateToken.
func AfrB15DogIllusionToken() game.Card { return TokenFromCatalog(printedAfrB15DogIllusionToken) }

// printedAfrB15IllusionToken is Minn's 1/1 blue Illusion with "This creature
// gets +1/+0 for each other Illusion you control." — a layer 7c modification
// counted live (CR 613.4c), the Illusions read through their creature-type
// words. Its slug is an on-disk identity: never renamed, never reused.
func printedAfrB15IllusionToken() TokenTemplate {
	const text = "This creature gets +1/+0 for each other Illusion you control."
	return TokenTemplate{
		Slug: "illusion-pumped-by-other-illusions",
		Card: game.Card{
			Name:      "Illusion",
			TypeLine:  "Token Creature — Illusion",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"U"},
		},
		Static: []game.StaticAbility{{
			Label:     text,
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, o := range g.BattlefieldCardsForEffect() {
					if o.InstanceID != source.InstanceID && o.Controller == source.Controller &&
						o.HasSubtype(source.TextSubtype("Illusion")) {
						n++
					}
				}
				c.Power += n
			},
		}},
		Text: text,
	}
}

// AfrB15IllusionToken is that Illusion's card value for CreateToken.
func AfrB15IllusionToken() game.Card { return TokenFromCatalog(printedAfrB15IllusionToken) }

// afrB15CreatureDiedWithMorePower is Drizzt's intervening "if": a creature
// died (any controller) whose power as it last existed on the battlefield
// (CR 603.10a) is greater than the source's power now. It returns the dead
// creature's id so the resolution can recheck (CR 603.4).
func AfrB15CreatureDiedWithMorePower(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	if !ok {
		return false
	}
	return B17LastKnownPowerOffBattlefield(g, dead.InstanceID) > source.CurrentPower()
}
