package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Abyss — World Enchantment {3}{B}:
//
//	"At the beginning of each player's upkeep, destroy target
//	 nonartifact creature that player controls of their choice. It
//	 can't be regenerated."
//
// #1997, ADR 0109 amendment 2026-10-23 with ADR 0019's of the same day.
// The trigger fires on every player's upkeep (AtEachUpkeep). Its target
// is chosen by the player whose upkeep it is, "that player" of the
// upkeep event, though the trigger's controller still controls it and
// still targets (the 2004 ruling: "The controller of The Abyss controls
// the ability, but the current player chooses what gets destroyed"):
// OfThatPlayersChoice sends the CR 603.3d target prompt to them, and
// TheyControl keeps the choice to the creatures they control. Legality
// is judged for the controller, so an upkeep player's hexproof creature
// is not a legal choice for an opponent's Abyss. With no nonartifact
// creature they control that can be targeted, the trigger is removed
// (CR 603.3d). The target is checked again as it resolves (CR 608.2b),
// and the destruction ignores regeneration (CR 701.19c).
//
// The clause is Harmful (#4249): on its controller's own upkeep the bot
// picks among its own creatures with no `chosen_for` on the prompt, and
// the flag is what makes it give up the least valuable one.
//
// A world permanent: the world rule (CR 704.5k) puts it into its owner's
// graveyard when a newer one enters. Magus of the Abyss prints the same
// trigger (magus_of_the_abyss.go).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a02a7816-c967-4503-bb08-f8db44915250",
		Name:         "The Abyss",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{theAbyssTrigger("The Abyss")},
	})
}

// theAbyssTrigger is the upkeep trigger The Abyss and Magus of the Abyss
// share, labelled with the card's name.
func theAbyssTrigger(name string) game.TriggeredAbility {
	return Targeting(
		AtEachUpkeep(name+" — destroy target nonartifact creature that player controls of their choice; it can't be regenerated", theAbyssDestroy),
		Harmful(TheyControl(OfThatPlayersChoice(TargetCreature("target nonartifact creature that player controls of their choice", Not(Artifact()))))),
	)
}

// theAbyssDestroy destroys the chosen creature if it is still a legal
// target, ignoring regeneration.
func theAbyssDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind == game.TargetCard {
			return DestroyTarget{Target: ref.ID, CantBeRegenerated: true}.Apply(ctx)
		}
	}
	return nil
}
