package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venser, Corpse Puppet — Legendary Creature — Phyrexian Zombie Wizard
// {U}{B}, 1/3 (Phyrexia: All Will Be One, #5653):
//
//	"Lifelink, toxic 1
//	 Whenever you proliferate, choose one —
//	 • If you don't control a creature named The Hollow Sentinel, create
//	   The Hollow Sentinel, a legendary 3/3 colorless Phyrexian Golem
//	   artifact creature token.
//	 • Target artifact creature you control gains flying and lifelink
//	   until end of turn."
//
// One modal trigger off WheneverYouProliferate. The mode is chosen as
// the trigger is put on the stack (CR 603.3c), and the second is offered
// only with an artifact creature to target. The first bullet's "if" is
// part of its instruction, not an intervening "if": the bullet may be
// chosen with a Hollow Sentinel on the battlefield, and it is read only
// as the trigger resolves, by name and through the layers. The second
// does nothing if its target is illegal by then (CR 608.2b).
//
// No simplification.
func init() {
	venser := WheneverYouProliferate("Venser, Corpse Puppet — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	venser.Modes = ChooseOne(
		ModeDoing("If you don't control a creature named The Hollow Sentinel, create The Hollow Sentinel, a legendary 3/3 colorless Phyrexian Golem artifact creature token.",
			nil, func(item *game.StackItem, ctx *Context, _ int) error {
				if !youControlNoCreatureNamed(ctx.Game, item.Controller, "The Hollow Sentinel") {
					return nil
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("3/3 colorless Legendary Phyrexian Golem artifact named The Hollow Sentinel"),
					N:          1,
				}.Apply(ctx)
			}),
		ModeDoing("Target artifact creature you control gains flying and lifelink until end of turn.",
			TargetCreature("target artifact creature you control", Artifact(), YouControl()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"flying", "lifelink"},
					Label:    "Venser, Corpse Puppet — flying and lifelink until end of turn",
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "17c934a7-3480-4309-a217-ca2550538ae7",
		Name:            "Venser, Corpse Puppet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "toxic 1"},
		Triggered:       []game.TriggeredAbility{venser},
	})
}
