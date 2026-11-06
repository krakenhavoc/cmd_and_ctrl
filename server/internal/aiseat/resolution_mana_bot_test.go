package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4980, whole-engine: the real Soulbright Seeker with three Mountains
// and a vanilla {3}{R} 4/4 in hand, in the bot's first main phase,
// through the real enumerator, view and policy. Three activations cost
// {R}{R}{R} and the third resolution adds {R}{R}{R}{R}, the four-drop's
// mana. With PriceResolutionMana on (the default) the bot runs the
// Seeker three times and casts the 4/4; the view's resolved_this_turn,
// the engine's own count, is 3 afterwards. With it off the row is a
// trample grant that gets past no blocker, below a pass, and the bot
// does neither.
func TestHeuristicRunsSoulbrightSeekerForTheCastItPaysFor(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		t.Run(map[bool]string{true: "default", false: "no resolution mana"}[enabled], func(t *testing.T) {
			g := newSettledTable(t, 4980)
			advanceToStep(t, g, game.StepPrecombatMain)
			me := g.Seats[g.Turn.ActiveSeat]
			everyone := seatIDs(g)
			seeker := uuid.New()
			ogre := game.Card{InstanceID: uuid.New(), Name: "Test Ogre", TypeLine: "Creature — Ogre",
				ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
				KnownBy: map[uuid.UUID]bool{me.ID: true}}
			g.WithWriteLock(func() {
				for _, s := range g.Seats {
					s.Hand.Cards = nil
					s.Command.Cards = nil
				}
				me.Hand.PushTop(ogre)
				c := game.Card{InstanceID: seeker, Name: "Soulbright Seeker", OracleID: "762679dd-4345-43c3-929b-15959443aeb0",
					TypeLine: "Creature — Elemental Sorcerer", ManaCost: "{R}", Power: 2, Toughness: 1,
					Owner: me.ID, Controller: me.ID}
				c.AddKnowersAll(everyone)
				g.Battlefield.PushTop(c)
				for i := 0; i < 3; i++ {
					m := game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
						Owner: me.ID, Controller: me.ID}
					m.AddKnowersAll(everyone)
					g.Battlefield.PushTop(m)
				}
			})
			g.BumpLayerVersionForTest()
			cfg := heuristic.DefaultConfig()
			cfg.PriceResolutionMana = enabled
			made := playHeliodMainPhase(t, g, me, heuristic.NewWithConfig(cfg))
			activated, cast := 0, false
			for _, m := range made {
				if m.move.Kind == legal.KindActivate && m.move.Source == seeker {
					activated++
				}
			}
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == ogre.InstanceID {
					cast = true
				}
			}
			resolved := 0
			v := protocol.ViewOfGameFor(g, me.ID.String())
			for _, c := range v.Battlefield.Cards {
				if c.InstanceID == seeker.String() && len(c.ActivatedAbilities) > 0 {
					resolved = c.ActivatedAbilities[0].ResolvedThisTurn
				}
			}
			want, wantResolved := 3, 3
			if !enabled {
				want, wantResolved = 0, 0
			}
			if activated != want || cast != enabled || resolved != wantResolved {
				t.Errorf("the Seeker was activated %d times (want %d), the 4/4 cast %v (want %v), resolved_this_turn %d (want %d); the bot made %v",
					activated, want, cast, enabled, resolved, wantResolved, labelsOf(made))
			}
		})
	}
}
