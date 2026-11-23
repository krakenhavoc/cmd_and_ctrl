package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vislor Turlough — Legendary Creature — Rogue {3}{B}, 2/5 (Doctor Who,
// #5948):
//
//	"Deal with the Black Guardian — When Vislor Turlough enters, you may have
//	 an opponent gain control of it. If you do, it's goaded for as long as they
//	 control it.
//	 At the beginning of your end step, draw a card, then you lose life equal
//	 to the number of cards in your hand.
//	 Doctor's companion (You can have two commanders if the other is the
//	 Doctor.)"
//
// The enter trigger asks its controller which opponent (or nobody) in one
// prompt; the chosen opponent gains control with no end (CR 611.2a, an
// indefinite control effect), and the goad is Kang's "for as long as" record
// (DurationWhileYouControlSource, with the opponent as the controller), made
// by the trigger's controller (CR 701.15b). "Your end step" is whoever
// controls Vislor then; the draw happens first and the loss counts the hand
// after it. Doctor's companion is a deck-construction rule read off the
// oracle text (Nardole).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdf5c31c-c324-471f-b6d7-06a06e91daa3",
		Name:         "Vislor Turlough",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vislor Turlough — you may have an opponent gain control of it; if you do, it's goaded for as long as they control it", vislorAsk),
			effects.AtYourEndStep("Vislor Turlough — draw a card, then you lose life equal to the number of cards in your hand", vislorEndStep),
		},
	})
}

const vislorKeepLabel = "Keep control of it"

func vislorAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	var options []game.ChoiceOption
	for _, o := range vislorOpponents(ctx) {
		options = append(options, game.ChoiceOption{Label: "Have " + effects.AttackTargetName(g, o) + " gain control of it"})
	}
	if len(options) == 0 {
		return nil
	}
	options = append(options, game.ChoiceOption{Label: vislorKeepLabel})
	return effects.PickOption{
		Question: "Vislor Turlough — have an opponent gain control of it?",
		Options:  options,
		ThenKey:  vislorAnswer,
		WithItem: true,
	}.Apply(ctx)
}

// vislorOpponents are the opponents still in the game, in seat order from the
// controller. The prompt and its answer read the same list.
func vislorOpponents(ctx *effects.Context) []uuid.UUID {
	var out []uuid.UUID
	for _, o := range ctx.Opponents() {
		if p := ctx.Game.PlayerByIDForEffect(o); p != nil && !p.Eliminated {
			out = append(out, o)
		}
	}
	return out
}

// vislorAnswer hands Vislor to the chosen opponent and goads it for as long
// as they control it. An on-disk key: never renamed, never reused.
var vislorAnswer = effects.OptionPickThen("option-pick/vislor-turlough-deal-with-the-black-guardian", func(ctx *effects.Context, r game.OptionPicked) error {
	opponents := vislorOpponents(ctx)
	if r.Index < 0 || r.Index >= len(opponents) {
		return nil
	}
	self, opp, you := ctx.Item.SourceCardID, opponents[r.Index], ctx.Controller()
	if !effects.OnBattlefield(ctx.Game, self) || ctx.IsNewSourceObject(self) {
		return nil
	}
	if err := (effects.GainControl{Target: self, Controller: opp, Duration: game.IndefiniteDuration(), Label: "Vislor Turlough — an opponent gains control"}).Apply(ctx); err != nil {
		return err
	}
	// The control change is a scoped effect the layers have not applied yet,
	// so the duration is written out: "for as long as they control it".
	now, ok := ctx.Game.LookupCardForEffect(self)
	if !ok {
		return nil
	}
	dur := game.Duration{
		Kind:            game.ForAsLongAs,
		Condition:       game.WhileYouControlSource,
		Player:          opp,
		Source:          self,
		SourceEnteredAt: now.EnteredBattlefieldAt,
	}
	return effects.ScopedEffectFor{
		Target:   self,
		Mods:     []game.Mod{game.GoadMod(you)},
		Duration: dur,
		Label:    "Vislor Turlough — goaded for as long as they control it",
	}.Apply(ctx)
})

func vislorEndStep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Eliminated || p.Hand == nil {
		return nil
	}
	if n := p.Hand.Size(); n > 0 {
		return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -n)
	}
	return nil
}
