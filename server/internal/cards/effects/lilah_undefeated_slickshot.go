package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lilah, Undefeated Slickshot — Legendary Creature — Human Rogue {1}{U}{R},
// 3/3 (Outlaws of Thunder Junction, #4468):
//
//	"Prowess
//	 Whenever you cast a multicolored instant or sorcery spell from your
//	 hand, exile that spell instead of putting it into your graveyard as it
//	 resolves. If you do, it becomes plotted. (You may cast it as a sorcery
//	 on a later turn without paying its mana cost.)"
//
// The trigger marks the spell on the stack (Game.MarkSpellExilesAndPlotsForEffect),
// and the mark is the spell's own resolution exit (resolution_exits.go): a
// countered or fizzled spell goes to the graveyard as usual, and a spell
// that resolves is exiled and plotted once it is there. The cast must come
// from the hand, gated on the trigger so a flashback or plotted cast never
// puts it on the stack.
//
// No simplification.
func init() {
	mark := WheneverYouCast(And(Or(Instant(), Sorcery()), Multicolored()),
		"Lilah, Undefeated Slickshot — exile that spell instead of putting it into your graveyard as it resolves; it becomes plotted",
		func(g *game.Game, item *game.StackItem) error {
			ev := NewContext(g, item).Trigger().Event
			if ev.OldZone != game.ZoneHand || ev.CardID == uuid.Nil {
				return nil
			}
			g.MarkSpellExilesAndPlotsForEffect(ev.CardID)
			return nil
		})
	inner := mark.AppliesTo
	mark.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return ev.OldZone == game.ZoneHand && inner(ev, source, ch, g)
	}
	Register(Spec{
		OracleID:        "0757c5a7-e51b-4c39-bb4d-b32657c10cb4",
		Name:            "Lilah, Undefeated Slickshot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered:       []game.TriggeredAbility{mark},
	})
}
