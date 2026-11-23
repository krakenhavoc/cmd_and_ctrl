package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firbolg Flutist — Creature — Giant Bard {4}{R}{R}, 4/4 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Enthralling Performance — When this creature enters, gain control of target
//	 creature you don't control until end of turn. Untap it. It gains haste and
//	 myriad until end of turn. (Whenever it attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and attacking that
//	 player or a planeswalker they control. Exile the tokens at end of combat.)"
//
// Dominating Vampire's control change (until end of turn, CR 611.2), untap and haste,
// with Ironwill Forger's myriad grant (GrantedMyriad in an AbilityGrant, given
// until end of turn): the stolen creature's controller, you, makes the token copies
// when it attacks, one per opponent other than its defending player. Each effect is
// pinned to the one object chosen (CR 400.7) and the target is re-checked as the
// trigger resolves (CR 608.2b). "Enthralling Performance" is an ability word with no
// rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	const myriadGrant = "firbolg-flutist/myriad"
	effects.Register(effects.Spec{
		OracleID:     "5bf88cac-5997-4254-8349-2473de854c50",
		Name:         "Firbolg Flutist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       myriadGrant,
			Text:      "Myriad",
			Triggered: []game.TriggeredAbility{effects.GrantedMyriad("Firbolg Flutist")},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Firbolg Flutist — gain control of target creature you don't control until end of turn; untap it; it gains haste and myriad until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.GainControl{
						Target:   id,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Firbolg Flutist — gain control until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Firbolg Flutist — haste until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantAbilitiesFor{Target: id, Keys: []string{myriadGrant}, Label: "Firbolg Flutist — myriad until end of turn"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
		},
	})
}
