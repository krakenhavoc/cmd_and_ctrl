package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ukkima, Stalking Shadow — Legendary Creature — Whale Wolf {1}{U}{B},
// 2/2 (Outlaws of Thunder Junction Commander, #5368):
//
//	"Partner with Cazur, Ruthless Stalker (When this creature enters,
//	 target player may put Cazur into their hand from their library, then
//	 shuffle.)
//	 Ukkima can't be blocked.
//	 When Ukkima leaves the battlefield, it deals X damage to target
//	 player and you gain X life, where X is its power."
//
// "Can't be blocked" is the self restriction (Blighted Agent). The leave
// trigger reads X from Ukkima's last-known power (CR 603.10a, so counters
// and pumps count) and is damage from Ukkima as it last existed
// (ctx.SourceRef, CR 608.2h). The target player is chosen as the trigger
// goes on the stack; if that player is no longer a legal target nothing
// happens, including the life gain (CR 608.2b). With X of zero or less it
// deals no damage and gains no life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b8c3cb84-0a30-420d-8fc3-bb85ce69b5e8",
		Name:         "Ukkima, Stalking Shadow",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Ukkima, Stalking Shadow", "Cazur, Ruthless Stalker"),
			Targeting(WhenThisLeaves("Ukkima, Stalking Shadow — it deals X damage to target player and you gain X life, where X is its power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player, ok := firstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					info, ok := ctx.SourcePermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					x := info.Power
					d := DealDamage{Source: ctx.Source(), Target: player, Amount: x}
					if ref, hasRef := ctx.SourceRef(); hasRef {
						d = DealDamage{SourceObject: &ref, Target: player, Amount: x}
					}
					if err := d.Apply(ctx); err != nil {
						return err
					}
					return GainLife{Amount: x}.Apply(ctx)
				}),
				TargetPlayer("target player")),
		},
	})
}
