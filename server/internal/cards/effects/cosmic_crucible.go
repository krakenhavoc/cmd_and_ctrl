package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cosmic Crucible — {4}{G}{U} Enchantment:
//
//	"At the beginning of your first main phase, add four mana in any
//	 combination of colors.
//	 Whenever you cast a noncreature spell, you may copy it. You may choose
//	 new targets for the copy. Do this only once each turn. (A copy of a
//	 permanent spell becomes a token.)"
//
// The mana is a precombat-main trigger (the first main phase is the
// precombat one, CR 505.1, so it never fires in the second) adding four
// INDEPENDENT any-colour slots (AnyCombinationOfColors: the controller
// names each, so {W}{U}{B}{R} is legal).
//
// The copy trigger is Baron Strucker's once-each-turn shape: "only once
// each turn" limits the DOING, not the offer, so declining the first
// noncreature spell does not use the turn up and a second spell cast
// before the first copy is made is offered too but only the first "yes"
// copies. The copy is made from last-known information, since "copy it"
// does not target (CR 608.2h, CR 707.10): the spell being countered in
// response to the trigger still gets its copy.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	const (
		manaLabel = "Cosmic Crucible — add four mana in any combination of colors"
		copyLabel = "Cosmic Crucible — you may copy the noncreature spell"
	)
	Register(Spec{
		OracleID:     "2469c0b7-a045-4733-bf53-28f3fdbe3797",
		Name:         "Cosmic Crucible",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain(manaLabel, Do(AddMana{Produced: AnyCombinationOfColors(4)})),
			Optional(On(game.EventCast, AllOf(YouCast(Noncreature()), msc18CrucibleNotYetCopied),
				copyLabel, msc18CrucibleCopy),
				"Cosmic Crucible — copy that spell?"),
		},
	})
}

const msc18CrucibleDone = "Cosmic Crucible — copied a spell this turn"

// msc18CrucibleNotYetCopied is "do this only once each turn": a copy was not
// already made by this Crucible this turn.
func msc18CrucibleNotYetCopied(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && !b28DoneThisTurn(g, source.InstanceID, msc18CrucibleDone)
}

// msc18CrucibleCopy copies the spell that triggered the ability, if no copy
// was made this turn.
func msc18CrucibleCopy(g *game.Game, item *game.StackItem) error {
	if b28DoneThisTurn(g, item.SourceCardID, msc18CrucibleDone) {
		return nil
	}
	b28NoteDoneThisTurn(g, item.SourceCardID, msc18CrucibleDone)
	return CopySpell{
		StackID:          item.Trigger.Event.CardID,
		Controller:       item.Controller,
		ChooseNewTargets: true,
		FromLastKnown:    true,
	}.Apply(NewContext(g, item))
}
