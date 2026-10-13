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
// The attack trigger mills three, then asks "you may return a land
// card from your graveyard" as it resolves (b25MillThreeThenReturnLandTapped,
// CR 608.2c): a floor of zero for the "may", over the lands in the
// graveyard AFTER the mill, so a land the trigger itself milled can be
// the one that comes back. Nothing is targeted, so the trigger always
// mills.
//
// The second trigger is Teval's Judgment's condition
// (b16CardLeftYourGraveyard — a move out of the controller's own
// graveyard, or a flashback cast from it) with the per-label "one or
// more" dedup, so the land the attack trigger returns fires it once
// and a Bojuka Bog on the graveyard fires it once. The token is the
// Judgment's 2/2 black Zombie Druid.
//
// The land enters tapped: "tapped" rides its entry event (#1872), so
// nothing sees an untapped land arrive or a land become tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c8cbf0ec-ec98-4cb3-8068-60e92bbd740d",
		Name:            "Teval, the Balanced Scale",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			}, b25TevalAttackLabel, b25MillThreeThenMayReturnLandTapped),
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16CardLeftYourGraveyard(ev, source, g)
			}, b25TevalLeftGraveyardLabel, Do(CreateToken{Template: TokenCard("2/2 black Zombie Druid"), N: 1}))),
		},
	})
}
