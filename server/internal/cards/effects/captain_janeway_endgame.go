package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Janeway, Endgame — Legendary Creature — Human Officer
// {1}{U}{R}{W}, 3/3 (Star Trek Commander, #4974):
//
//	"Whenever one or more Officers you control deal combat damage to a
//	 player, you draw a card and create a 1/1 red Officer creature token.
//	 {2}, Tap an untapped creature you control: Captain Janeway can't be
//	 blocked this turn."
//
// The trigger is Grazilaxx's batched shape narrowed to Officers, read off
// each damage source through the layers: once per player your Officers
// connect with in a combat damage step (CR 603.2c), however many of them
// do. The token is the set's printed 1/1 red Officer (Captain Jean-Luc
// Picard's). The activated ability's tap is a cost with no "another"
// (TapAnUntapped), so Janeway may tap herself to pay it, and it is not the
// {T} symbol, so a creature that arrived this turn may be tapped; the
// restriction lands on Janeway only while she is still the permanent
// that activated it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "488ff0d9-a74c-4ba9-9c10-3aa88230c2bd",
		Name:         "Captain Janeway, Endgame",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(HasSubtype("Officer"),
				"Captain Janeway, Endgame — you draw a card and create a 1/1 red Officer creature token",
				Do(DrawCards{N: 1}, CreateToken{Template: TokenCard("1/1 red Officer"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Tap an untapped creature you control: Captain Janeway can't be blocked this turn.",
			Cost:    Plus(ManaCost("{2}"), TapAnUntapped("an untapped creature you control", Creature())),
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Captain Janeway, Endgame — can't be blocked"}.Apply(NewContext(g, item))
			},
		}},
	})
}
