package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arvinox, the Mind Flail — Legendary Enchantment Creature — Horror
// {4}{B}{B}{B}, 9/9 (Duskmourn Commander, #5457):
//
//	"Arvinox isn't a creature unless you control three or more permanents
//	 you don't own.
//	 At the beginning of your end step, exile the bottom card of each
//	 opponent's library face down. For as long as those cards remain
//	 exiled, you may look at them, you may cast permanent spells from
//	 among them, and you may spend mana as though it were mana of any
//	 color to cast those spells."
//
// The first ability is a Theros God's clause with another count: a layer-4
// type loss, so Arvinox stays a legendary enchantment and loses Horror
// with the creature type (CR 205.1a). The count is Agent of Treachery's
// (permanentsYouControlButDontOwn), read off the board, so Arvinox, which
// its controller usually owns, does not count itself.
//
// The end step trigger takes index 0 of each opponent's library (the
// bottom), in seat order, and exiles them face down with one grant each
// to Arvinox's controller: that grant is what lets the controller, and
// nobody else, look at the card. It is a cast permission, so a land stays
// exiled, limited to permanent spells (an instant or sorcery stays
// exiled too), spending mana as though it were any color (not any type),
// for as long as the card remains exiled, whether or not Arvinox is
// still on the battlefield. An opponent with an empty library exiles
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9abecd23-a2d9-43c7-8d01-f8afdc53026e",
		Name:         "Arvinox, the Mind Flail",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Isn't a creature unless you control three or more permanents you don't own.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID &&
					permanentsYouControlButDontOwn(g, source.Controller) < 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				notACreature(c)
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Arvinox, the Mind Flail — exile the bottom card of each opponent's library face down; you may cast permanent spells from among them",
				arvinoxExileBottoms),
		},
	})
}

// arvinoxExileBottoms is Arvinox's end step trigger resolving.
func arvinoxExileBottoms(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
			continue
		}
		// Index 0 is the bottom (Zone.Bottom).
		ids = append(ids, p.Library.Cards[0].InstanceID)
	}
	return g.ExileCardsFaceDownWithPermissionForEffect(ids, item.Controller, game.CastPermission{
		CastOnly:   true,
		AnyColor:   true,
		Duration:   game.WhileInZoneDuration(),
		Filter:     game.PermissionFilter{NonLandPermanentOnly: true},
		Source:     item.SourceCardID,
		SourceName: "Arvinox, the Mind Flail",
		Label:      "For as long as it remains exiled, you may cast it if it's a permanent spell, spending mana as though it were mana of any color",
	})
}
