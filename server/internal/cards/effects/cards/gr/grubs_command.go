package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grub's Command — Kindred Sorcery — Goblin {3}{B}{R} (Lorwyn Eclipsed,
// #4458):
//
//	"Choose two —
//	 • Create a token that's a copy of target Goblin you control.
//	 • Creatures target player controls get +1/+1 and gain haste until
//	   end of turn.
//	 • Destroy target artifact or creature.
//	 • Target player mills five cards, then puts each Goblin card milled
//	   this way into their hand."
//
// Ashling's Command's "choose two" template (ChooseN 2..2): each chosen
// bullet has its own target, checked again as the spell resolves (CR
// 608.2b), and the bullets run in printed order (CR 608.2c). The copy
// bullet is KindredCommandCopyMode. The pump fixes its set of creatures
// as it resolves (CR 611.2c), Goblin War Party's +1/+1 and haste over
// the target player's creatures. The mill is Cantankerous Keepers': the
// clause after it runs on the cards that actually reached that player's
// graveyard (CR 701.17a), and every Goblin card among them still there
// goes to its owner's hand, with no prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8173e160-9a22-43f7-90d1-9ada043bc1b7",
		Name:         "Grub's Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.KindredCommandCopyMode("Goblin"),
			effects.ModeWithPurpose(effects.ModeDoing("Creatures target player controls get +1/+1 and gain haste until end of turn.",
				effects.TargetPlayer("target player"), grubsCommandPumpAndHaste),
				game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant}),
			effects.ModeWithPurpose(effects.ModeDoing("Destroy target artifact or creature.",
				effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
				effects.DestroyTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.ModeDoing("Target player mills five cards, then puts each Goblin card milled this way into their hand.",
				effects.TargetPlayer("target player"), grubsCommandMill),
				game.Purpose{Answers: game.AnswerValue}),
		),
	})
}

// grubsCommandPumpAndHaste is "creatures target player controls get
// +1/+1 and gain haste until end of turn".
func grubsCommandPumpAndHaste(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	theirs := effects.And(effects.Creature(), effects.ControlledByPlayer(t.ID))
	if err := (effects.BoostUntilEOT{Match: theirs, Power: 1, Toughness: 1, Label: "Grub's Command — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Match: theirs, Keywords: []string{"haste"}, Label: "Grub's Command — haste"}.Apply(ctx)
}

// grubsCommandMill mills five cards from the target player's library.
func grubsCommandMill(_ *game.StackItem, ctx *effects.Context, occ int) error {
	t, ok := effects.ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return effects.MillToZone{Player: t.ID, N: 5, ThenTail: ctx.ItemTail(grubsCommandMilled)}.Apply(ctx)
}

// grubsCommandMilled is the clause after the mill: every milled Goblin
// card still in the graveyard goes to its owner's hand. An on-disk key.
var grubsCommandMilled = effects.CardTail("grubs-command/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.HobA11AllMilledToHand(ctx, o, effects.HasSubtype(ctx.TextSubtype("Goblin")))
})
