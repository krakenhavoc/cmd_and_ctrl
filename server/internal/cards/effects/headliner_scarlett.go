package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Headliner Scarlett — Legendary Creature — Human Warlock {3}{R}, 3/3
// (Ravnica: Clue Edition, #5929):
//
//	"Haste
//	 When Headliner Scarlett enters, creatures target player controls can't
//	 block this turn.
//	 At the beginning of your upkeep, exile the top card of your library face
//	 down. You may look at and play that card this turn."
//
// "Can't block" is a rules effect, so CR 611.2c does not lock its set: the
// scope reads the target player's creatures live until cleanup
// (ScopeYourCreatures registered with the target as its "you", as Incite War
// does), and a creature that player gets later can't block either. The
// upkeep trigger is Gonti's face-down impulse exile (ExileTopWithPermission,
// FaceDown): only the controller may look at the card, a land is played, a
// spell cast, and the permission lapses at end of turn, leaving the card in
// exile.
//
// No simplification.
func init() {
	const enterLabel = "Headliner Scarlett — creatures target player controls can't block this turn"
	Register(Spec{
		OracleID:        "f5420901-b5e6-4d54-b3e8-5f45e37f1fab",
		Name:            "Headliner Scarlett",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(enterLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				p := firstLegalTargetPlayer(ctx)
				if p == [16]byte{} {
					return nil
				}
				g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, p,
					[]game.Mod{game.AddRestrictionsMod(game.CantBlock)}, DurationUntilEndOfTurn(ctx), enterLabel)
				return nil
			}), TargetPlayer("target player")),
			AtYourUpkeep("Headliner Scarlett — exile the top card of your library face down; you may look at and play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1, FaceDown: true}.Apply(NewContext(g, item))
				}),
		},
	})
}
