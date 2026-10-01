package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blessed Reincarnation — Instant {3}{U}:
//
//	"Exile target creature an opponent controls. That player reveals
//	 cards from the top of their library until a creature card is
//	 revealed. The player puts that card onto the battlefield, then
//	 shuffles the rest into their library.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// "That player" is the creature's controller as the spell resolves,
// read before the exile. The reveal runs from the exile's continuation,
// whether or not the creature actually reached exile (a commander that
// took the command zone still left): the reveal is not an "if you do".
// A library with no creature card reveals itself entirely, nothing
// enters, and every revealed card is shuffled back in. The creature
// enters under that player's control, since they put it there.
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca29588c-f117-418e-be9e-fa2ee89862ca",
		Name:            "Blessed Reincarnation",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			player, ok := controllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			return ExileTarget{Target: id, Then: func(ctx *Context, _ bool) error {
				return blessedReincarnationReveal(ctx, player)
			}}.Apply(ctx)
		},
	})
}

// blessedReincarnationReveal is the second and third sentences: reveal
// until a creature card, put it onto the battlefield, shuffle the rest
// in.
func blessedReincarnationReveal(ctx *Context, player uuid.UUID) error {
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		"Blessed Reincarnation — reveal until a creature card")
	// The shuffle rides Then, so an entry that pauses for its own
	// question (a Clone's copy) is on the battlefield before the rest
	// is shuffled in.
	return PutFromLibraryOntoBattlefield{
		Player: player,
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:    true,
		Then: func(g *game.Game, _ PutFromLibraryResult) error {
			return g.ShuffleLibraryForEffect(player)
		},
	}.Apply(ctx)
}
