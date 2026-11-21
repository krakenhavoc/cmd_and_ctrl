package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Martha Jones — Legendary Creature — Human Cleric {2}{U}, 3/2 (Doctor Who,
// #5948):
//
//	"Woman Who Walked the Earth — When Martha Jones enters, investigate.
//	 Whenever you sacrifice a Clue, Martha Jones and up to one other target
//	 creature can't be blocked this turn.
//	 Doctor's companion"
//
// The entry trigger is Investigate. The sacrifice trigger is Eloise's
// reading of a sacrifice: the event is announced while the Clue is still on
// the battlefield, so its subtype is read there, and only the controller's
// own sacrifice counts. Martha herself is always made unblockable (when she
// is still on the battlefield), the target is optional ("up to one other"),
// and the clause is a restriction on blocks until end of turn. Doctor's
// companion is a deck-construction rule (internal/deck reads it off the
// oracle text).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b7ccdd2-40aa-4619-a147-0792dc02b2e5",
		Name:         "Martha Jones",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Martha Jones — investigate", effects.Do(effects.Investigate{N: 1})),
			effects.Targeting(
				effects.On(game.EventSacrifice, youSacrificedAClue,
					"Martha Jones — she and up to one other target creature can't be blocked this turn",
					marthaJonesUnblockable),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature")),
			),
		},
	})
}

// youSacrificedAClue is "whenever you sacrifice a Clue".
func youSacrificedAClue(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Clue")
}

// marthaJonesUnblockable makes Martha and the chosen creature unblockable.
func marthaJonesUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if src, ok := g.LookupCardForEffect(ctx.Source()); ok && g.FindCardZoneForEffect(src.InstanceID) != nil {
		if err := (effects.RestrictUntilEOT{Target: src.InstanceID, Restrictions: game.CantBeBlocked,
			Label: "Martha Jones — can't be blocked"}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.TargetCantBeBlockedThisTurn("Martha Jones")(g, item)
}
