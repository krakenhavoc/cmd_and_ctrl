package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadly Dispute — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact
//	 or creature. Draw two cards and create a Treasure token."
//
// The Treasure is what separates this from Altar's Reap: it refunds
// one of the two mana, so eating a Treasure to cast it is mana-neutral
// and eating a Clue or a token creature is close to free. That is why
// artifact-heavy lists play it over the strictly cheaper Village
// Rites — the fodder is already on the board.
//
// "An artifact or creature" is a wider clause than the other two, and
// the reason SacrificeCost takes predicates rather than a boolean.
func init() {
	Register(Spec{
		OracleID:       "457af74a-02b3-4659-846d-63e482667f34",
		Name:           "Deadly Dispute",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TreasureToken(),
				N:          1,
			}.Apply(ctx)
		},
	})
}
