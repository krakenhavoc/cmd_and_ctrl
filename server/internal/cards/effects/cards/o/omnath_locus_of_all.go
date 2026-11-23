package cardso

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Omnath, Locus of All — Legendary Creature — Phyrexian Elemental
// {W}{U}{B/P}{R}{G}, 4/4:
//
//	"If you would lose unspent mana, that mana becomes black instead.
//	 At the beginning of your first main phase, look at the top card of your
//	 library. You may reveal that card if it has three or more colored mana
//	 symbols in its mana cost. If you do, add three mana in any combination
//	 of its colors and put it into your hand. If you don't reveal it, put it
//	 into your hand."
//
// The first sentence is a mana-pool static (game.ManaPoolBecomesBlack, the
// Kruphix / Ozai family): at every step and phase boundary the unspent mana
// of its controller becomes black and stays, keeping its restrictions and
// riders. Black mana already in the pool stays black.
//
// The trigger looks at the top card (private to the looker). A card with
// three or more colored mana symbols (a hybrid or Phyrexian symbol is one
// colored symbol, CR 107.4e, 107.4f; a generic or colorless symbol is none)
// is offered as a reveal; revealing adds three mana, each independently one
// of the card's colors (an "any combination" slot per mana, as Selvala's), and
// the card goes into the hand either way: revealed after the mana, or
// unrevealed because the choice was declined or it was never eligible. The
// card moves from the library to the hand without being drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26445dc7-8363-4205-aee8-1cafeb4ba4c0",
		Name:         "Omnath, Locus of All",
		Completeness: effects.CompletenessFull,
		ManaPool:     []game.ManaPoolStatic{{Kind: game.ManaPoolBecomesBlack}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Omnath, Locus of All — look at the top card of your library; you may reveal it if it has three or more colored mana symbols and add three mana in any combination of its colors; put it into your hand",
				omnathLook),
		},
	})
}

// omnathColoredSymbols counts the colored mana symbols of a printed cost.
func omnathColoredSymbols(cost string) int {
	return strings.Count(effects.ColoredSymbolsAsProduced(cost), "{")
}

func omnathLook(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	looked := g.LookAtTopOfLibraryForEffect(player, 1)
	if len(looked) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(looked[0])
	if !ok {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if omnathColoredSymbols(card.ManaCost) < 3 {
		return effects.BounceToHand{Target: card.InstanceID}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Omnath, Locus of All — reveal it to add three mana in any combination of its colors? (It goes into your hand either way.)",
		Cards:    looked,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  omnathRevealed,
		Carry:    looked,
	})
	return nil
}

var omnathRevealed = game.RegisterChooseCardsThen("choose-cards/s82-mom-b21-omnath-reveal", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	ctx := effects.KeyedContext(g, r.Chooser, r.Source)
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if len(r.Picked) > 0 {
		if err := (effects.RevealCards{Player: r.Chooser, Cards: []uuid.UUID{id},
			Reason: "Omnath, Locus of All — reveal the top card of your library"}).Apply(ctx); err != nil {
			return err
		}
		if colors := card.EffectiveColors(); len(colors) > 0 {
			slot := "{" + strings.Join(colors, "|") + "}"
			if err := (effects.AddMana{Player: r.Chooser, Produced: strings.Repeat(slot, 3)}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	return effects.BounceToHand{Target: id}.Apply(ctx)
})
