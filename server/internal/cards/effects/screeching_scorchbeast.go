package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Screeching Scorchbeast — Creature — Bat Mutant {4}{B}{B}, 5/5:
//
//	"Flying, menace
//	 Whenever this creature attacks, each player gets two rad counters.
//	 Whenever one or more nonland cards are milled, you may create that
//	 many 2/2 black Zombie Mutant creature tokens. Do this only once
//	 each turn."
//
// #2809. The mill trigger is WheneverOneOrMoreNonlandCardsAreMilled: one
// trigger per mill, whoever milled, and "that many" is the nonland cards
// that mill put into graveyards (a land milled beside them is not
// counted). "Do this only once each turn" limits the creating, not the
// triggering (CR 603.2h): the ability triggers only while no tokens have
// been made by it this turn, asks its "you may" as it resolves (CR
// 603.5), and only a "yes" uses up the turn, as Terrasymbiosis's does.
// The record is per object (b28DoneThisTurn).
//
// No simplification.
func init() {
	t := WheneverOneOrMoreNonlandCardsAreMilled(screechingScorchbeastLabel,
		func(g *game.Game, item *game.StackItem) error {
			if b28DoneThisTurn(g, item.SourceCardID, screechingScorchbeastDone) {
				return nil
			}
			b28NoteDoneThisTurn(g, item.SourceCardID, screechingScorchbeastDone)
			return CreateToken{Template: TokenCard("2/2 black Zombie Mutant"), N: nonlandCardsMilledThisWay(item)}.Apply(NewContext(g, item))
		})
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
		return when(ev, source, ch, g) && !b28DoneThisTurn(g, source.InstanceID, screechingScorchbeastDone)
	}
	Register(Spec{
		OracleID:        "ed5889c9-e524-40c1-9011-7559a992bb1b",
		Name:            "Screeching Scorchbeast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Screeching Scorchbeast — each player gets two rad counters",
				func(g *game.Game, item *game.StackItem) error {
					return eachPlayerGetsRadCounters(g, item.Controller, 2)
				}),
			Optional(t, "Screeching Scorchbeast — create that many 2/2 black Zombie Mutant tokens?"),
		},
	})
}

// screechingScorchbeastLabel is the mill trigger's stack label.
const screechingScorchbeastLabel = "Screeching Scorchbeast — create that many 2/2 black Zombie Mutant tokens"

// screechingScorchbeastDone is the once-each-turn record's label: the
// tokens were made this turn (b28DoneThisTurn).
const screechingScorchbeastDone = "Screeching Scorchbeast — tokens created this turn"
