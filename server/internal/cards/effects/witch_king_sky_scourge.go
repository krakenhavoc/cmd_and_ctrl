package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witch-king, Sky Scourge — Legendary Creature — Wraith Noble
// {5}{B}{R}, 5/5:
//
//	"Flying
//	 Whenever you attack with one or more Wraiths, exile the top X cards
//	 of your library, where X is their total power. You may play those
//	 cards this turn.
//	 Undying"
//
// One trigger per attack declaration (OncePerBatch over the batch's
// EventAttack). X is read as the trigger resolves (CR 608.2h): each
// Wraith declared in that batch counts its power then, or as it last
// existed if it has left the battlefield. Flying and undying are
// PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "84137013-7a23-4bc8-b855-04889373fdae",
		Name:            "Witch-king, Sky Scourge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasSubtype("Wraith")
			}, "Witch-king, Sky Scourge — exile the top X cards, X the attacking Wraiths' total power; you may play them this turn",
				witchKingImpulse)),
		},
	})
}

func witchKingImpulse(g *game.Game, item *game.StackItem) error {
	batch := NewContext(g, item).Trigger().Event.Batch
	total := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventAttack || ev.Batch != batch || ev.Actor != item.Controller {
			continue
		}
		if c, ok := battlefieldCardForEffect(g, ev.CardID); ok {
			if c.HasSubtype("Wraith") && c.PowerForComparison() > 0 {
				total += c.PowerForComparison()
			}
			continue
		}
		if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok && info.Power > 0 &&
			(info.Characteristic.AllCreatureTypes || hasFold(info.Characteristic.Subtypes, "Wraith")) {
			total += info.Power
		}
	}
	if total <= 0 {
		return nil
	}
	_, err := b12ImpulseExileForTurn(g, item, total)
	return err
}

// battlefieldCardForEffect is the permanent with this ID, when it is on
// the battlefield.
func battlefieldCardForEffect(g *game.Game, id uuid.UUID) (game.Card, bool) {
	if !onBattlefield(g, id) {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(id)
}
