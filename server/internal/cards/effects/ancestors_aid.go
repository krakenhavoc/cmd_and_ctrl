package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestors' Aid — Instant {1}{R} (EDHREC rank 4357):
//
//	"Target creature gets +2/+0 and gains first strike until end of
//	 turn.
//	 Create a Treasure token."
//
// A combat trick that leaves a mana behind. The Treasure is why it
// plays in Commander at all: a two-mana trick that refunds one of
// them is a two-mana trick you cast on somebody else's turn without
// falling behind. Roadmap batch 42 (#449) files it under
// until-end-of-turn, shipped by #279.
//
// Two turn-scoped statics rather than one: the +2/+0 is a Layer 7c
// modification and first strike is a Layer 6 ability grant, and CR
// 613.1 puts them in different layers even though one sentence
// printed both.
//
// The Treasure is unconditional — it is a separate sentence, so a
// resolving Ancestors' Aid makes one even if the creature it was
// aimed at is somehow no longer there to pump. (A spell whose only
// target is illegal is countered on resolution and makes nothing;
// that is CR 608.2b doing its job before this code runs.)
//
// No simplification. The Treasure is the real token with its real
// mana ability, so it can be cracked for any colour in the usual way.
func init() {
	Register(Spec{
		OracleID:     "a95a0b13-d50f-41cf-8668-de60d445e7b0",
		Name:         "Ancestors' Aid",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				target := item.Targets[0].ID
				if err := (BoostUntilEOT{
					Target: target,
					Power:  2,
					Label:  "Ancestors' Aid — +2/+0 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
					Target:   target,
					Keywords: []string{"first strike"},
					Label:    "Ancestors' Aid — first strike until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
