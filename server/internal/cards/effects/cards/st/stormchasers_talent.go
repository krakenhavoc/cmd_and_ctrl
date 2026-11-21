package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormchaser's Talent — Enchantment — Class {U} (Bloomburrow, #4467):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create a 1/1 blue and red Otter creature
//	 token with prowess.
//	 {3}{U}: Level 2
//	 When this Class becomes level 2, return target instant or sorcery
//	 card from your graveyard to your hand.
//	 {5}{U}: Level 3
//	 Whenever you cast an instant or sorcery spell, create a 1/1 blue and
//	 red Otter creature token with prowess."
//
// A Class (CR 716): LevelUp climbs one level at a time at sorcery speed,
// BecomesLevel is Caretaker's Talent's level-2 trigger with Evolution
// Witness's return from your graveyard, and AtLevel gates the level-3
// cast trigger, which therefore resolves before the spell that caused
// it. The Otter is the token table's "1/1 blue and red Otter with
// prowess" row.
//
// No simplification.
func init() {
	const otter = "1/1 blue and red Otter with prowess"
	effects.Register(effects.Spec{
		OracleID:     "00c1788d-c52f-4839-9cd8-11fb86c418b6",
		Name:         "Stormchaser's Talent",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{3}{U}")),
			effects.LevelUp(3, effects.ManaCost("{5}{U}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stormchaser's Talent — create a 1/1 blue and red Otter creature token with prowess",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(otter), N: 1})),
			effects.Targeting(effects.BecomesLevel(2, "Stormchaser's Talent — return target instant or sorcery card from your graveyard to your hand",
				effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn())),
			effects.AtLevel(3, effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Stormchaser's Talent — create a 1/1 blue and red Otter creature token with prowess (instant or sorcery cast)",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(otter), N: 1}))),
		},
	})
}
