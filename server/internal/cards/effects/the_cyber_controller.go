package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Cyber-Controller — Legendary Artifact Creature — Cyberman {X}{U}{U}{B},
// 3/3 (Doctor Who):
//
//	"When The Cyber-Controller enters, each opponent mills X cards. Put all
//	 creature cards milled this way onto the battlefield face down under your
//	 control. They're 2/2 Cyberman artifact creatures.
//	 Other artifact creatures you control get +1/+1."
//
// X is the X the spell was cast with (CR 107.3m), read off the permanent as the
// trigger is built, so it is exact even if the Controller has left by the time
// the trigger resolves. Each opponent mills in turn order, then
// every creature card that really reached a graveyard (a card a replacement
// effect exiled instead was not "milled this way") is returned to the
// battlefield face down under the Controller's control as Missy's and
// Cybership's Cyberman body (CybermanBody, CR 708.2): nameless 2/2 artifact
// creatures that run no enters trigger (CR 708.2a) and that only their
// controller may look at (CR 708.5). The Cybermen are artifact creatures, so
// the anthem reaches them and the Controller's own other artifact creatures.
//
// No simplification beyond one: the cards enter one after another rather than
// all in the same instant.
func init() {
	Register(Spec{
		OracleID:     "06c25d71-f3f4-4089-b9bc-de96fbad73a5",
		Name:         "The Cyber-Controller",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(b27OtherArtifactCreaturesYouControl, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersWithCastX("The Cyber-Controller — each opponent mills X cards; put all creature cards milled this way onto the battlefield face down as Cybermen",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := item.Params.Amount
					if x <= 0 {
						return nil
					}
					var milled []game.Card
					for _, opp := range ctx.Opponents() {
						ids, err := g.MillToZoneForEffect(opp, x, game.ZoneGraveyard)
						if err != nil {
							return err
						}
						for _, id := range ids {
							z := g.FindCardZoneForEffect(id)
							c, ok := g.LookupCardForEffect(id)
							if ok && z != nil && z.Kind == game.ZoneGraveyard && c.IsCreature() {
								milled = append(milled, c)
							}
						}
					}
					for _, c := range milled {
						if z := g.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneGraveyard {
							continue
						}
						if _, err := g.ReturnFromGraveyardFaceDownForEffect(c.InstanceID, item.Controller, false, CybermanBody()); err != nil && !errors.Is(err, game.ErrCardNotFound) {
							return err
						}
					}
					return nil
				}),
		},
	})
}
