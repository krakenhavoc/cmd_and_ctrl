package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roxi, Publicist to the Stars — Legendary Creature — Human Employee
// {2}{U}{R}, */4:
//
//	"Flying
//	 Roxi's power is equal to the number of permanents you control with an
//	 art sticker plus the number of cards in your graveyard with an art
//	 sticker.
//	 When Roxi enters, distribute up to two art stickers among one or two
//	 nonland permanents you own."
//
// The power is a characteristic-defining ability (layer 7a, CR 604.3),
// read every recompute from the stickers on the battlefield and in the
// controller's graveyard, which keep them there (CR 123.5).
//
// "Distribute up to two art stickers among one or two nonland permanents"
// is two "you may put an art sticker on a nonland permanent you own"
// instructions in order (CR 123.3): each picks its own sticker and its own
// permanent, so the two can go on one permanent or on two, and either can
// be declined. Art stickers cost no tickets.
//
// No simplification.
func init() {
	art := []game.StickerKind{game.StickerArt}
	Register(Spec{
		OracleID:        "5e24a352-7c90-458b-9bb2-bd4bc22bc748",
		Name:            "Roxi, Publicist to the Stars",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of permanents you control with an art sticker plus the number of cards in your graveyard with an art sticker.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = artStickeredPermanentsAndGraveyardCards(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Roxi, Publicist to the Stars — distribute up to two art stickers among one or two nonland permanents you own",
				Do(
					PutASticker{On: StickerOnNonlandPermanentYouOwn, Kinds: art},
					PutASticker{On: StickerOnNonlandPermanentYouOwn, Kinds: art},
				)),
		},
	})
}

// artStickeredPermanentsAndGraveyardCards is "the number of permanents you
// control with an art sticker plus the number of cards in your graveyard
// with an art sticker". It walks the zones directly: a layer pass may hold
// only the read lock.
func artStickeredPermanentsAndGraveyardCards(g *game.Game, you uuid.UUID) int {
	n := 0
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == you && game.StickerCountOfKind(c, game.StickerArt) > 0 {
				n++
			}
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.ID != you || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			if game.StickerCountOfKind(&p.Graveyard.Cards[i], game.StickerArt) > 0 {
				n++
			}
		}
	}
	return n
}
