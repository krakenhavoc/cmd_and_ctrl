package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Matterbending Mage — Creature — Human Wizard {2}{U}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, return up to one other target creature to
//	 its owner's hand.
//	 Whenever you cast a spell with {X} in its mana cost, this creature
//	 can't be blocked this turn."
//
// The bounce targets up to one creature other than the Mage, any
// controller's, as the trigger goes on the stack; with no target it does
// nothing, and a target that left is skipped (CR 608.2b). The second
// ability reads the cost printed on the spell (socB29SpellWithXInItsManaCost),
// so an {X} that is only in an additional cost does not count. Each such cast
// is its own trigger, and the evasion is a turn-long restriction on the Mage
// as it is now: if it left in response it is a new object and gets nothing
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94688e25-28cb-45aa-95ea-0d6f547cae57",
		Name:         "Matterbending Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Matterbending Mage — return up to one other target creature to its owner's hand", bounceChosenTarget),
				Another(UpToOneTargetCreature("up to one other target creature")),
			),
			WheneverYouCast(socB29SpellWithXInItsManaCost,
				"Matterbending Mage — this creature can't be blocked this turn",
				matterbendingMageEvasion),
		},
	})
}

// matterbendingMageEvasion makes the Mage unblockable for the turn.
func matterbendingMageEvasion(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return RestrictUntilEOT{
		Target:       item.SourceCardID,
		Restrictions: game.CantBeBlocked,
		Label:        "Matterbending Mage — can't be blocked",
	}.Apply(NewContext(g, item))
}
