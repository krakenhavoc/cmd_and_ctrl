package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mishra, Claimed by Gix — Legendary Creature — Phyrexian Human Artificer
// {2}{B}{R}, 3/5 (The Brothers' War, #5482):
//
//	"Whenever you attack, each opponent loses X life and you gain X life,
//	 where X is the number of attacking creatures. If Mishra, Claimed by
//	 Gix and a creature named Phyrexian Dragon Engine are attacking, and
//	 you both own and control them, exile them, then meld them into
//	 Mishra, Lost to Phyrexia. It enters tapped and attacking."
//
// "Whenever you attack" is one trigger per attack declaration
// (OncePerBatch). X is counted as the trigger resolves, over every
// attacking creature whoever controls it. The meld sentence is not an
// intervening "if": it is read as the ability resolves, after the life is
// lost and gained, and does nothing unless Mishra and a creature named
// Phyrexian Dragon Engine are both still attacking and both owned and
// controlled by the ability's controller (CR 701.42a). A Clone named
// Phyrexian Dragon Engine is exiled but can't be melded (CR 701.42c). The
// melded permanent's "enters tapped and attacking" is the entry replacement
// on Mishra, Lost to Phyrexia, which asks what it attacks (CR 508.4).
func init() {
	Register(Spec{
		OracleID:     "26a623a1-60c9-47e3-af96-616e29551865",
		Name:         "Mishra, Claimed by Gix",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Mishra, Claimed by Gix — each opponent loses X life and you gain X life; meld with Phyrexian Dragon Engine",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x := s82BroB20AttackingCreatures(g)
					if x > 0 {
						if err := (EachOpponentLosesLife{N: x}).Apply(ctx); err != nil {
							return err
						}
						if err := (GainLife{Player: ctx.Controller(), Amount: x}).Apply(ctx); err != nil {
							return err
						}
					}
					partner := s82BroB20AttackingPartner(g, ctx.Source(), ctx.Controller(), "Phyrexian Dragon Engine")
					if partner == uuid.Nil {
						return nil
					}
					return g.MeldForEffect(ctx.Source(), partner, ctx.Controller())
				})),
		},
	})
}
