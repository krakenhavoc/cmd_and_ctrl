package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devourer of Destiny — Creature — Eldrazi {5}{C}{C}, 6/6 (Modern Horizons
// 3, S82 GAP2 batch 13, #5948, #6101):
//
//	"You may reveal this card from your opening hand. If you do, at the
//	 beginning of your first upkeep, look at the top four cards of your
//	 library. You may put one of those cards back on top of your library.
//	 Exile the rest.
//	 When you cast this spell, exile target permanent that's one or more
//	 colors."
//
// The reveal is the opening-hand reveal slot (ADR 0133 amendment
// 2026-11-24): the card stays in hand, revealed, and a delayed trigger
// controlled by its owner fires at the beginning of their first upkeep,
// the starting player's included. As it resolves its controller looks at
// the top four cards (only they see them), may keep one of them, which is
// then the top card once the rest are gone, and exiles the rest face up
// together.
//
// The cast trigger is Ugin, Eye of the Storms' (WhenYouCastThisSpell): it
// fires from the stack, is controlled by the caster and resolves even if
// the spell is countered. "One or more colors" is any permanent that is
// not colorless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89215e6d-ebe7-4eb0-9999-11ee2e6eb1d2",
		Name:         "Devourer of Destiny",
		Completeness: effects.CompletenessFull,
		OpeningHand: effects.RevealFromOpeningHandForYourFirstUpkeep(
			"Devourer of Destiny — look at the top four cards, keep one on top, exile the rest",
			devourerOfDestinyFirstUpkeep),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenYouCastThisSpell("Devourer of Destiny — exile target permanent that's one or more colors",
					effects.Do(effects.ExileChosenTarget{})),
				effects.TargetPermanent("target permanent that's one or more colors", effects.Not(effects.Colorless()))),
		},
	})
}

// devourerOfDestinyFirstUpkeep is the revealed card's upkeep trigger: look at
// the top four cards and ask which one, if any, goes back on top. An on-disk
// key: never renamed, never reused.
var devourerOfDestinyFirstUpkeep = game.SimpleDelayedBody("devourer-of-destiny/first-upkeep-look", func(g *game.Game, item *game.StackItem) error {
	you := item.Controller
	looked := g.LookAtTopOfLibraryForEffect(you, 4)
	if len(looked) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   item.SourceCardID,
		Question: "Devourer of Destiny — you may put one of those cards back on top of your library; the rest are exiled",
		Cards:    looked,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  devourerOfDestinyKept,
	})
	return nil
})

// devourerOfDestinyKept exiles every looked-at card but the one kept, which
// stays where it is and so is the top card once the rest have gone. An
// on-disk key: never renamed, never reused.
var devourerOfDestinyKept = game.RegisterChooseCardsThen("choose-cards/devourer-of-destiny-keep-one", func(g *game.Game, r game.CardsChosen) error {
	if !r.Answered {
		return nil
	}
	kept := make(map[uuid.UUID]bool, len(r.Picked))
	for _, id := range r.Picked {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, id := range r.Candidates {
		if kept[id] {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneLibrary {
			rest = append(rest, id)
		}
	}
	g.ExileCardsForEffect(rest)
	return nil
})
