package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teval, the Balanced Scale — Legendary Creature — Spirit Dragon
// {1}{B}{G}{U}, 4/4 (EDHREC rank 2675):
//
//	"Flying
//	 Whenever Teval attacks, mill three cards. Then you may return a
//	 land card from your graveyard to the battlefield tapped.
//	 Whenever one or more cards leave your graveyard, create a 2/2
//	 black Zombie Druid creature token."
//
// The Duskmourn self-mill commander whose own attack trigger feeds
// its own token trigger. Flying rides PrintedKeywords.
//
// The attack trigger mills three, then returns the chosen land. Its
// "you may return a land card from your graveyard" is written as an
// "up to one" target clause over the controller's graveyard (Grapple
// with the Past's posture), declared TWICE for the reason Hazel's
// Brewmaster gives: the engine drops a targeted trigger whose legal
// set is empty, which would drop the mill with it, so the targeted
// entry fires only while the graveyard holds a land card and the
// untargeted mill-only entry fires only when it does not. Exactly
// one applies to any attack.
//
// The second trigger is Teval's Judgment's condition
// (b16CardLeftYourGraveyard — a move out of the controller's own
// graveyard, or a flashback cast from it) with the per-label "one or
// more" dedup, so the land the attack trigger returns fires it once
// and a Bojuka Bog on the graveyard fires it once. The token is the
// Judgment's 2/2 black Zombie Druid.
//
// Two declared simplifications, both weaker than printed:
//
//   - The land to return is picked when the attack trigger goes on
//     the stack, before the three cards are milled, so a land milled
//     by the trigger itself cannot be the one returned and opponents
//     see the pick before it resolves. A pick that left the
//     graveyard in response counters the trigger (CR 608.2b) and the
//     mill does not happen either; a pick of no land simply mills.
//   - The returned land enters untapped and is tapped a beat later
//     inside the same resolution (Lumra's gap — the graveyard return
//     path has no tapped flag), so anything watching for a land
//     being tapped sees one.
func init() {
	Register(Spec{
		OracleID:     "c8cbf0ec-ec98-4cb3-8068-60e92bbd740d",
		Name:         "Teval, the Balanced Scale",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The land to return is picked when the attack trigger goes on the stack, before the three cards are milled — so a land milled by the trigger itself can't be the one that comes back.",
			"The returned land enters untapped and is tapped immediately afterwards, so anything watching for a land being tapped sees one.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return attackDeclared(ev, source) && b25GraveyardHasLandCard(g, source.Controller)
				},
				Targets: TargetCardInGraveyard("up to one land card in your graveyard to return tapped", YouOwn(), Land()).WithCount(0, 1),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b25TevalAttackLabel, b25MillThreeThenReturnChosenLandTapped)
				},
			},
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && !b25GraveyardHasLandCard(g, source.Controller)
			}, b25TevalAttackLabel, b25MillThreeThenReturnChosenLandTapped),
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16CardLeftYourGraveyard(ev, source, g)
			}, b25TevalLeftGraveyardLabel, Do(CreateToken{Template: TokenCard("2/2 black Zombie Druid"), N: 1}))),
		},
	})
}
