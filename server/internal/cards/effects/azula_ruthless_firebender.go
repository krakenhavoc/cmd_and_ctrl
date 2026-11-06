package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azula, Ruthless Firebender — Legendary Creature — Human Noble {2}{B},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Firebending 1 (Whenever this creature attacks, add {R}. This mana
//	 lasts until end of combat.)
//	 Whenever Azula attacks, you may discard a card. Then you get an
//	 experience counter for each player who discarded a card this turn.
//	 {2}{B}: Until end of turn, Azula gets +1/+1 for each experience
//	 counter you have and gains menace."
//
// Firebending is Fire Sages' trigger (firebending(1)), a triggered
// ability that uses the stack, separate from the second one.
//
// The discard is optional ("you may", up to one card). Whatever the
// answer, the count is then taken from the per-turn tallies of every
// player still in the game, the discard just made included: each player
// with at least one discard this turn is one experience counter. The
// counters go on you, the player (CR 122.1), and stay after Azula leaves.
// All experience counters are the same counter, so Aang's and Zuko's
// count here too.
//
// The pump reads the experience counters once, as it resolves (CR
// 608.2h), so a counter gained later this turn adds nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7139ead5-86f0-46ae-8a6d-ad66d0408f73",
		Name:         "Azula, Ruthless Firebender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Azula, Ruthless Firebender — firebending 1", Do(firebending(1))),
			WheneverThisAttacks("Azula, Ruthless Firebender — you may discard a card, then get experience counters",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						UpTo:     true,
						Question: "Azula — you may discard a card",
					}, avB30dAzulaExperience, nil)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Until end of turn, Azula gets +1/+1 for each experience counter you have and gains menace.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				self := item.SourceCardID
				if !onBattlefield(g, self) {
					return nil
				}
				n := experienceCounters(g, item.Controller)
				if err := (BoostUntilEOT{Target: self, Power: n, Toughness: n, Label: "Azula — +1/+1 for each experience counter"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: self, Keywords: []string{"menace"}, Label: "Azula — menace"}.Apply(ctx)
			},
		}},
	})
}

// avB30dAzulaExperience is "then you get an experience counter for each
// player who discarded a card this turn", once the optional discard is
// answered. An on-disk key, never renamed.
var avB30dAzulaExperience = game.RegisterDiscardThen("avatar-b30/azula-experience-after-discard",
	func(g *game.Game, r game.DiscardsDone) error {
		n := 0
		for _, p := range tablePlayers(keyedContext(g, r.Player, r.Source)) {
			if g.TurnTallyFor(p).CardsDiscarded > 0 {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return g.AddPlayerCounterByForEffect(r.Player, r.Player, game.CounterExperience, n)
	})
