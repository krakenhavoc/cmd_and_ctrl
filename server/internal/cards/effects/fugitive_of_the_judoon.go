package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fugitive of the Judoon — Enchantment — Saga {4}{G} (Doctor Who, #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a 1/1 white Human creature token with ward {2} and a 4/4 white
//	 Alien Rhino creature token.
//	 II — Investigate.
//	 III — You may exile a Human you control and an artifact you control. If you
//	 do, search your library for a Doctor card, put it onto the battlefield,
//	 then shuffle."
//
// Chapter I makes both tokens. Chapter III asks for one set of two different
// permanents the controller controls: a Human and an artifact (a single Human
// artifact cannot be both), or nothing. If they choose two they are exiled
// together and the controller searches their library for a Doctor card (a
// card with the Doctor creature type), puts it onto the battlefield and
// shuffles; declining or choosing a set that is not a Human and an artifact
// does nothing.
//
// No simplification.
func init() {
	const name = "Fugitive of the Judoon"
	Register(Spec{
		OracleID:     "cacda1c2-2fe8-40a3-89f3-38a835f10af3",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, gapB92ChapterLabel(name, 1, "create a 1/1 white Human creature token with ward {2} and a 4/4 white Alien Rhino creature token"),
				Do(CreateToken{Template: gapB92WardedHumanToken(), N: 1}, CreateToken{Template: gapB92AlienRhinoToken(), N: 1})),
			ChapterTrigger(2, gapB92ChapterLabel(name, 2, "investigate"), Do(Investigate{N: 1})),
			ChapterTrigger(3, gapB92ChapterLabel(name, 3, "you may exile a Human and an artifact you control; if you do, search your library for a Doctor card, put it onto the battlefield, then shuffle"),
				fugitiveChapterIII),
		},
	})
}

func fugitiveChapterIII(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	human := ctx.TextSubtype("Human")
	return ChoosePermanents{
		Question: "Fugitive of the Judoon — you may exile a Human you control and an artifact you control: choose the Human (or none)",
		Keyed:    fugitiveHumanChosen,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.HasSubtype(human) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, min(1, len(out))
		},
	}.Apply(ctx)
}

// fugitiveHumanChosen is the first half of the "and": with a Human chosen,
// the controller is asked for a different artifact they control. An on-disk
// key: never renamed, never reused.
var fugitiveHumanChosen = RegisterPermanentsPick("prompt-run/fugitive-of-the-judoon-human", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		chosen := picked.Cards()
		if len(chosen) != 1 {
			return nil
		}
		human := chosen[0]
		return ChoosePermanents{
			Question: "Fugitive of the Judoon — choose an artifact you control to exile with the Human (or none)",
			Keyed:    fugitiveExiled,
			Carry:    []uuid.UUID{human},
			Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == of && c.IsArtifact() && c.InstanceID != human {
						out = append(out, c.InstanceID)
					}
				}
				return out, 0, min(1, len(out))
			},
		}.Apply(ctx)
	},
})

// fugitiveExiled exiles the pair and fetches the Doctor. An on-disk key:
// never renamed, never reused.
var fugitiveExiled = RegisterPermanentsPick("prompt-run/fugitive-of-the-judoon-exile", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		artifacts := picked.Cards()
		if len(artifacts) != 1 || len(r.Carry) != 1 {
			return nil
		}
		for _, id := range []uuid.UUID{r.Carry[0], artifacts[0]} {
			if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		doctor := ctx.TextSubtype("Doctor")
		return SearchLibrary{
			Player:    ctx.Controller(),
			Predicate: func(c game.Card) bool { return c.HasSubtype(doctor) },
			Dest:      game.ZoneBattlefield,
			Limit:     1,
			Shuffle:   true,
			Reason:    "Fugitive of the Judoon — a Doctor card",
		}.Apply(ctx)
	},
})
