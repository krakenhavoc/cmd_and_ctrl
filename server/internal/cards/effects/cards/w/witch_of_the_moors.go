package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch of the Moors — Creature — Human Warlock {3}{B}{B}, 4/4
// (EDHREC rank 2569):
//
//	"Deathtouch
//	 At the beginning of your end step, if you gained life this turn,
//	 each opponent sacrifices a creature of their choice and you
//	 return up to one target creature card from your graveyard to
//	 your hand."
//
// The lifegain deck's edict engine. Deathtouch rides
// PrintedKeywords. The end-step trigger's intervening-if is "you
// gained life this turn" (b15LifeGainedThisTurn, the event-log
// walk), checked as the trigger would fire and again as it resolves
// (CR 603.4). "Each opponent sacrifices a creature of their choice"
// is the EachPlayerSacrifices fan-out — one prompt per opponent, no
// targeting, so hexproof is irrelevant.
//
// "Up to one target creature card from your graveyard" is declared
// as TWO TriggeredAbility entries with one label, the Hazel's
// Brewmaster split: the engine drops a targeted trigger outright
// when its legal set is empty (CR 603.3d), which is wrong for "up to
// one … AND each opponent sacrifices" — with no creature card in the
// graveyard the edict must still happen. So the targeted entry
// fires only while the controller's graveyard holds a creature card
// and the untargeted entry only when it does not; exactly one
// applies to any end step. A chosen card that leaves the graveyard
// in response counters the ability by game rules, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ad97dd9-1342-4ed4-bea9-1bb21d748e04",
		Name:            "Witch of the Moors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventBeginEndStep},
				Key:     b24WitchOfTheMoorsLabel,
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B24YourEndStepAndYouGainedLifeThisTurn(ev, source, g) && effects.B24GraveyardHasCreatureCard(g, source.Controller)
				},
				Targets: effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
				Effect:  b24WitchOfTheMoorsEdictAndReturn,
			}),
			effects.TriggerEdict(effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B24YourEndStepAndYouGainedLifeThisTurn(ev, source, g) && !effects.B24GraveyardHasCreatureCard(g, source.Controller)
			}, b24WitchOfTheMoorsLabel, b24WitchOfTheMoorsEdictAndReturn)),
		},
	})
}

// b24WitchOfTheMoorsLabel is the stack label both declarations share
// — one printed ability, one name on the stack.
const b24WitchOfTheMoorsLabel = "Witch of the Moors — each opponent sacrifices a creature, return up to one creature card to your hand"

// b24WitchOfTheMoorsEdictAndReturn is the trigger's body: re-check
// the intervening-if, fan out the edict, then return the chosen
// graveyard card if one was chosen and is still there.
func b24WitchOfTheMoorsEdictAndReturn(g *game.Game, item *game.StackItem) error {
	if effects.B15LifeGainedThisTurn(g, item.Controller) <= 0 {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
