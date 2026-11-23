package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain America, Team Leader — {R}{W}{U} Legendary Creature — Human
// Soldier Hero 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever another Hero you control enters, it gains vigilance and
//	 haste until end of turn. Put a +1/+1 counter on that Hero and a
//	 +1/+1 counter on Captain America."
//
// Dragon Tempest's enters trigger (enteredUnderYourControl, another)
// narrowed to the Hero subtype. The entering creature takes the grant
// and the first counter, and the source the second; a Hero that left
// before the trigger resolves gets nothing (CR 608.2b) while Captain
// America still grows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "916cf34a-fd2c-47fe-bebb-f7b52b3dabed",
		Name:         "Captain America, Team Leader",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && effects.CardHasSubtype(c, source.TextSubtype("Hero"))
			}, "Captain America — the Hero gains vigilance and haste; +1/+1 counters on it and on Captain America",
				captainAmericaTeamLeaderBoost),
		},
	})
}

func captainAmericaTeamLeaderBoost(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	hero := item.Trigger.Event.CardID
	if err := (effects.GrantKeywordUntilEOT{
		Target:   hero,
		Keywords: []string{"vigilance", "haste"},
		Label:    "Captain America, Team Leader",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.AddCounter{Target: hero, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.PlusOneCountersOnThis(1)(g, item)
}
