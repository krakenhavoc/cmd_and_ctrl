package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kimahri, Valiant Guardian — Legendary Creature — Cat Warrior
// {2}{G}{U}, 3/3:
//
//	"Vigilance
//	 Ronso Rage — At the beginning of combat on your turn, put a +1/+1
//	 counter on Kimahri and tap target creature an opponent controls.
//	 Then you may have Kimahri become a copy of that creature, except its
//	 name is Kimahri, Valiant Guardian and it has vigilance and this
//	 ability."
//
// The counter, the tap, then the "you may" asked as the trigger
// resolves (MayChoice). The copy is Dimir Doppelganger's shape: an
// indefinite copy effect (CR 707.2, 611.2a) with an except clause that
// renames it, adds vigilance and grants Ronso Rage back as a bundle, so
// the copied Kimahri keeps raging on later turns. The counter stays (a
// counter is not a copiable value), and a later Ronso Rage copies over
// the earlier one with a newer timestamp. With the target gone at
// resolution the counter is still put and nothing is copied (CR 608.2b).
//
// No simplification.
const (
	kimahriName  = "Kimahri, Valiant Guardian"
	kimahriGrant = "kimahri-valiant-guardian/ronso-rage"
	kimahriLabel = "Ronso Rage — put a +1/+1 counter on Kimahri and tap target creature an opponent controls"
)

func kimahriRonsoRage() game.TriggeredAbility {
	return effects.Targeting(effects.AtBeginningOfYourCombat(kimahriLabel, kimahriRonsoRageEffect),
		effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))
}

func init() {
	effects.Register(effects.Spec{
		OracleID:        "6d42e81a-84d3-4c90-9c69-c0d562c66cac",
		Name:            kimahriName,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Grants: []effects.AbilityGrant{{
			Key:       kimahriGrant,
			Text:      "Ronso Rage — At the beginning of combat on your turn, put a +1/+1 counter on Kimahri and tap target creature an opponent controls. Then you may have Kimahri become a copy of that creature, except its name is Kimahri, Valiant Guardian and it has vigilance and this ability.",
			Triggered: []game.TriggeredAbility{kimahriRonsoRage()},
		}},
		Triggered: []game.TriggeredAbility{kimahriRonsoRage()},
	})
}

func kimahriRonsoRageEffect(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self := item.SourceCardID
	if err := (effects.AddCounter{Target: self, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	victim, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.TapTarget{Target: victim}).Apply(ctx); err != nil {
		return err
	}
	return effects.MayChoice{
		Question: "Ronso Rage — have Kimahri become a copy of that creature?",
		Then:     kimahriAnswer,
		Carry:    []uuid.UUID{victim},
	}.Apply(ctx)
}

// kimahriAnswer is the "yes": Kimahri becomes a copy of the tapped creature
// (the carried ID). Keyed (#4291).
var kimahriAnswer = effects.MayYes("confirm/kimahri-ronso-rage-copy", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	victim := effects.CarryAt(r, 0)
	if !effects.OnBattlefield(ctx.Game, victim) {
		return nil
	}
	return effects.BecomeCopy{
		Targets:  []uuid.UUID{ctx.Item.SourceCardID},
		Of:       victim,
		Duration: effects.CopyIndefinite,
		Except: func(v *game.PrintedValues) {
			v.SetName(kimahriName)
			v.AddKeyword("vigilance")
			v.GrantAbility(kimahriGrant)
		},
		Label: "Ronso Rage — Kimahri becomes a copy",
	}.Apply(ctx)
})
