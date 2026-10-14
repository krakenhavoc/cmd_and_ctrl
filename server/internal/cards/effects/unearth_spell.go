package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unearth (the card) — Sorcery {B} (EDHREC top 1,000, #3654):
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Zombify with a mana value cap, plus the cycling ability every cycling
// card carries (cycling.go). The target is legal only while the card
// still has mana value 3 or less in the graveyard (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e0b647a-6e1a-421b-bca9-8d5e346bda80",
		Name:         "Unearth",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", Creature(), ManaValueLE(3), YouOwn()),
		Activated:    []ActivatedAbility{Cycling("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			_, _ = reanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
