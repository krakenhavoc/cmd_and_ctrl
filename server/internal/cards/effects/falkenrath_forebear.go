package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falkenrath Forebear — Creature — Vampire {2}{B}, 3/1 (Innistrad:
// Crimson Vow, #5495):
//
//	"Flying
//	 This creature can't block.
//	 Whenever this creature deals combat damage to a player, create a Blood
//	 token.
//	 {B}, Sacrifice two Blood tokens: Return this card from your graveyard
//	 to the battlefield."
//
// The last ability works from the graveyard only (Cauldron Familiar's
// shape); its cost is {B} and two Blood tokens, refused with fewer. A
// Forebear that has left the graveyard by resolution is not returned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "940080ce-c504-48c8-a763-326fef907217",
		Name:            "Falkenrath Forebear",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Falkenrath Forebear — create a Blood token",
				Do(CreateToken{Template: BloodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice two Blood tokens: Return this card from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{B}"), SacrificeN(2, "two Blood tokens", HasSubtype("Blood"))),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnThisFromGraveyardToBattlefield(g, item)
			},
		}},
	})
}
