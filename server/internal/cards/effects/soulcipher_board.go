package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soulcipher Board // Cipherbound Spirit — {1}{U} Artifact // Creature —
// Spirit 3/2 (Innistrad: Crimson Vow, #5495):
//
//	Front: "This artifact enters with three omen counters on it.
//	        {1}{U}, {T}: Look at the top two cards of your library. Put one
//	        of them into your graveyard.
//	        Whenever a creature card is put into your graveyard from
//	        anywhere, remove an omen counter from this artifact. Then if it
//	        has no omen counters on it, transform it."
//	Back:  "Flying
//	        This creature can block only creatures with flying.
//	        {3}{U}: Draw two cards, then discard a card."
//
// The trigger reads every route into the controller's graveyard (a mill, a
// discard or any other move, a death), one trigger per creature card
// (Voracious Brood's reader, not batched). The look is a hidden choice:
// the other card stays on top. A token is not a card and does not count.
//
// No simplification.
func init() {
	const oracle = "6469c7fd-6c0f-4233-8c38-4742ddcf3c3f"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Soulcipher Board",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("omen", 3, "Soulcipher Board: enters with three omen counters"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, {T}: Look at the top two cards of your library. Put one of them into your graveyard.",
			Purpose: game.Purpose{SelfMillTutor: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				player := item.Controller
				look := g.LookAtTopOfLibraryForEffect(player, 2)
				switch len(look) {
				case 0:
					return nil
				case 1:
					return restIntoGraveyard(g, look)
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  player,
					Source:   item.SourceCardID,
					Question: "Soulcipher Board — put one of them into your graveyard",
					Cards:    look,
					Min:      1,
					Max:      1,
					ThenKey:  soulcipherBoardPicked,
					LiveZone: game.ZoneLibrary,
				})
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			OnAny(graveyardArrivalKinds, creatureCardsToYourGraveyard.When(),
				"Soulcipher Board — remove an omen counter; if it has none, transform it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (AddCounter{Target: ctx.Source(), Kind: "omen", N: -1}).Apply(ctx); err != nil {
						return err
					}
					if c, ok := g.LookupCardForEffect(ctx.Source()); ok && c.Counters["omen"] == 0 {
						return TransformThis{}.Apply(ctx)
					}
					return nil
				}),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Cipherbound Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules:      nemBlocksOnlyFlyers(OnSelf()),
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Draw two cards, then discard a card.",
			Purpose: game.Purpose{Draws: 2, Discards: 1, Answers: game.AnswerValue},
			Cost:    ManaCost("{3}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, "")
			},
		}},
	})
}

// soulcipherBoardPicked puts the chosen card of the look into the graveyard;
// the other stays on top. An on-disk key: never renamed, never reused.
var soulcipherBoardPicked = game.RegisterChooseCardsThen("choose-cards/soulcipher-board-put-one-into-graveyard", func(g *game.Game, r game.CardsChosen) error {
	return restIntoGraveyard(g, r.Picked)
})
