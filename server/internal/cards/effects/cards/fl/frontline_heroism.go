package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frontline Heroism — Enchantment {2}{R}:
//
//	"When this enchantment enters, create a 1/1 red Soldier creature
//	 token with haste.
//	 Whenever you cast a spell that targets only a single creature you
//	 control, create a 1/1 red Soldier creature token with haste, then
//	 copy that spell. The copy targets that token."
//
// Two printed triggers, two registered triggers. The cast trigger reads
// the spell's targets as it is cast (the Ivy, Gleeful Spellthief
// condition: exactly one target, a creature you control). The token is
// created first, then the copy (CR 707.10) is put on the stack under the
// trigger's controller with its target changed to the token; a spell
// that has already left the stack leaves the token and no copy.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c340b82-22e1-445a-b236-82471b442031",
		Name:         "Frontline Heroism",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Frontline Heroism — create a 1/1 red Soldier creature token with haste",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Soldier with haste"), N: 1})),
			effects.On(game.EventCast,
				effects.DmuB17CastSpellTargetsOnly(true, func(target game.Card, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				}),
				"Frontline Heroism — create a 1/1 red Soldier with haste, then copy that spell; the copy targets that token",
				frontlineHeroismCopy),
		},
	})
}

func frontlineHeroismCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	spell := item.Trigger.Event.CardID
	target, haveTarget := g.StackItemSoleTargetForEffect(spell)
	ids, err := g.CreateTokensForEffect(item.Controller, ctx.TextToken(effects.TokenCard("1/1 red Soldier with haste")), 1, game.TokenEntryOptions{})
	if err != nil || len(ids) == 0 || !haveTarget || target.Kind != game.TargetCard {
		return err
	}
	g.CopySpellRetargetedForEffect(spell, item.Controller, target.ID, ids[0])
	return nil
}
