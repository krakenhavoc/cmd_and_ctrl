package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tamiyo's Safekeeping — Instant for {G}:
//
//	"Target permanent you control gains hexproof and indestructible
//	 until end of turn. You gain 2 life."
//
// Heroic Intervention for one permanent at one mana — and unlike
// every other card in this family, it says PERMANENT rather than
// creature, so it answers a Vindicate aimed at your mana rock or a
// Beast Within aimed at your land.
//
// Both keywords ride one registry entry (both are layer-6 ability
// additions); the life gain is an ordinary one-shot with no duration
// at all. The order matters not at all here, but the grant goes
// first so that a bug in it cannot be masked by a life total that
// moved anyway.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "bb2b324c-970a-4920-884e-c92ba49669f0",
		Name:         "Tamiyo's Safekeeping",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (GrantKeywordUntilEOT{
					Target:   item.Targets[0].ID,
					Keywords: []string{"hexproof", "indestructible"},
					Label:    "Tamiyo's Safekeeping — hexproof and indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			// The life gain is NOT conditional on the target: it is a
			// second sentence, so a spell that resolves with its
			// target gone still gains the 2. (A spell whose ONLY
			// target is illegal is countered on resolution and never
			// reaches this function at all — that is the rules
			// engine's job, not this card's.)
			return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
