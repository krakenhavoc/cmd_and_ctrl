package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirkwood Elk — Creature — Elk {5}{G}, 6/6:
//
//	"Trample
//	 Whenever this creature enters or attacks, return target Elf card
//	 from your graveyard to your hand. You gain life equal to that
//	 card's power."
//
// Nadaar's one ability with two trigger conditions (WhenThisEntersOrAttacks)
// and Restore with Empathy's graveyard-to-hand body. A trigger whose one
// target has gone does nothing at all (CR 608.2b), so no life is gained
// then. The life is the card's power as it sat in the graveyard, read
// before it moves (a `*` power counts what the card's own
// characteristic-defining ability says, CR 113.6a, via
// PowerAnywhereForEffect); a negative power gains nothing.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEntersOrAttacks("Mirkwood Elk — return target Elf card from your graveyard to your hand and gain life equal to its power",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			life := 0
			if c, ok := g.LookupCardForEffect(ids[0]); ok {
				life = g.PowerAnywhereForEffect(c)
			}
			if err := effects.ReturnTargetGraveyardCardToHand(g, item); err != nil {
				return err
			}
			if life <= 0 {
				return nil
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: life}.Apply(ctx)
		})
	trigger.Targets = effects.TargetCardInGraveyard("target Elf card in your graveyard", effects.YouOwn(), effects.HasSubtype("Elf"))
	effects.Register(effects.Spec{
		OracleID:        "5b29c26f-c0f0-4703-83db-3a249ebc5792",
		Name:            "Mirkwood Elk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
