package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ardbert, Warrior of Darkness — Legendary Creature — Spirit Warrior
// {1}{W}{B}, 2/2:
//
//	"Whenever you cast a white spell, put a +1/+1 counter on each
//	 legendary creature you control. They gain vigilance until end of
//	 turn.
//	 Whenever you cast a black spell, put a +1/+1 counter on each
//	 legendary creature you control. They gain menace until end of
//	 turn."
//
// Two triggers, so a white-black spell triggers both. The set is the
// legendary creatures you control as each trigger resolves, read once
// (CR 611.2c): the counters go on them as one placement each, and the
// keyword goes to the same creatures. Ardbert is legendary himself, so
// he is among them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9243760-b9b1-4d43-8550-3528b5571dae",
		Name:         "Ardbert, Warrior of Darkness",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(OfColor("W"), "Ardbert, Warrior of Darkness — +1/+1 counter on each legendary creature you control; they gain vigilance",
				ardbertCountersAndKeyword("vigilance")),
			WheneverYouCast(OfColor("B"), "Ardbert, Warrior of Darkness — +1/+1 counter on each legendary creature you control; they gain menace",
				ardbertCountersAndKeyword("menace")),
		},
	})
}

func ardbertCountersAndKeyword(keyword string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == item.Controller && c.IsCreature() && c.IsLegendary() {
				ids = append(ids, c.InstanceID)
			}
		}
		for _, id := range ids {
			if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{keyword},
				Label:    "Ardbert, Warrior of Darkness — " + keyword,
			}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
