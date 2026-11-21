package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delverOracle is shared by both faces: the front registers under it, the
// back under delverOracle + "#1".
const DelverOracle = "edd531b9-f615-4399-8c8c-1c5e18c4acbf"

// Delver of Secrets // Insectile Aberration — a transforming creature
// (Innistrad, reprinted in Midnight Hunt Commander, S82):
//
//	Delver of Secrets {U}, Creature — Human Wizard, 1/1
//	  "At the beginning of your upkeep, look at the top card of your
//	   library. You may reveal that card. If an instant or sorcery card is
//	   revealed this way, transform this creature."
//
//	Insectile Aberration, Creature — Human Insect, 3/2
//	  "Flying"
//
// Runo Stromkirk's upkeep look: private to the looker, with the reveal
// offered only for an instant or sorcery card, since revealing any other
// card changes nothing in the game. The turn over is in place only if the
// Delver is still the same face-up permanent (CR 701.27f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     DelverOracle,
		Name:         "Delver of Secrets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Delver of Secrets — look at the top card of your library; you may reveal it; if it is an instant or sorcery card, transform this creature",
				delverUpkeep),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        DelverOracle + "#1",
		Name:            "Insectile Aberration",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}

func delverUpkeep(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	looked := g.LookAtTopOfLibraryForEffect(player, 1)
	if len(looked) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(looked[0])
	if !ok || !(card.IsInstant() || card.IsSorcery()) {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Delver of Secrets — reveal the top card of your library to transform this creature?",
		Cards:    looked,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  delverRevealed,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// delverRevealed reveals the looked-at card and transforms the Delver. An
// on-disk key: never renamed, never reused.
var delverRevealed = effects.AskedPick("choose-cards/s82-rpr-b30-delver-reveal", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	if err := (effects.RevealCards{Player: r.Chooser, Cards: []uuid.UUID{r.Carry[0]},
		Reason: "Delver of Secrets — reveal the top card of your library"}).Apply(ctx); err != nil {
		return err
	}
	return effects.TransformThis{}.Apply(ctx)
})
