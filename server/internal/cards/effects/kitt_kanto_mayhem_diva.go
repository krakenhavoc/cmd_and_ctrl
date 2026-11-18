package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitt Kanto, Mayhem Diva — Legendary Creature — Cat Bard Druid
// {1}{R}{G}{W}, 3/3 (Streets of New Capenna Commander, #5491):
//
//	"When Kitt Kanto enters, create a 1/1 green and white Citizen creature
//	 token.
//	 At the beginning of combat on each player's turn, you may tap two
//	 untapped creatures you control. When you do, target creature that
//	 player controls gets +2/+2 and gains trample until end of turn. Goad
//	 that creature."
//
// The second ability is Rent Is Due's tap-two pick (exactly two untapped
// creatures you control, or none: the "you may" declined), asked as the
// trigger resolves; any untapped creature may be tapped, summoning sick
// or not, since this is not a {T} cost (CR 302.6) and Kitt may tap herself.
// "When you do" is a reflexive trigger (CR 603.12) that fires only if both
// creatures really were tapped, and it chooses its target then: a creature
// controlled by the player whose turn it is (the active player, fixed as
// the trigger resolves), who may be Kitt's controller. The target gets
// +2/+2 and trample until end of turn and is goaded (CR 701.15), through
// the same one body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6e3af71-0d68-4bf0-b886-c29dae15e9ae",
		Name:         "Kitt Kanto, Mayhem Diva",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kitt Kanto, Mayhem Diva — create a 1/1 green and white Citizen creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1})),
			AtBeginningOfEachCombat("Kitt Kanto, Mayhem Diva — you may tap two untapped creatures you control; when you do, target creature that player controls gets +2/+2 and gains trample until end of turn, then goad it",
				kittKantoAsk),
		},
	})
}

func kittKantoUntapped(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && !c.Tapped && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func kittKantoAsk(g *game.Game, item *game.StackItem) error {
	candidates := kittKantoUntapped(g, item.Controller)
	if len(candidates) < 2 || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Kitt Kanto, Mayhem Diva — tap two untapped creatures you control?",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		ThenKey:  kittKantoTap,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// kittKantoTap taps the pick, then the reflexive trigger, keyed (#4253):
// the asking item rides the prompt as data and "exactly two" is a
// registered set rule. An on-disk key: never renamed, never reused.
var kittKantoTap = askedPickRule("choose-cards/kitt-kanto-tap-two",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 2 {
			return nil
		}
		still := kittKantoUntapped(ctx.Game, ctx.Controller())
		tapped := 0
		for _, id := range r.Picked {
			if !slices.Contains(still, id) {
				continue
			}
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
			tapped++
		}
		if tapped != 2 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Kitt Kanto, Mayhem Diva — target creature that player controls gets +2/+2 and gains trample until end of turn; goad it",
			Body:   kittKantoBoostBody,
			Params: game.EffectParams{Player: activePlayerIDOf(ctx.Game)},
		}.Apply(ctx)
	})

var kittKantoBoostBody = game.ReflexiveBody("kitt-kanto/boost-and-goad",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Kitt Kanto — +2/+2"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Kitt Kanto — trample"}).Apply(ctx); err != nil {
			return err
		}
		return goadAndScheduleClear(ctx, []uuid.UUID{id})
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCreature("target creature that player controls", ControlledBy(p.Player))
	})
