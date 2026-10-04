package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Last March of the Ents — Sorcery {6}{G}{G}:
//
//	"This spell can't be countered.
//	 Draw cards equal to the greatest toughness among creatures you
//	 control, then put any number of creature cards from your hand
//	 onto the battlefield."
//
// "Can't be countered" is Spec.CantBeCountered (CR 101.2: a counter
// spell still resolves, it just does nothing to this one).
//
// The draw count is read as the spell resolves (Glint Weaver's
// greatest toughness, layered and with counters), zero with no
// creature. Then the caster picks any number of creature cards from
// the hand they now hold, the new cards included, and those enter the
// battlefield together (CR 603.6a), so each sees the others arrive.
// Picks are re-checked at the answer, since a prompt can wait.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3999cd64-ff5c-4e2c-8aa6-d14f9f8a2b4c",
		Name:            "Last March of the Ents",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := item.Controller
			n := glintWeaverGreatestToughnessControlledBy(ctx.Game, player)
			if n > 0 {
				if err := ctx.Game.DrawNForEffect(player, n); err != nil {
					return err
				}
			}
			return lastMarchPutCreatures(ctx.Game, item)
		},
	})
}

// lastMarchPutCreatures offers every creature card in the caster's
// hand and puts the picks onto the battlefield together.
func lastMarchPutCreatures(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	candidates := handCardsMatching(g, player, Creature())
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Last March of the Ents — put any number of creature cards from your hand onto the battlefield",
		Cards:    candidates,
		Min:      0,
		Max:      len(candidates),
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			entries := make([]game.BatchEntry, 0, len(picked))
			for _, id := range picked {
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneHand {
					continue
				}
				entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneHand})
			}
			if len(entries) == 0 {
				return nil
			}
			return g.PutOntoBattlefieldTogetherThenForEffect(entries, game.ZoneEntryOptions{Controller: player}, nil)
		},
	})
	return nil
}
