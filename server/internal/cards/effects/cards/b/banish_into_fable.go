package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banish into Fable — Instant {4}{W}{U} (Throne of Eldraine, S82 old sets):
//
//	"When you cast this spell from your hand, copy it if you control an
//	 artifact, then copy it if you control an enchantment. You may choose
//	 new targets for the copies.
//	 Return target nonland permanent to its owner's hand. You create a 2/2
//	 white Knight creature token with vigilance."
//
// A cast trigger (CR 603.2) that resolves above the spell: it reads "you
// control an artifact" and then "an enchantment" as it resolves, in that
// order, so an artifact made by the first copy would count for the second
// test. Each copy is its own CR 707.10 copy under the caster, with the
// new-target prompt. A spell cast from anywhere but the hand does not
// trigger it. The spell itself bounces its target and then makes the Knight.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f18e63d-ab1f-426c-afd5-ce89433394ca",
		Name:         "Banish into Fable",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		Triggered: []game.TriggeredAbility{
			fromHandOnly(effects.WhenYouCastThisSpell("Banish into Fable — copy it if you control an artifact, then copy it if you control an enchantment", banishIntoFableCopies)),
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.OldB145BounceTheTargetPermanent),
			effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1},
		),
	})
}

// fromHandOnly narrows a cast trigger to a spell cast from its owner's hand.
func fromHandOnly(t game.TriggeredAbility) game.TriggeredAbility {
	inner := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ev.OldZone == game.ZoneHand && inner(ev, source, lki, g)
	}
	return t
}

func banishIntoFableCopies(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, kind := range []effects.CardPredicate{effects.Artifact(), effects.Enchantment()} {
		if len(effects.MatchingBattlefield(ctx, effects.And(kind, effects.YouControl()))) == 0 {
			continue
		}
		if err := effects.GapB106CopyTheCastSpell(g, item); err != nil {
			return err
		}
	}
	return nil
}
