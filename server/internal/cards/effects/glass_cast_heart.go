package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glass-Cast Heart — Artifact {2}{B} (Commander, #5368):
//
//	"Whenever one or more Vampires you control attack, create a Blood
//	 token.
//	 {B}, {T}, Pay 1 life: Create a 1/1 white and black Vampire creature
//	 token with lifelink.
//	 {B}{B}, {T}, Sacrifice this artifact and thirteen Blood tokens: Each
//	 opponent loses 13 life and you gain 13 life."
//
// The attack trigger is General Kreat's shape: once per declaration
// (OncePerBatch) however many Vampires attack, for attackers read
// through the layers (a changeling is a Vampire). The second ability's
// costs are mana, {T} and life (CR 119.4). The third sacrifices the
// Heart and thirteen Blood tokens as costs (CR 602.2b): the cost is
// refused with fewer than thirteen Blood tokens, and the drain is
// unconditional once paid — every opponent loses 13, then you gain 13.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3edd6272-20fa-4c3a-bf34-2f253cc8ac9e",
		Name:         "Glass-Cast Heart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21CreatureOfSubtypeYouControlAttacked(ev, source, g, "Vampire")
			}, "Glass-Cast Heart — create a Blood token",
				Do(CreateToken{Template: BloodToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{B}, {T}, Pay 1 life: Create a 1/1 white and black Vampire creature token with lifelink.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Cost:    Plus(ManaCost("{B}"), TapCost(), PayLife(1)),
				Effect:  Do(CreateToken{Template: TokenCard("1/1 white and black Vampire with lifelink"), N: 1}),
			},
			{
				Label:   "{B}{B}, {T}, Sacrifice this artifact and thirteen Blood tokens: Each opponent loses 13 life and you gain 13 life.",
				Purpose: game.Purpose{LifeGain: 13},
				Cost: Plus(ManaCost("{B}{B}"), TapCost(), SacrificeThis(),
					SacrificeN(13, "thirteen Blood tokens", HasSubtype("Blood"))),
				Effect: Do(EachOpponentLosesLife{N: 13}, GainLife{Amount: 13}),
			},
		},
	})
}
