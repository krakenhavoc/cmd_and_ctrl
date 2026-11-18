package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b19_helpers.go — the Streets of New Capenna triome-creature
// cycle's hand ability (Masked Bandits, Rakish Revelers, Spara's
// Adjudicators, Glamorous Outlaw, Shattered Seraph):
//
//	"{2}, Exile this card from your hand: Target land gains "{T}: Add
//	 <X>, <Y>, or <Z>" until this card is cast from exile. You may cast
//	 this card for as long as it remains exiled."
//
// It is Emrakul, the Exigent Doom's ability (#3520) with a colour choice
// in place of {C}{C}: an activated ability from the hand whose exile is a
// cost, a duration grant that ends when THIS card is cast from exile, and a
// stored cast permission that lasts while it stays exiled.

// sncB19ExileForLand is the cycle's hand ability and the grant it hands out.
// `key` is the grant's on-disk identity, `produced` the colour pipe
// ("{B|R|G}"), `colors` the oracle's colour list ("{B}, {R}, or {G}").
func sncB19ExileForLand(name, key, produced, colors string) (AbilityGrant, ActivatedAbility) {
	grant := TapForManaGrant(key, produced, "Add "+colors, "{T}: Add "+colors+".")
	ability := ActivatedAbility{
		Label: "{2}, Exile this card from your hand: Target land gains \"{T}: Add " + colors +
			"\" until this card is cast from exile. You may cast this card for as long as it remains exiled.",
		Cost:    Plus(ManaCost("{2}"), ExileThis()),
		Zones:   []game.ZoneKind{game.ZoneHand},
		Targets: TargetPermanent("target land", Land()),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			land, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card := item.SourceCardID
			d, inExile := g.UntilCardCastDuration(card)
			if !inExile {
				d = game.IndefiniteDuration()
			}
			if err := (GrantAbilitiesFor{
				Target:   land,
				Keys:     []string{key},
				Duration: d,
				Label:    name + " — the land taps for " + colors + " until " + name + " is cast from exile",
			}).Apply(ctx); err != nil {
				return err
			}
			if inExile {
				g.GrantCastPermissionOverCardForEffect(card, game.CastPermission{
					Player:     item.Controller,
					Duration:   game.WhileInZoneDuration(),
					CastOnly:   true,
					Source:     card,
					SourceName: name,
					Label:      name + " — cast it from exile",
				})
			}
			return nil
		},
	}
	return grant, ability
}
