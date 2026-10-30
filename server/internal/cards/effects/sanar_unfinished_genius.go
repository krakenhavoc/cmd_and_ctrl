package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanar, Unfinished Genius // Wild Idea — Legendary Creature — Goblin
// Sorcerer {U}{R}, 0/4 // Sorcery {3}{U}{R} (Secrets of Strixhaven, #4456;
// preparation card, CR 722):
//
//	"Sanar enters prepared. (While it's prepared, you may cast a copy of
//	 its spell. Doing so unprepares it.)
//	 {T}: Create a Treasure token. Activate only if you've cast an
//	 instant or sorcery spell this turn."
//
//	Wild Idea — "Search your library for an instant or sorcery card,
//	 reveal it, put it into your hand, then shuffle."
//
// "Enters prepared" is SelfEntersPrepared. The activation condition reads
// the per-turn cast tally's instant-and-sorcery count, taken as each spell
// became cast (CR 601.2i): a copy of a prepare spell cast out of exile is
// cast, so casting Wild Idea's copy turns the Treasure on, and it stays
// counted after the copy has resolved and ceased to exist. A copy that was
// not cast (a storm or Twincast copy) does not count (CR 707.10). The {T}
// is the symbol (CR 302.6), so Sanar can't use it the turn it arrives.
// Wild Idea is Solve the Equation word for word, so it shares that
// card's body.
//
// No simplification.
const sanarUnfinishedGeniusOracleID = "a8c44bd9-6470-47d7-934f-77c76a44a984"

func init() {
	Register(Spec{
		OracleID:     sanarUnfinishedGeniusOracleID,
		Name:         "Sanar, Unfinished Genius",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Activated: []ActivatedAbility{{
			Label:     "{T}: Create a Treasure token. Activate only if you've cast an instant or sorcery spell this turn.",
			Cost:      TapCost(),
			Condition: sanarCastAnInstantOrSorceryThisTurn,
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:     sanarUnfinishedGeniusOracleID + "#1",
		Name:         "Wild Idea",
		Completeness: CompletenessFull,
		OnResolve:    b06TutorToHand("Wild Idea — an instant or sorcery card, revealed, to your hand", b08IsInstantOrSorceryCard),
	})
}

// sanarCastAnInstantOrSorceryThisTurn is "activate only if you've cast an
// instant or sorcery spell this turn".
func sanarCastAnInstantOrSorceryThisTurn(g *game.Game, controller, _ uuid.UUID) bool {
	return g.CastTallyFor(controller).InstantSorcery > 0
}
