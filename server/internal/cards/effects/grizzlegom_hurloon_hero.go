package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grizzlegom, Hurloon Hero — {5}{R}{G} Legendary Creature — Minotaur
// Warrior 4/4 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Rulebreaker — A deck with this commander can have any land cards.
//	 Whenever Grizzlegom attacks, create a 1/1 white Soldier creature
//	 token for each Plains you control. Draw a card for each Island you
//	 control. Each opponent loses 1 life for each Swamp you control. Put
//	 a +1/+1 counter on Grizzlegom for each Mountain you control. You
//	 gain 1 life for each Forest you control."
//
// One trigger, five instructions in the printed order. Each count is the
// number of lands with that land type its controller controls as the
// instruction is carried out, read as a word of the source's text (CR
// 612.2); a count of zero does nothing. The counters go on only if
// Grizzlegom is still the same permanent (CR 400.7). The Rulebreaker
// paragraph is a deckbuilding exception, read by deck validation off the
// oracle text (internal/deck, rulebreaker.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f5db55a-8bae-4e97-86bd-f2a2dacf07a4",
		Name:         "Grizzlegom, Hurloon Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Grizzlegom, Hurloon Hero — Soldiers for each Plains, cards for each Island, drain for each Swamp, counters for each Mountain, life for each Forest",
				grizzlegomAttack),
		},
	})
}

// landsOfTypeYouControl counts the lands with the land type `word` that
// `player` controls, reading the word through the reader's text.
func landsOfTypeYouControl(g *game.Game, reader *game.Card, item *game.StackItem, word string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsLand() && game.ThroughTextOf(reader, c).HasSubtypeWord(word) {
			n++
		}
	}
	return n
}

func grizzlegomAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	card, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	reader := &card
	if n := landsOfTypeYouControl(g, reader, item, "Plains"); n > 0 {
		if err := (CreateToken{Template: TokenCard("1/1 white Soldier"), N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	if n := landsOfTypeYouControl(g, reader, item, "Island"); n > 0 {
		if err := (DrawCards{N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	if n := landsOfTypeYouControl(g, reader, item, "Swamp"); n > 0 {
		if err := (EachOpponentLosesLife{N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	if n := landsOfTypeYouControl(g, reader, item, "Mountain"); n > 0 && onBattlefield(g, item.SourceCardID) && !sourceIsNewObject(g, item) {
		if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	if n := landsOfTypeYouControl(g, reader, item, "Forest"); n > 0 {
		return GainLife{Amount: n}.Apply(ctx)
	}
	return nil
}
