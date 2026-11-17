package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gate to the Aether — Artifact {6} (Mirrodin, #5606):
//
//	"At the beginning of each player's upkeep, that player reveals the
//	 top card of their library. If it's an artifact, creature,
//	 enchantment, or land card, the player may put it onto the
//	 battlefield."
//
// An each-upkeep trigger, controlled by the Gate's controller, whose
// "that player" is the player whose upkeep it is (the event's actor).
// That player reveals their top card to the table (CR 701.20a); a card
// of one of the four types may go onto the battlefield under their
// control, their choice (PutFromLibraryOntoBattlefield, optional). Any
// other card, or one they decline, stays on top of the library. An
// empty library reveals nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a03ab63a-83f3-4dd6-926d-08d09ce2e696",
		Name:         "Gate to the Aether",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Gate to the Aether — that player reveals the top card of their library and may put an artifact, creature, enchantment or land card onto the battlefield",
				gateToTheAetherReveal),
		},
	})
}

// gateToTheAetherReveal is the trigger's resolution, for the player whose
// upkeep it is.
func gateToTheAetherReveal(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	player := item.Trigger.Event.Actor
	revealed := g.RevealTopOfLibraryForEffect(player, item.SourceCardID, 1, "Gate to the Aether — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    revealed,
		Match:    Or(Artifact(), Creature(), Enchantment(), Land()),
		Max:      1,
		Optional: true,
		Label:    "Gate to the Aether — you may put the revealed card onto the battlefield",
	}.Apply(NewContext(g, item))
}
