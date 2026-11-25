package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ledev Champion — Creature — Elf Knight {1}{G}{W}, 2/2 (Guilds of Ravnica,
// #5525):
//
//	"Whenever this creature attacks, you may tap any number of untapped
//	 creatures you control. This creature gets +1/+1 until end of turn for each
//	 creature tapped this way.
//	 {3}{G}{W}: Create a 1/1 white Soldier creature token with lifelink."
//
// Siege Striker's attack trigger: the controller picks any number of untapped
// creatures (none included), each one that really becomes tapped counts, and
// the Champion, if it is still the same permanent, gets +1/+1 for each. The pick
// is a keyed choose_cards prompt, so a table waiting on it is a restore point.
// The token is Dawn of Hope's 1/1 white Soldier with lifelink.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97b2aa77-122f-41e3-ba28-c7040dae0f01",
		Name:         "Ledev Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ledev Champion — you may tap any number of untapped creatures you control; +1/+1 for each", ledevChampionAsk),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}{W}: Create a 1/1 white Soldier creature token with lifelink.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{3}{G}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// ledevChampionAsk offers the controller's untapped creatures.
func ledevChampionAsk(g *game.Game, item *game.StackItem) error {
	cands := effects.UntappedCreaturesControlledBy(g, item.Controller)
	if len(cands) == 0 || !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Ledev Champion — you may tap any number of untapped creatures you control",
		Cards:    cands,
		Min:      0,
		Max:      0, // zero means every candidate: "any number"
		LiveZone: game.ZoneBattlefield,
		ThenKey:  ledevChampionTap,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// ledevChampionTap taps the picks and gives the Champion +1/+1 for each one
// that really became tapped. An on-disk key: never renamed, never reused.
var ledevChampionTap = effects.AskedPick("choose-cards/s82-old-b190-ledev-champion-tap", func(ctx *effects.Context, r game.CardsChosen) error {
	tapped := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() {
			continue
		}
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		tapped++
	}
	if tapped == 0 || !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.BoostUntilEOT{Target: ctx.Source(), Power: tapped, Toughness: tapped, Label: "Ledev Champion — +1/+1 for each creature tapped this way"}.Apply(ctx)
})
