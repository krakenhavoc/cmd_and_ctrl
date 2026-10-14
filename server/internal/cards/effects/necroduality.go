package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necroduality — Enchantment {3}{U}:
//
//	"Whenever a nontoken Zombie you control enters, create a token
//	 that's a copy of that creature."
//
// The copy is of the entering OBJECT (the trigger's Object ref), by
// TokenCopyTemplateForObjectForEffect: its copiable values while it is
// still on the battlefield as the same object, and its last-known ones
// once it has left (CR 707.2, CR 608.2h). The token is a token, so it
// does not trigger Necroduality again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d36acdfe-2b1b-4c19-8b81-0b3127e480fe",
		Name:         "Necroduality",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, nontokenZombieEnteredUnderYourControl,
				"Necroduality — create a token that's a copy of that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					obj := ctx.Trigger().Object
					if obj == nil {
						return nil
					}
					tmpl, ok := g.TokenCopyTemplateForObjectForEffect(obj.Ref())
					if !ok {
						return nil
					}
					return CreateToken{Template: tmpl, N: 1}.Apply(ctx)
				}),
		},
	})
}

// nontokenZombieEnteredUnderYourControl is "whenever a nontoken Zombie
// you control enters".
func nontokenZombieEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && !c.IsToken() && c.HasSubtype("Zombie")
}
