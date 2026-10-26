package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuffie, Materia Hunter — Legendary Creature — Human Ninja {2}{R}, 3/3
// (#3656):
//
//	"Ninjutsu {1}{R}
//	 When Yuffie enters, gain control of target noncreature artifact for
//	 as long as you control Yuffie. Then you may attach an Equipment you
//	 control to Yuffie."
//
// Ninjutsu is the keyword's activated ability. The theft is Orcish
// Squatters' for-as-long-as-you-control-the-source control change: a
// Yuffie that left and came back before the trigger resolved is not the
// one named, so nothing is taken. The attach offer follows in the same
// resolution: a choose-cards prompt over the Equipment the controller
// has, now including the one just taken, with a floor of zero. The pick
// is re-checked on submit, and the attach refuses an Equipment or a
// Yuffie that has gone.
//
// No simplification.
func init() {
	const label = "Yuffie, Materia Hunter — gain control of target noncreature artifact for as long as you control Yuffie; then you may attach an Equipment you control to Yuffie"
	Register(Spec{
		OracleID:     "acdcbb49-9d71-4c36-8ccb-650fb11c3a58",
		Name:         "Yuffie, Materia Hunter",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(label, yuffieETB),
				TargetPermanent("target noncreature artifact", Artifact(), Noncreature())),
		},
	})
}

func yuffieETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if target, ok := b16FirstLegalTargetCard(ctx); ok {
		if d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller()); ok {
			if err := (GainControl{Target: target, Duration: d, Label: "Yuffie, Materia Hunter — gain control of the artifact"}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	controller, yuffie := item.Controller, item.SourceCardID
	if !g.Battlefield.Contains(yuffie) {
		return nil
	}
	var equipment []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Equipment") && c.InstanceID != yuffie {
			equipment = append(equipment, c.InstanceID)
		}
	}
	if len(equipment) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   yuffie,
		Question: "Yuffie, Materia Hunter — you may attach an Equipment you control to Yuffie",
		Cards:    equipment,
		Min:      0,
		Max:      1,
		ThenKey:  yuffieAttached,
		LiveZone: game.ZoneBattlefield,
	})
	return nil
}

// yuffieAttached is the optional Equipment attach, keyed (#4253): Yuffie is
// the asking card and the chooser her controller, so the prompt carries
// nothing else.
var yuffieAttached = game.RegisterChooseCardsThen("choose-cards/yuffie-materia-hunter-attach",
	func(g *game.Game, r game.CardsChosen) error {
		yuffie, controller := r.Source, r.Chooser
		if len(r.Picked) == 0 || !g.Battlefield.Contains(yuffie) || !g.Battlefield.Contains(r.Picked[0]) {
			return nil
		}
		eq, ok := g.LookupCardForEffect(r.Picked[0])
		if !ok || eq.Controller != controller || !eq.HasSubtype("Equipment") {
			return nil
		}
		return g.AttachForEffect(r.Picked[0], game.TargetRef{Kind: game.TargetCard, ID: yuffie})
	})
