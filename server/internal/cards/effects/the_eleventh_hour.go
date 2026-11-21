package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Eleventh Hour — Enchantment — Saga {3}{U} (Doctor Who):
//
//	"I — Search your library for a Doctor card, reveal it, put it into your
//	 hand, then shuffle.
//	 II — Create a Food token and a 1/1 white Human creature token with
//	 'Doctor spells you cast cost {1} less to cast.'
//	 III — Create a token that's a copy of target creature, except it's a
//	 legendary Alien named Prisoner Zero."
//
// I is a library search for a card with the Doctor subtype. II makes the
// table's Food and the Human token whose printed ability is a cost modifier
// (gapDoctorHumanToken). III is a token copy of the targeted creature with
// the copy exception (CR 707.9b): its name is Prisoner Zero, it is
// Legendary, and its creature types are replaced by Alien (the exception
// does not say "in addition").
//
// No simplification.
func init() {
	const name = "The Eleventh Hour"
	copyIt := ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "create a token that's a copy of target creature, except it's a legendary Alien named Prisoner Zero"),
		TargetCreature("target creature"), theEleventhHourCopy)
	Register(Spec{
		OracleID:     "6b3e1aaf-6909-400a-88f2-9e9fde9a0cf7",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Tokens: 2, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "search your library for a Doctor card"),
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: gapIsDoctorCard,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    name + " — a Doctor card",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "create a Food token and a 1/1 white Human creature token"),
				Do(CreateToken{Template: FoodToken(), N: 1}, CreateToken{Template: gapDoctorHumanToken(), N: 1})),
			copyIt,
		},
	})
}

func theEleventhHourCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except: func(tok *game.Card) {
				tok.Name = "Prisoner Zero"
				tok.TypeLine = legendaryTypeLine(s82MidB11TypeLineWithCreatureType(tok.TypeLine, "Alien"))
			},
		}.Apply(ctx)
	}
	return nil
}
