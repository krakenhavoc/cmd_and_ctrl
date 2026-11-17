package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bladecoil Serpent — Artifact Creature — Serpent {X}{6}, 5/4 (The
// Brothers' War):
//
//	"When this creature enters, for each {U}{U} spent to cast it, draw a
//	 card.
//	 When this creature enters, for each {B}{B} spent to cast it, each
//	 opponent discards a card.
//	 When this creature enters, for each {R}{R} spent to cast it, it gets
//	 +1/+0 and gains trample and haste until end of turn."
//
// Three enters triggers, each counting pairs of that colour in the mana
// that paid for the spell (WhenThisEntersForEachPairSpent): four blue
// mana is two cards. The count is fixed as the trigger is built, so it is
// exact if the Serpent leaves first. The red trigger touches only the
// permanent that triggered it (CR 400.7).
func init() {
	const (
		drawLabel    = "Bladecoil Serpent — for each {U}{U} spent, draw a card"
		discardLabel = "Bladecoil Serpent — for each {B}{B} spent, each opponent discards a card"
		pumpLabel    = "Bladecoil Serpent — for each {R}{R} spent, +1/+0 and trample and haste until end of turn"
	)
	Register(Spec{
		OracleID:       "e469bb22-57f1-4555-b3c0-9ea108e6cc12",
		Name:           "Bladecoil Serpent",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersForEachPairSpent("U", drawLabel, Do(drawPerPairSpent{})),
			WhenThisEntersForEachPairSpent("B", discardLabel, func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				for _, opp := range NewContext(g, item).Opponents() {
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player: opp,
						Source: item.SourceCardID,
						N:      item.Params.Amount,
					})
				}
				return nil
			}),
			WhenThisEntersForEachPairSpent("R", pumpLabel, func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 || !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: item.SourceCardID, Power: item.Params.Amount,
					Label: "Bladecoil Serpent — +1/+0 for each {R}{R}"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample", "haste"},
					Label: "Bladecoil Serpent — trample and haste"}.Apply(ctx)
			}),
		},
	})
}
