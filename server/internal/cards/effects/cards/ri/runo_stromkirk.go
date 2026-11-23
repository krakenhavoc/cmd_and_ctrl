package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Runo Stromkirk // Krothuss, Lord of the Deep — {1}{U}{B} Legendary
// Creature — Vampire Cleric 1/4 // Legendary Creature — Kraken Horror 3/5
// (Innistrad: Crimson Vow, #5495):
//
//	Front: "Flying
//	        When Runo enters, put up to one target creature card from your
//	        graveyard on top of your library.
//	        At the beginning of your upkeep, look at the top card of your
//	        library. You may reveal that card. If a creature card with mana
//	        value 6 or greater is revealed this way, transform Runo."
//	Back:  "Flying
//	        Whenever Krothuss attacks, create a tapped and attacking token
//	        that's a copy of another target attacking creature. If that
//	        creature is a Kraken, Leviathan, Octopus, or Serpent, create two
//	        of those tokens instead."
//
// The upkeep look is private to the looker. Revealing is only offered for
// a card that could transform Runo (a creature card with mana value 6 or
// greater), as Omnath, Locus of All offers its reveal only for a card it
// can use: revealing any other card changes nothing in the game. The
// creature types are words of the target as it is (CR 612.2, a changeling
// is all of them). The tokens are copies made tapped and attacking, never
// declared as attackers (CR 508.4): their controller chooses what they
// attack, offered what the original attacks first.
//
// No simplification.
func init() {
	const oracle = "70d7d1a7-9d39-44ec-a4ea-a70ffa453d23"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Runo Stromkirk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Runo Stromkirk — put up to one target creature card from your graveyard on top of your library",
					effects.PutChosenTargetOnTopOfLibrary),
				effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 1)),
			effects.AtYourUpkeep("Runo Stromkirk — look at the top card of your library; you may reveal it; if it is a creature card with mana value 6 or greater, transform Runo",
				runoUpkeep),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Krothuss, Lord of the Deep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Krothuss, Lord of the Deep — create a tapped and attacking token that's a copy of another target attacking creature (two if it is a Kraken, Leviathan, Octopus, or Serpent)",
					krothussAttack),
				effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))),
		},
	})
}

func runoUpkeep(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	looked := g.LookAtTopOfLibraryForEffect(player, 1)
	if len(looked) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(looked[0])
	if !ok || !card.IsCreature() || card.ManaValue() < 6 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Runo Stromkirk — reveal the top card of your library to transform Runo?",
		Cards:    looked,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  runoRevealed,
		Carry:    looked,
	})
	return nil
}

// runoRevealed reveals the looked-at card and transforms Runo. An on-disk
// key: never renamed, never reused.
var runoRevealed = game.RegisterChooseCardsThen("choose-cards/s82-vow-b15-runo-reveal", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	ctx := effects.KeyedContext(g, r.Chooser, r.Source)
	if err := (effects.RevealCards{Player: r.Chooser, Cards: []uuid.UUID{r.Carry[0]},
		Reason: "Runo Stromkirk — reveal the top card of your library"}).Apply(ctx); err != nil {
		return err
	}
	return effects.Transform{Target: r.Source}.Apply(ctx)
})

func krothussAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		orig, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		tmpl, ok := effects.TokenCopyTemplate(g, t.ID)
		if !ok {
			return nil
		}
		n := 1
		for _, kind := range []string{"Kraken", "Leviathan", "Octopus", "Serpent"} {
			if orig.HasSubtypeWord(kind) {
				n = 2
				break
			}
		}
		return effects.CreateTokensAttackingYourChoice{
			Controller: item.Controller,
			Template:   tmpl,
			Copy:       true,
			N:          n,
			Tapped:     true,
			Prefer:     orig.AttackingTarget,
			Label:      "Krothuss, Lord of the Deep",
		}.Apply(ctx)
	}
	return nil
}
