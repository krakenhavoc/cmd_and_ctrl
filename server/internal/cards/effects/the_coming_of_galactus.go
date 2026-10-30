package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Coming of Galactus — Enchantment — Saga {2}{B}{B}{G} (Marvel Super
// Heroes, #4455):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Destroy up to one target nonland permanent.
//	 II, III — Each opponent loses 2 life.
//	 IV — Create Galactus, a legendary 16/16 black Elder Alien creature
//	 token with flying, trample, and "Whenever Galactus attacks, destroy
//	 target land.""
//
// Chapter I is a targeted chapter with "up to one" (ChapterTriggerTargeting,
// the clause allowing zero), so an empty battlefield drops it with no prompt.
// II and III are one effect on two chapters. Galactus is a catalog token
// (galactusToken): the attack trigger lives on the token, so it survives
// the Saga's sacrifice after IV and restores with the table. His trigger
// targets a land when he attacks (CR 603.3d), any player's.
//
// No simplification.
func init() {
	const name = "The Coming of Galactus"
	chapterOne := ChapterTriggerTargeting(1,
		SagaChapterLabel(name, 1, "destroy up to one target nonland permanent"),
		TargetPermanent("up to one target nonland permanent", Nonland()).WithCount(0, 1),
		destroyChosenPermanent)
	Register(Spec{
		OracleID:     "e05f23ba-f2b2-4c40-83fa-8defce3cb978",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapterOne,
			ChapterTrigger(2, SagaChapterLabel(name, 2, "each opponent loses 2 life"), theComingOfGalactusDrain),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "each opponent loses 2 life"), theComingOfGalactusDrain),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "create Galactus, a 16/16 legendary Elder Alien"),
				Do(CreateToken{Template: galactusToken(), N: 1})),
		},
	})
}

func theComingOfGalactusDrain(g *game.Game, item *game.StackItem) error {
	return eachOpponentLosesLife(g, item, 2)
}

// galactusToken is the legendary 16/16 black Elder Alien creature token with
// flying, trample and "Whenever Galactus attacks, destroy target land."
func galactusToken() game.Card { return tokenFromCatalog(printedGalactusToken) }

// printedGalactusToken is Galactus as PRINTED.
func printedGalactusToken() tokenTemplate {
	attack := WheneverThisAttacks("Galactus — destroy target land", destroyChosenPermanent)
	attack.Targets = TargetPermanent("target land", Land())
	return tokenTemplate{
		Slug: "galactus",
		Card: game.Card{
			Name:      "Galactus",
			TypeLine:  "Token Legendary Creature — Elder Alien",
			Power:     16,
			Toughness: 16,
			Colors:    []string{"B"},
			Keywords:  []string{"flying", "trample"},
		},
		Triggered: []game.TriggeredAbility{attack},
		Text:      "Flying, trample\nWhenever Galactus attacks, destroy target land.",
	}
}
