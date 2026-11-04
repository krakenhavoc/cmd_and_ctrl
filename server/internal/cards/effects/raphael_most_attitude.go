package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raphael, Most Attitude — Legendary Creature — Mutant Ninja Turtle
// {3}{R}, 4/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Menace
//	 Alliance — Whenever another creature you control enters, you may
//	 exile the top card of your library.
//	 Whenever Raphael attacks, until end of turn, you may play a card
//	 exiled with Raphael."
//
// Alliance is an ability word (CR 207.2c): the trigger is
// WheneverAnotherCreatureEntersUnderYourControl with a CR 603.5 "you
// may". The exile is linked (CR 607.2a): the card carries Raphael as
// the object he is now, so a Raphael that left and came back has exiled
// nothing.
//
// The attack trigger grants a stored permission over exile for the
// rest of the turn, narrowed to the cards linked to this Raphael and
// good for ONE play ("a card": CastsLeft 1, spent by a cast or a land
// play, the land play using a land drop as usual). It covers a card
// exiled after the trigger resolved, the way Urianger Augurelt's
// "cards exiled with" does, and it outlives Raphael leaving the
// battlefield: the cards stay linked to the object he was.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eaf194a7-d95f-4738-9052-92ad00443a76",
		Name:            "Raphael, Most Attitude",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverAnotherCreatureEntersUnderYourControl("Raphael, Most Attitude — Alliance: exile the top card of your library",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Library == nil || p.Library.Size() == 0 {
						return nil
					}
					top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
					return g.ExileLinkedForEffect(src, top)
				}), "Raphael, Most Attitude — exile the top card of your library?"),
			WheneverThisAttacks("Raphael, Most Attitude — until end of turn, you may play a card exiled with Raphael",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					g.GrantCastPermissionForEffect(game.CastPermission{
						Player: item.Controller,
						Zone:   game.ZoneExile,
						Scope:  game.ScopeStanding,
						Filter: game.PermissionFilter{
							ExiledWithSource: true,
							ExiledWith:       game.PermissionCardRef(src),
						},
						CastsLeft: 1,
						Source:    item.SourceCardID,
						Label:     "Raphael, Most Attitude — play a card exiled with Raphael",
					})
					return nil
				}),
		},
	})
}
