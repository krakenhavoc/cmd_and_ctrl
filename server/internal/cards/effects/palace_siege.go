package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Palace Siege — Enchantment {3}{B}{B}:
//
//	"As this enchantment enters, choose Khans or Dragons.
//	 • Khans — At the beginning of your upkeep, return target creature
//	   card from your graveyard to your hand.
//	 • Dragons — At the beginning of your upkeep, each opponent loses
//	   2 life and you gain 2 life."
//
// A #1572 anchor-word card: each bullet is an upkeep trigger gated on
// its word (ADR 0071), so exactly one of them exists once the choice
// is made.
//
// Khans targets (CR 603.3d): a graveyard with no creature card drops
// the trigger with no prompt, and a target that leaves in response is
// skipped on resolution (CR 608.2b). Dragons gains 2 once, however
// many opponents lost life — "you gain 2 life" is not per opponent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d819d35-5cd3-48f1-a77a-b7c3ff82c62e",
		Name:         "Palace Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Palace Siege", "Khans", "Dragons"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Khans", Targeting(
				AtYourUpkeep("Palace Siege — return target creature card to your hand", palaceSiegeReturn),
				TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()))),
			WhenChosen("Dragons", AtYourUpkeep("Palace Siege — each opponent loses 2 life and you gain 2 life",
				palaceSiegeDrain)),
		},
	})
}

// palaceSiegeReturn is the Khans body: the chosen creature card, if it
// is still a legal target, goes to its owner's hand.
func palaceSiegeReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
	}
	return nil
}

// palaceSiegeDrain is the Dragons body.
func palaceSiegeDrain(g *game.Game, item *game.StackItem) error {
	if err := eachOpponentLosesLife(g, item, 2); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
}
