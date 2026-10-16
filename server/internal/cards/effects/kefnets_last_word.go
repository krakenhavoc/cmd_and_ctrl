package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kefnet's Last Word — Sorcery {2}{U}{U}:
//
//	"Gain control of target artifact, creature, or enchantment. Lands
//	 you control don't untap during your next untap step."
//
// The control change states no duration, so it lasts for the rest of
// the game (CR 611.2a), as Agent of Treachery's does. Then #3995's
// player-wide skip over the caster's lands. With its only target gone
// the spell doesn't resolve (CR 608.2b), so the lands untap as usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67cf6ba1-55ad-4008-ba1f-96c363a96b16",
		Name:         "Kefnet's Last Word",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact, creature, or enchantment", Or(Artifact(), Creature(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GainControl{
					Target:     t.ID,
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Kefnet's Last Word — gain control",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return LandsYouControlDontUntapNextStep(ctx)
		},
	})
}
