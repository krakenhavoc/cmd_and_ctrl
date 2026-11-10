package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hoarding Dragon — Creature — Dragon {3}{R}{R}, 4/4 (Foundations, #4465):
//
//	"Flying
//	 When this creature enters, you may search your library for an
//	 artifact card, exile it, then shuffle.
//	 When this creature dies, you may put the exiled card into its
//	 owner's hand."
//
// The last two abilities are linked (CR 607.2a): the search exiles the card
// face up linked to this Dragon OBJECT (SearchLibrary.LinkTo), and the dies
// trigger reads only the cards linked to that object back (ExiledWithForEffect),
// so each Dragon has its own hoard. A Dragon that died before its search
// resolved still stamps the link, and that card stays in exile for good (the
// ruling): the dies trigger has already resolved and found nothing. The
// "you may" on the dies trigger is its optional prompt, and the card goes
// to its OWNER's hand.
//
// No simplification.
func init() {
	const name = "Hoarding Dragon"
	Register(Spec{
		OracleID:        "d36d3c93-ed24-4447-9d15-3071f5c0989b",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(WhenThisEnters(name+" — you may search your library for an artifact card, exile it, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ref, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() },
						Dest:      game.ZoneExile,
						Limit:     1,
						Optional:  true,
						Shuffle:   true,
						LinkTo:    game.PermissionCardRef(ref),
						Reason:    name + " — search your library for an artifact card",
					}.Apply(NewContext(g, item))
				}), game.Purpose{}),
			Optional(WhenThisDies(name+" — you may put the exiled card into its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ref, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
						ids = append(ids, c.InstanceID)
					}
					for _, id := range ids {
						if _, err := g.ReturnFromAnyZoneToHandForEffect(item.Controller, item.SourceCardID, id); err != nil {
							return err
						}
					}
					return nil
				}), name+" — put the exiled card into its owner's hand?"),
		},
	})
}
