package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shark Shredder, Killer Clone — Legendary Creature — Shark Octopus Ninja
// {2}{B}{B}, 4/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Sneak {3}{B}{B}
//	 First strike
//	 Whenever Shark Shredder deals combat damage to a player, put up to one
//	 target creature card from that player's graveyard onto the battlefield
//	 under your control. It enters tapped and attacking that player."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The clause
// is built from the triggering event (TargetsFrom): "that player" is the
// player dealt the damage, so the card must be one they own in their
// graveyard, and the CR 608.2b re-check asks the same. The card is put onto
// the battlefield under the controller's control tapped and attacking that
// player, never declared as an attacker (CR 508.4), through the Grim
// Captain's entry (putAttacking). The clause reads only the event, never
// the board, so the trigger is a restore point.
//
// No simplification.
func init() {
	damage := WheneverThisDealsCombatDamageToAPlayer(
		"Shark Shredder — put up to one target creature card from that player's graveyard onto the battlefield under your control, tapped and attacking that player",
		sharkShredderReanimate)
	damage.TargetsFrom = sharkShredderClause
	Register(Spec{
		OracleID:         "dabd0867-ee3f-40cb-8314-b4a3966f8b24",
		Name:             "Shark Shredder, Killer Clone",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"first strike"},
		AlternativeCosts: []game.AlternativeCost{Sneak("{3}{B}{B}")},
		Triggered:        []game.TriggeredAbility{damage},
	})
}

// sharkShredderClause is "up to one target creature card from that
// player's graveyard", that player being the one the damage was dealt to.
func sharkShredderClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	return TargetCardInGraveyard("up to one target creature card from that player's graveyard",
		Creature(), OwnedBy(tc.Event.Target)).WithCount(0, 1)
}

// sharkShredderReanimate puts the still-legal pick onto the battlefield
// under the controller's control, tapped and attacking the damaged player.
func sharkShredderReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return putAttacking(g, ctx.Controller(), id, putAttackingForm{from: game.ZoneGraveyard}, ctx.Trigger().Event.Target)
}
