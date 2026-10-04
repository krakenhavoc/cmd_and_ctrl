package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boromir, Warden of the Tower — Legendary Creature — Human Soldier
// {2}{W}, 3/3:
//
//	"Vigilance
//	 Whenever an opponent casts a spell, if no mana was spent to cast
//	 it, counter that spell.
//	 Sacrifice Boromir: Creatures you control gain indestructible until
//	 end of turn. The Ring tempts you."
//
// The second ability is Vexing Bauble's, narrowed to opponents: an
// intervening if (CR 603.4) read from what was actually paid, checked
// as the spell is cast and again as the ability resolves. A spell cast
// with an alternative cost that includes mana, or for free but with
// mana paid for an additional cost, had mana spent on it and is not
// countered (2023-06-16 rulings). The creatures that gain
// indestructible are fixed as the ability resolves (CR 611.2c).
//
// Caveat: the engine knows what was spent only with strict mana on.
// With it off, NoManaWasSpentToCast answers "unknown", and Boromir
// never counters anything — the weaker direction, as Vexing Bauble
// declares.
func init() {
	Register(Spec{
		OracleID:     "d72b57e5-1c6e-4f02-968f-7e2dd24f0d3c",
		Name:         "Boromir, Warden of the Tower",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"With strict mana off, the game doesn't track what was spent, so Boromir never counters a spell — turn strict mana on for it to work.",
		},
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AllOf(AnOpponentCast(nil), func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return NoManaWasSpentToCast(g, ev.CardID)
			}), "Boromir, Warden of the Tower — counter that spell", counterTheSpellIfNoManaWasSpent),
		},
		Activated: []ActivatedAbility{{
			Label:  "Sacrifice Boromir: Creatures you control gain indestructible until end of turn. The Ring tempts you",
			Cost:   SacrificeThis(),
			Effect: boromirLastStand,
		}},
	})
}

// boromirLastStand is the sacrifice ability's effect.
func boromirLastStand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"indestructible"},
		Label:    "Boromir, Warden of the Tower — indestructible",
	}).Apply(ctx); err != nil {
		return err
	}
	return TheRingTemptsYou{}.Apply(ctx)
}
