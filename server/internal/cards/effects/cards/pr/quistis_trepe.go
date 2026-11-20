package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quistis Trepe — Legendary Creature — Human Wizard {2}{U}, 2/2 (Final
// Fantasy, #4462):
//
//	"Blue Magic — When Quistis Trepe enters, you may cast target instant or
//	 sorcery card from a graveyard, and mana of any type can be spent to cast
//	 that spell. If that spell would be put into a graveyard, exile it
//	 instead."
//
// Emet-Selch of the Third Seat's graveyard cast over any player's graveyard
// (offerAnyGraveyardCastWithAnyMana): the target is chosen as the trigger
// goes on the stack and re-checked as it resolves (CR 608.2b); the cast is
// made during the resolution (CR 608.2g), at the card's mana cost, with
// timing restrictions ignored, and any mana, colourless or of any colour,
// may pay for it (CR 609.4b). The spell is exiled instead of going to a
// graveyard, whether it resolves, is countered or fizzles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "821e8648-222c-4b33-a8bd-e8bfff7dcd9e",
		Name:         "Quistis Trepe",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Quistis Trepe — Blue Magic: you may cast target instant or sorcery card from a graveyard, spending mana as though it were mana of any type; exile it instead of a graveyard",
					quistisBlueMagic),
				effects.TargetCardInGraveyard("target instant or sorcery card from a graveyard", effects.Or(effects.Instant(), effects.Sorcery()))),
		},
	})
}

func quistisBlueMagic(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.OfferAnyGraveyardCastWithAnyMana(ctx, t.ID, "Quistis Trepe")
	}
	return nil
}
