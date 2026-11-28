package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lodestone Bauble — Artifact {0} (Alliances, #5634):
//
//	"{1}, {T}, Sacrifice this artifact: Put up to four target basic land cards
//	 from a player's graveyard on top of their library in any order. That
//	 player draws a card at the beginning of the next turn's upkeep."
//
// The targets are up to four basic land cards from ONE graveyard (a set rule,
// UpToCardsFromASingleGraveyard); each still legal as the ability resolves goes
// on top of its owner's library, the controller choosing the order (CR
// 401.4), and the owner of that graveyard draws at the next upkeep at the
// table (CR 603.7a). With no card chosen there is no "that player", so nothing
// is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6048b9ac-7d5e-4486-a89a-30adb45aab27",
		Name:         "Lodestone Bauble",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Put up to four target basic land cards from a player's graveyard on top of their library in any order. That player draws a card at the beginning of the next turn's upkeep.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.UpToCardsFromASingleGraveyard("up to four target basic land cards from a player's graveyard", 4, effects.OldB814BasicLandCard()),
			Effect:  oldB814LodestoneBauble,
		}},
	})
}

func oldB814LodestoneBauble(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	var owner uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		ids = append(ids, t.ID)
		owner = c.Owner
	}
	if len(ids) == 0 {
		return nil
	}
	if err := effects.OldB814DrawAtTheNextUpkeep("Lodestone Bauble — that player draws a card", owner, 1).Apply(ctx); err != nil {
		return err
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   ctx.Controller(),
		Source:    item.SourceCardID,
		Cards:     ids,
		From:      game.ZoneGraveyard,
		Placement: game.LibraryPlaceTop,
		Reason:    "Lodestone Bauble — put the cards on top of their library in any order",
	})
}
