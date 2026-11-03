package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Killian's Confidence — Sorcery {W}{B} (Secrets of Strixhaven, #4456):
//
//	"Target creature gets +1/+1 until end of turn. Draw a card.
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, you may pay {W/B}. If you do, return this card from your
//	 graveyard to your hand."
//
// The spell pumps the target and draws, in that order; it draws only if it
// resolves at all (a spell whose only target is illegal does nothing, CR
// 608.2b). The second paragraph is a triggered ability that works only from
// the graveyard (InGraveyard, the Silvan Reveler shape): one trigger per
// player the controller's creatures connect with in a combat damage step
// (CR 603.2c), whatever their number, and "you" there is the card's owner
// (CR 108.4). The {W/B} is asked as the ability resolves, and the card
// comes back only if it is still the very object that triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c17f0e65-e5a3-4113-b782-0b192ab52046",
		Name:         "Killian's Confidence",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 1, 1))},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Killian's Confidence — +1/+1 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Killian's Confidence — you may pay {W/B} to return it to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return mayPayForGraveyardReturnToHand(g, item, "{W/B}", "Killian's Confidence — pay {W/B} to return it to your hand?")
				})),
		},
	})
}
