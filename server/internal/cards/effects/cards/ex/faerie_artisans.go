package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faerie Artisans — Creature — Faerie Artificer {3}{U}, 2/2 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"Flying
//	 Whenever a nontoken creature an opponent controls enters, create a token
//	 that's a copy of that creature except it's an artifact in addition to its
//	 other types. Then exile all other tokens created with this creature."
//
// Abyssal Harvester's copy-then-exile-the-others shape. The copy is Saheeli's
// "also an artifact" template exception. Each token is stamped with the
// Artisans OBJECT that made it (game.Card.CreatedWith, CR 607.1d,
// Tombstone Stairwell's link), so "all other tokens created with this
// creature" is exactly the tokens that object made earlier, and an Artisans
// that left and came back has no claim on the old ones (CR 400.7). The link
// and the new token ride the creation's keyed tail.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7dda40f2-7b5f-479c-86b6-8bc11fff9f4a",
		Name:            "Faerie Artisans",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && !c.IsToken() && c.Controller != source.Controller
			}, "Faerie Artisans — create a token copy of that creature, except it's an artifact; then exile all other tokens created with this creature",
				faerieArtisansCopy), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		},
	})
}

func faerieArtisansCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		return nil
	}
	tmpl, ok := effects.TokenCopyTemplate(g, item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	effects.TokenCopyIsAlsoAnArtifact(&tmpl)
	link := game.PermissionCardRef(ref)
	next := ctx.ItemTail(faerieArtisansExileOthers)
	next.Carry = []uuid.UUID{link.ID}
	next.N = link.Epoch
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: effects.CopiedToken(tmpl), Count: 1, CreatedWith: link}},
	}, next)
}

// faerieArtisansExileOthers is "Then exile all other tokens created with this
// creature": every token linked to the Artisans object except the ones the
// creation just made. An on-disk key: never renamed, never reused.
var faerieArtisansExileOthers = effects.CardTail("faerie-artisans/exile-other-tokens", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	link := game.PermissionCardRef{ID: o.Tail.Carry[0], Epoch: o.Tail.N}
	spared := map[uuid.UUID]bool{}
	for _, id := range o.IDs {
		spared[id] = true
	}
	for _, id := range ctx.Game.TokensCreatedWithForEffect(link) {
		if spared[id] {
			continue
		}
		if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
