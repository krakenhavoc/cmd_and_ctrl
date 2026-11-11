package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s51_helpers.go — shared pieces of the batch 51 planeswalkers.

// cmdr2S51WolfToken is Garruk, Cursed Huntsman's 2/2 black and green Wolf
// with "When this creature dies, put a loyalty counter on each Garruk you
// control." Its slug is an on-disk identity: never renamed, never reused.
func cmdr2S51WolfToken() game.Card { return tokenFromCatalog(printedCmdr2S51WolfToken) }

func printedCmdr2S51WolfToken() tokenTemplate {
	return tokenTemplate{
		Slug: "wolf-garruk-loyalty",
		Card: game.Card{
			Name:      "Wolf",
			TypeLine:  "Token Creature — Wolf",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"B", "G"},
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Wolf — put a loyalty counter on each Garruk you control", cmdr2S51GarrukLoyalty),
		},
		Text: "When this creature dies, put a loyalty counter on each Garruk you control.",
	}
}

// cmdr2S51GarrukLoyalty puts a loyalty counter on each Garruk the dying
// token's controller controls when the trigger resolves.
func cmdr2S51GarrukLoyalty(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsPlaneswalker() && c.HasSubtype("Garruk") {
			ids = append(ids, c)
		}
	}
	for _, c := range ids {
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterLoyalty, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2S51TreefolkToken is Wrenn and Seven's green Treefolk with reach and
// "This creature's power and toughness are each equal to the number of
// lands you control."
func cmdr2S51TreefolkToken() game.Card { return tokenFromCatalog(printedCmdr2S51TreefolkToken) }

func printedCmdr2S51TreefolkToken() tokenTemplate {
	return tokenTemplate{
		Slug: "treefolk-lands",
		Card: game.Card{
			Name:     "Treefolk",
			TypeLine: "Token Creature — Treefolk",
			Colors:   []string{"G"},
			Keywords: []string{"reach"},
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "This creature's power and toughness are each equal to the number of lands you control.",
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.Controller == source.Controller && p.IsLand() {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Text: "This creature's power and toughness are each equal to the number of lands you control.",
	}
}

// cmdr2S51ArtifactAbilityYouActivate is the ability being priced: one the
// modifier's controller activates, of an artifact on the battlefield (any
// controller's). Mana abilities count as activated abilities, and the
// engine prices them through the same pass.
func cmdr2S51ArtifactAbilityYouActivate() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Controller == q.Source.Controller &&
			q.FromZone == game.ZoneBattlefield &&
			q.Card.IsArtifact()
	}
}

// cmdr2S51FirstArtifactActivationThisTurn passes while the activator has
// not yet activated any artifact's ability this turn (Tezzeret, Betrayer of
// Flesh). Reads Game.ArtifactActivationsThisTurn, written when an
// activation is announced (a countered one still counts, a refused one
// never does), and counted per activator across every artifact.
func cmdr2S51FirstArtifactActivationThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.ArtifactActivationsThisTurn(q.Controller) == 0
	}
}

// cmdr2S51ArtifactBecameTapped is "an artifact you control becomes tapped"
// for the source's controller: a tap event, or an attack that left the
// attacker tapped (an attack taps without a tap event; a vigilant attacker
// did not become tapped).
func cmdr2S51ArtifactBecameTapped(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTapCard, game.EventAttack:
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller || !c.IsArtifact() {
		return false
	}
	return ev.Kind != game.EventAttack || c.Tapped
}
