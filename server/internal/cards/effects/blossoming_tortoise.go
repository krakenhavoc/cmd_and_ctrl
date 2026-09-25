package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossoming Tortoise — Creature — Turtle {2}{G}{G}, 3/3 (EDHREC rank
// 3278):
//
//	"Whenever this creature enters or attacks, mill three cards, then
//	 return a land card from your graveyard to the battlefield tapped.
//	 Activated abilities of lands you control cost {1} less to
//	 activate.
//	 Land creatures you control get +1/+1."
//
// The lands deck's engine: a land back from the graveyard on entry
// and on every attack. The trigger is ONE printed ability with two
// trigger conditions (b21SelfEnteredOrAttacked) and Teval's body
// (b25MillThreeThenReturnChosenLandTapped). The two used to differ
// only in a clamp that bounded the mill by the library; a mill never
// loses the game (CR 701.17b, #767), so the clamp went and the bodies
// became one.
//
// "Then return a land card from your graveyard" is written as a
// target clause over the controller's graveyard (Teval's posture),
// and — because the printed return is not optional — it is a
// mandatory single pick whenever the graveyard holds a land. It is
// declared TWICE for the reason Hazel's Brewmaster gives: the engine
// drops a targeted trigger whose legal set is empty, which would
// drop the mill with it, so the targeted entry fires only while the
// graveyard holds a land card and the untargeted mill-only entry
// fires only when it does not. Exactly one applies to any entry or
// attack.
//
// The anthem is a layer 7c modify over land creatures the
// controller controls — post-layer types, so an animated land
// counts.
//
// Three declared simplifications, all weaker than printed:
//
//   - The land to return is picked when the trigger goes on the
//     stack, before the three cards are milled, so a land milled by
//     the trigger itself cannot be the one returned and opponents
//     see the pick before it resolves. A pick that left the
//     graveyard in response counters the trigger (CR 608.2b) and the
//     mill does not happen either.
//   - The returned land enters untapped and is tapped a beat later
//     inside the same resolution (Lumra's gap — the graveyard return
//     path has no tapped flag), so anything watching for a land
//     being tapped sees one.
//   - The activated-ability discount is not implemented. The
//     engine's cost modifiers price SPELLS at cast (CR 601.2f) and
//     nothing prices an activation, so a land's ability costs its
//     printed amount. The Tortoise is still the mill-and-return body
//     and the anthem, which is what it is played for.
func init() {
	Register(Spec{
		OracleID:     "5e1bf23b-7fb0-45ff-8544-fce9fa3eba00",
		Name:         "Blossoming Tortoise",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The land to return is picked when the trigger goes on the stack, before the three cards are milled — so a land milled by the trigger itself can't be the one that comes back.",
			"The returned land enters untapped and is tapped immediately afterwards, so anything watching for a land being tapped sees one.",
			"Activated abilities of lands you control don't cost {1} less — the discount isn't implemented.",
		},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b31LandCreaturesYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB, game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b21SelfEnteredOrAttacked(ev, source) && b25GraveyardHasLandCard(g, source.Controller)
				},
				Targets: TargetCardInGraveyard("a land card in your graveyard to return tapped", YouOwn(), Land()),
				Key:     b31BlossomingTortoiseLabel,
				Effect:  b25MillThreeThenReturnChosenLandTapped,
			},
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21SelfEnteredOrAttacked(ev, source) && !b25GraveyardHasLandCard(g, source.Controller)
			}, b31BlossomingTortoiseLabel, b25MillThreeThenReturnChosenLandTapped),
		},
	})
}
