package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Custodi Squire — Creature — Spirit Cleric {4}{W}, 3/3:
//
//	"Flying
//	 Will of the council — When this creature enters, starting with
//	 you, each player votes for an artifact, creature, or enchantment
//	 card in your graveyard. Return each card with the most votes or
//	 tied for most votes to your hand."
//
// An object vote (CR 701.38b, ADR 0146) over CARDS IN A GRAVEYARD, which
// is the case the amendment of 2026-10-13 added: the vote knows the zone
// its options are in (Vote.Zone), and an option whose card has left that
// graveyard is no longer offered. The options are the controller's
// graveyard as the trigger resolves (nontoken; the three types in the
// printed list), fixed then; nothing is targeted. Every card tied for
// the most votes that is still in the graveyard returns, to its owner's
// hand. With nothing to vote for nobody votes and nothing returns.
//
// A bot casting it votes for the best card in its own graveyard; an
// opponent's bot votes for the worst (aiseat/heuristic/vote.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "07c7193e-f591-4aee-a3cc-3a2c52a38f9b",
		Name:            "Custodi Squire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Custodi Squire — each player votes for an artifact, creature, or enchantment card in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ids := graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
						return !c.IsToken() && (c.IsArtifact() || c.IsCreature() || c.IsEnchantment())
					})
					return Vote{
						Question: "Custodi Squire — vote for an artifact, creature, or enchantment card in its controller's graveyard",
						Options:  VoteForPermanents(g, ids),
						Zone:     game.ZoneGraveyard,
						Then:     custodiSquireVoted,
					}.Apply(ctx)
				}),
		},
	})
}

var custodiSquireVoted = VoteResultThen("vote/custodi-squire", func(ctx *Context, r game.VoteResult) error {
	var firstErr error
	for _, i := range r.MostVotes() {
		id := VotedPermanent(r, i)
		if id == uuid.Nil {
			continue
		}
		// A card that left the graveyard while the players voted is a
		// new object (CR 400.7) and is not returned.
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
})
