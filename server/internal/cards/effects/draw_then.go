package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draw_then.go — "draw N cards, then X" (#2391, CR 608.2c).
//
// A clause printed after a draw must wait for the draw to resolve, and
// a draw can pause on a prompt (a dredge offer, a draw-instead pick, a
// CR 616 ordering prompt). Writing the clause as the statement after
// DrawCards runs it as soon as the call returns, with the draw still
// unanswered. Every "draw, then …" card therefore hands the rest of
// itself to Game.DrawNThenForEffect as a registered body plus scalars
// (game.DrawThen), which the engine runs once the last draw is done.
//
// The bodies here are stateless and capture only a label, so they sit
// in a registry keyed by what they capture, exactly as
// game.RegisterDrawInstead does.

// drawThenDiscardRef registers "then discard N" with a prompt
// question (empty for the default).
func drawThenDiscardRef(question string) game.DrawThenRef {
	return game.RegisterDrawThen("discard:"+question, func(g *game.Game, d game.DrawThen) error {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        d.N,
			Question: question,
		})
		return nil
	})
}

// drawThenDiscardQuestions is every question a caller hands
// drawThenDiscard. Each body is registered at init (#4253): a draw paused
// on a CR 616 prompt is a restore point, its frame names the body by
// key, and a freshly booted binary must already know that key when it
// restores the file. TestEveryDrawThenDiscardQuestionIsRegisteredAtInit
// reads the callers and fails on a question missing here.
var drawThenDiscardQuestions = []string{
	"",
	"Anvil of Bogardan — discard a card",
	"Arni, Humble Scribe — discard a card",
	"Depth Defiler — discard a card",
}

func init() {
	for _, q := range drawThenDiscardQuestions {
		drawThenDiscardRef(q)
	}
}

// drawThenDiscard is "<player> draws `draw` cards, then discards
// `discard` cards" — the discard prompt opens only after every draw has
// resolved, so the drawn cards are legal discards.
func drawThenDiscard(g *game.Game, player, source uuid.UUID, draw, discard int, question string) error {
	ref := drawThenDiscardRef(question)
	return g.DrawNThenForEffect(player, draw, game.DrawThen{Ref: ref, Player: player, Source: source, N: discard})
}

// discardTwoUnlessOne registers "then discard two cards unless you
// discard <a card matching qualifies>" (Thirst for Knowledge, Thirst for
// Discovery): a prompt whose floor is one and whose set rule accepts a
// single card only when it qualifies, or two cards of any kind (the
// Compulsive Research shape). A hand of one card is discarded whole, and
// an empty one asks nothing (CR 701.9a).
func discardTwoUnlessOne(key, question string, qualifies func(game.Card) bool) game.DrawThenRef {
	// The prompt is keyed (#4253): its set rule is registered with the
	// discard continuation, so a table waiting on the discard is a restore
	// point. The continuation itself does nothing: the discard is the clause.
	asked := discardTwoUnlessOneRule("choose-cards/"+key, qualifies)
	return game.RegisterDrawThen(key, func(g *game.Game, d game.DrawThen) error {
		p := g.PlayerByIDForEffect(d.Player)
		if p == nil {
			return nil
		}
		switch len(p.Hand.Cards) {
		case 0:
			return nil
		case 1:
			g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: 1, Question: question})
			return nil
		}
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        d.N,
			Min:      1,
			Question: question,
		}, asked, nil)
	})
}

// discardTwoUnlessOneRule registers the keyed discard behind "discard two
// cards unless you discard <a card that qualifies>": a single card is a
// legal answer only when it qualifies, two cards of any kind are (the
// Compulsive Research shape). The key is an on-disk identity ("choose
// <key>"): never renamed, never reused.
func discardTwoUnlessOneRule(key string, qualifies func(game.Card) bool) game.DiscardThen {
	return game.RegisterDiscardThenRule(key, func(picked []game.Card) bool {
		switch len(picked) {
		case 2:
			return true
		case 1:
			return qualifies(picked[0])
		default:
			return false
		}
	}, func(*game.Game, game.DiscardsDone) error { return nil })
}

var (
	// "then discard two cards unless you discard an artifact card" —
	// Thirst for Knowledge.
	drawThenDiscardTwoUnlessArtifact = discardTwoUnlessOne("discard-two-unless-artifact",
		"Thirst for Knowledge — discard two cards, or one artifact card",
		func(c game.Card) bool { return c.IsArtifact() })

	// "then discard two cards unless you discard a basic land card" —
	// Thirst for Discovery.
	drawThenDiscardTwoUnlessBasicLand = discardTwoUnlessOne("discard-two-unless-basic-land",
		"Thirst for Discovery — discard two cards, or one basic land card",
		IsBasicLand)

	// "then, unless you attacked this turn, discard N" — Chart a Course.
	drawThenDiscardUnlessAttacked = game.RegisterDrawThen("discard-unless-attacked", func(g *game.Game, d game.DrawThen) error {
		if b18AttackedThisTurn(g, d.Player) {
			return nil
		}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
		return nil
	})

	// "then put two cards from your hand on top of your library in any
	// order" — Brainstorm.
	drawThenBrainstormPutBack = game.RegisterDrawThen("brainstorm-put-back", func(g *game.Game, d game.DrawThen) error {
		ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
		return PutFromHandOnTopInAnyOrder{
			Player: d.Player,
			N:      d.N,
			Label:  "Brainstorm — put two cards from your hand on top of your library",
		}.Apply(ctx)
	})

	// "then put a card from your hand on top of your library" — Enter
	// the Infinite.
	drawThenEnterTheInfinitePutBack = game.RegisterDrawThen("enter-the-infinite-put-back", func(g *game.Game, d game.DrawThen) error {
		ctx := NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
		return PutFromHandOnTopInAnyOrder{
			Player: d.Player,
			N:      d.N,
			Label:  "Enter the Infinite — put a card from your hand on top of your library",
		}.Apply(ctx)
	})
)
