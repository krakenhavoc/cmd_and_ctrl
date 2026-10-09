package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jace, Multiverse Architect — Legendary Planeswalker — Jace
// {1}{W}{U}{B}{R}, loyalty 4 (Reality Fracture commander set):
//
//	"At the beginning of combat on each opponent's turn, they may pay
//	 {2}. If they don't, creatures they control can't attack Jaces you
//	 control this turn.
//	 +1: Draw two cards, then put a card from your hand on the bottom of
//	     your library.
//	 −3: Exile another target planeswalker or creature you control.
//	     Reveal cards from the top of your library until you reveal a
//	     creature or planeswalker card. Put that card onto the battlefield
//	     and the rest on the bottom of your library in a random order.
//	 Jace, Multiverse Architect can be your commander."
//
// DECLARED SIMPLIFICATION, weaker than printed: the combat tax is not
// built. "Creatures they control can't attack Jaces you control this
// turn" needs a turn-long attack restriction on a player, scoped to one
// subtype of planeswalker. The only player-scoped restriction
// (GrantCantAttackPlayerForEffect) protects the whole player and every
// permanent they control for a later turn, which would be stronger than
// printed; a creature-scoped one is a static of that creature. Leaving
// the tax out makes Jaces easier to attack, never harder. The +1, the −3
// and commander eligibility are all as printed.
func init() {
	Register(Spec{
		OracleID:     "3321c134-bf5b-4f63-8035-fca0bbdfa86b",
		Name:         "Jace, Multiverse Architect",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The combat tax isn't implemented — opponents' creatures can attack your Jaces without anyone paying {2}."},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Draw two cards, then put a card from your hand on the bottom of your library.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
						return err
					}
					return rfPutCardFromHandOnBottom(ctx, "Jace, Multiverse Architect — put a card from your hand on the bottom of your library")
				},
			},
			{
				Label:   "−3: Exile another target planeswalker or creature you control. Reveal cards from the top of your library until you reveal a creature or planeswalker card. Put that card onto the battlefield and the rest on the bottom of your library in a random order.",
				Cost:    LoyaltyCost(-3),
				Targets: Another(TargetPermanent("another target planeswalker or creature you control", Or(Creature(), Planeswalker()), YouControl())),
				Effect:  rfExileTargetThenRevealCreatureOrPlaneswalker,
			},
		},
	})
}
