package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terminal Velocity — Sorcery {4}{R}{R} (Edge of Eternities, #4461):
//
//	"You may put an artifact or creature card from your hand onto the
//	 battlefield. That permanent gains haste, "When this permanent leaves
//	 the battlefield, it deals damage equal to its mana value to each
//	 creature," and "At the beginning of your end step, sacrifice this
//	 permanent.""
//
// Spelunking's put from hand (PutFromHandOntoBattlefield, "you may") with
// its keyed tail, told what entered: only a permanent that really arrived
// is given anything. The gift is Fake Your Own Death's duration grant with
// no stated duration (CR 611.2a), pinned to the new object for as long as
// it stays: haste, and the two abilities as one catalog bundle, so they are
// the permanent's own triggers (its controller controls them, a later
// "loses all abilities" takes them, CR 613.6). The leaves trigger deals
// damage equal to the permanent's mana value as it last existed on the
// battlefield (CR 608.2h, the rulings), with {X} counted as 0, to every
// creature including itself if it still is one; the sacrifice is at the
// beginning of its controller's end step.
//
// No simplification.
const terminalVelocityBundle = "terminal-velocity/leave-and-sacrifice"

func init() {
	Register(Spec{
		OracleID:     "cdda2f04-d59e-4ef8-a26d-964d80750940",
		Name:         "Terminal Velocity",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: terminalVelocityBundle,
			Triggered: []game.TriggeredAbility{
				WhenThisLeaves("Terminal Velocity — damage equal to its mana value to each creature", terminalVelocityLeaves),
				AtYourEndStep("Terminal Velocity — sacrifice this permanent", func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || c.Controller != item.Controller {
						return nil
					}
					return g.SacrificePermanentForEffect(item.SourceCardID)
				}),
			},
			Text: "When this permanent leaves the battlefield, it deals damage equal to its mana value to each creature. At the beginning of your end step, sacrifice this permanent.",
		}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutFromHandOntoBattlefield{
				Match:    Or(Artifact(), Creature()),
				Optional: true,
				Label:    "Terminal Velocity — you may put an artifact or creature card from your hand onto the battlefield",
				ThenTail: terminalVelocityGift,
			}.Apply(ctx)
		},
	})
}

// terminalVelocityLeaves deals damage equal to the leaving permanent's
// mana value, as it last existed, to each creature.
func terminalVelocityLeaves(g *game.Game, item *game.StackItem) error {
	mv := 0
	if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
		mv = info.ManaValue
	} else if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
		mv = c.ManaValue()
	}
	if mv <= 0 {
		return nil
	}
	return damageEachMatching(NewContext(g, item), Creature(), mv)
}

// terminalVelocityGift is the put's tail: haste and the bundle for the
// permanent that entered. An on-disk key.
var terminalVelocityGift = PutFromHandTail("terminal-velocity/gift", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil || !onBattlefield(ctx.Game, res.Entered) {
		return nil
	}
	return GrantAbilitiesFor{
		Target:   res.Entered,
		Keys:     []string{terminalVelocityBundle},
		Also:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), res.Entered),
		Label:    "Terminal Velocity — haste, a leaves trigger and an end step sacrifice",
	}.Apply(ctx)
})
