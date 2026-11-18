package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Restoration of Eiganjo // Architect of Restoration — an
// Enchantment — Saga {2}{W} (Kamigawa: Neon Dynasty, S82):
//
//	I   — Search your library for a basic Plains card, reveal it, put
//	      it into your hand, then shuffle.
//	II  — You may discard a card. When you do, return target permanent
//	      card with mana value 2 or less from your graveyard to the
//	      battlefield tapped.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// Chapter II is Awaken the Honored Dead's shape: a "you may discard"
// whose "when you do" is a reflexive trigger (CR 603.12) that picks its
// target as it goes on the stack, so the card is chosen after the discard
// and no discard means no trigger. Chapter III is ADR 0079's
// exile-and-return, as on Fable of the Mirror-Breaker.
//
// The back face is registered under "<oracle_id>#1" (Architect of
// Restoration, an Enchantment Creature — Fox Monk 3/4): vigilance, and
// "Whenever this creature attacks or blocks, create a 1/1 colorless Spirit
// creature token."
//
// No simplification.
func init() {
	const name = "The Restoration of Eiganjo"
	Register(Spec{
		OracleID:     theRestorationOfEiganjoOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "search your library for a basic Plains card, reveal it, put it into your hand, then shuffle"),
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLandWithSubtype("plains"),
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    name + " — a basic Plains card",
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "you may discard a card; when you do, return target permanent card with mana value 2 or less from your graveyard to the battlefield tapped"),
				theRestorationOfEiganjoDiscard),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile this Saga, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        theRestorationOfEiganjoOracleID + "#1",
		Name:            "Architect of Restoration",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Architect of Restoration — create a 1/1 colorless Spirit creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: 1})),
		},
	})
}

// theRestorationOfEiganjoOracleID is shared with the back face.
const theRestorationOfEiganjoOracleID = "9d9b52b2-2edc-4f7f-a8d9-e024b1398847"

var theRestorationOfEiganjoReturnBody = game.ReflexiveBody("the-restoration-of-eiganjo/return-tapped",
	simpleBody(returnFirstLegalGraveyardTargetToBattlefieldTapped),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target permanent card with mana value 2 or less from your graveyard",
			YouOwn(), Permanent(), ManaValueLE(2))
	}))

func theRestorationOfEiganjoDiscard(g *game.Game, item *game.StackItem) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        1,
		UpTo:     true,
		Question: "The Restoration of Eiganjo — you may discard a card to return a permanent card with mana value 2 or less from your graveyard to the battlefield tapped",
	}, theRestorationOfEiganjoDiscarded, nil)
}

var theRestorationOfEiganjoDiscarded = game.RegisterDiscardThen("choose-cards/the-restoration-of-eiganjo-return", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Discarded.Cards()) == 0 {
		return nil
	}
	return WhenYouDo("The Restoration of Eiganjo — return target permanent card with mana value 2 or less from your graveyard to the battlefield tapped",
		theRestorationOfEiganjoReturnBody).Apply(keyedContext(g, r.Player, r.Source))
})
