package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Annie Flash, the Veteran — Legendary Creature — Human Rogue
// {3}{R}{G}{W}, 4/5 (Outlaws of Thunder Junction, #4468):
//
//	"Flash
//	 When Annie Flash enters, if you cast it, return target permanent
//	 card with mana value 3 or less from your graveyard to the
//	 battlefield tapped.
//	 Whenever Annie Flash becomes tapped, exile the top two cards of
//	 your library. You may play those cards this turn."
//
// "If you cast it" is Geological Appraiser's read of the entering
// permanent's cast provenance (CR 603.4), so an Annie put onto the
// battlefield another way does not trigger. The return is Primary
// Research's graveyard target, any permanent card (a land too), entering
// tapped. "Becomes tapped" is any tap, an attack included
// (ThisBecameTapped, #4534), and the two cards are an ordinary impulse
// exile: "play", so a land among them may be played as the turn's land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "858067f9-eb1f-4569-9831-30441cba5f90",
		Name:            "Annie Flash, the Veteran",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.GeologicalAppraiserWasCast),
					"Annie Flash, the Veteran — return target permanent card with mana value 3 or less from your graveyard to the battlefield tapped",
					effects.ReturnFirstLegalGraveyardTargetToBattlefieldTapped),
				effects.TargetCardInGraveyard("target permanent card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Permanent(), effects.ManaValueLE(3))),
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Annie Flash, the Veteran — exile the top two cards of your library; you may play them this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
