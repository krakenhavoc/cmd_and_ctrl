package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// M'Baku, Jabari Chieftain — Legendary Creature — Human Noble Warrior
// {1}{G}{G}, 4/3 (Marvel Super Heroes Commander, #4455):
//
//	"At the beginning of your end step, if there is no monarch, target
//	 opponent becomes the monarch.
//	 Whenever a creature attacks one of your opponents, if that player is
//	 the monarch, that creature gets +1/+1 and gains trample until end of
//	 turn."
//
// Both triggers carry an intervening "if" (CR 603.4), checked when the
// trigger event happens and again as it resolves: nobody has taken the
// crown by the time the first resolves, and the defending player still
// holds it when the second does. "A creature attacks one of your
// opponents" is CR 508.3a: the creature was declared attacking that
// opponent ITSELF (attackedPlayer). A creature attacking the monarch's
// planeswalker or battle is not attacking the monarch and gets nothing
// (#4839). "One of your opponents" excludes M'Baku's own controller. The
// creature may be any player's, M'Baku's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e4656ba-c375-4ef8-ba1d-5a7b0833242f",
		Name:         "M'Baku, Jabari Chieftain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.CondBoard{
				Name: "there is no monarch",
				Test: func(ctx *effects.Context) bool { return ctx.Game.Monarch == uuid.Nil },
			}, effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, ch, g) && g.Monarch == uuid.Nil
			}, "M'Baku — target opponent becomes the monarch", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.BecomeTheMonarch{Player: opp}.Apply(ctx)
			})), effects.TargetPlayer("target opponent", effects.Opponent())),
			effects.InterveningIf(effects.CondBoard{
				Name: "the defending player is the monarch",
				Test: func(ctx *effects.Context) bool {
					if ctx.Item == nil || ctx.Item.Trigger == nil {
						return false
					}
					return mbakuAttackedTheMonarch(ctx.Game, ctx.Item.Trigger.Event, ctx.Controller())
				},
			}, effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return mbakuAttackedTheMonarch(g, ev, source.Controller)
			}, "M'Baku — the attacker gets +1/+1 and gains trample until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := item.Trigger.Event.CardID
				if !effects.OnBattlefield(g, id) {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "M'Baku — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "M'Baku — trample"}.Apply(ctx)
			})),
		},
	})
}

// mbakuAttackedTheMonarch is both halves of the second ability's trigger and
// its intervening "if": the attack is aimed at an opponent of `you` (the
// player, not a planeswalker or battle of theirs), and that opponent is the
// monarch.
func mbakuAttackedTheMonarch(g *game.Game, ev game.Event, you uuid.UUID) bool {
	defender := effects.AttackedPlayer(g, ev)
	return defender != uuid.Nil && defender != you && effects.YoureTheMonarch(g, defender)
}
