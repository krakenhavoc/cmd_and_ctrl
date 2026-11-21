package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const jonIrenicusGrant = "jon-irenicus-shattered-one/cant-be-sacrificed"

// Jon Irenicus, Shattered One — {2}{U}{B} Legendary Creature — Elf Wizard 3/3
//
//	"At the beginning of your end step, target opponent gains control of up to
//	 one target creature you control. Put two +1/+1 counters on it and tap it.
//	 It's goaded for the rest of the game and it gains "This creature can't be
//	 sacrificed."
//	 Whenever a creature you own but don't control attacks, you draw a card."
//
// The control change has no duration (CR 611.2a), so it lasts until the
// creature or the game is gone. Everything after the first sentence happens to
// the creature only if the opponent really took it (it is still on the
// battlefield under them): the counters, the tap, a goad by Jon's controller
// that lasts for the rest of the game (CR 701.15), and the layer 6 "can't be
// sacrificed" grant. The draw is one per creature that attacks (CR 508.1),
// Jon's controller owning it and someone else controlling it.
func init() {
	end := effects.Targeting(
		effects.AtYourEndStep("Jon Irenicus, Shattered One — target opponent gains control of up to one target creature you control; two +1/+1 counters, tap it, goaded for the rest of the game, can't be sacrificed",
			jonIrenicusGive),
		effects.Clauses(
			effects.TargetPlayer("target opponent", effects.Opponent()),
			effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1)))
	effects.Register(effects.Spec{
		OracleID:     "c5c19d6e-16d5-4d24-be2d-c63646bdf00c",
		Name:         "Jon Irenicus, Shattered One",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:    jonIrenicusGrant,
			Static: []game.StaticAbility{effects.CantBeSacrificed()},
			Text:   "This creature can't be sacrificed.",
		}},
		Triggered: []game.TriggeredAbility{
			end,
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Owner == source.Controller && c.Controller != source.Controller
			}, "Jon Irenicus, Shattered One — you draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

func jonIrenicusGive(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	opp, ok := ctx.ClauseTarget(0)
	if !ok || opp.Kind != game.TargetPlayer {
		return nil
	}
	perm, ok := ctx.ClauseTarget(1)
	if !ok || perm.Kind != game.TargetCard {
		return nil
	}
	if err := (effects.GainControl{Target: perm.ID, Controller: opp.ID, Duration: game.IndefiniteDuration(),
		Label: "Jon Irenicus — target opponent gains control"}).Apply(ctx); err != nil {
		return err
	}
	if c, ok := g.LookupCardForEffect(perm.ID); !ok || c.Controller != opp.ID || !effects.OnBattlefield(g, perm.ID) {
		return nil
	}
	if err := (effects.AddCounter{Target: perm.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.TapTarget{Target: perm.ID}).Apply(ctx); err != nil {
		return err
	}
	g.GoadForRestOfGameForEffect(perm.ID, ctx.Controller())
	return effects.GrantAbilitiesFor{
		Target:   perm.ID,
		Keys:     []string{jonIrenicusGrant},
		Duration: game.IndefiniteDuration(),
		Label:    "Jon Irenicus — it can't be sacrificed",
	}.Apply(ctx)
}
