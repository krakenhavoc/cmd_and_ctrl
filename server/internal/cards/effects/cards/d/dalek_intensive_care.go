package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dalek Intensive Care — Plane — The Dalek Asylum (Doctor Who, #4870):
//
//	"When you planeswalk to Dalek Intensive Care and at the beginning of
//	 your upkeep, exile a non-Dalek creature you control. If you do,
//	 create a 3/3 black Dalek artifact creature token with menace. It
//	 gains haste until end of turn.
//	 Whenever chaos ensues, target Dalek you control deals damage equal
//	 to its power to target creature you don't control."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability is one ability with two trigger conditions. The exile
// is a choice at resolution, not a target (CR 608.2), mandatory while the
// planar controller has a non-Dalek creature; "if you do" reads whether
// the creature really reached exile (a keyed tail), and the Dalek gains
// haste until end of turn. The chaos ability (CR 311.7) is a one-sided
// bite: the Dalek, judged as it resolves, deals damage equal to its power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2edd018b-6d82-4ff1-b9d0-5da9b628af3f",
		Name:         "Dalek Intensive Care",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenYouPlaneswalkToThisAndAtYourUpkeep("Dalek Intensive Care — exile a non-Dalek creature you control; if you do, create a hasty 3/3 Dalek",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChoosePermanents{
						Question:   "Dalek Intensive Care — exile a non-Dalek creature you control",
						Candidates: dalekIntensiveCareCandidates,
						Keyed:      dalekIntensiveCareExile,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.Targeting(effects.WheneverChaosEnsues("Dalek Intensive Care — target Dalek you control deals damage equal to its power to target creature you don't control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.OneSidedBite(item, effects.NewContext(g, item))
				}),
				effects.Clauses(
					effects.TargetPermanent("target Dalek you control", effects.Subtype("Dalek"), effects.YouControl()),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				)),
		},
	})
}

// dalekIntensiveCareCandidates is the planar controller's non-Dalek
// creatures; one must be chosen.
func dalekIntensiveCareCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsCreature() && !c.HasSubtypeWord("Dalek") {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil, 0, 0
	}
	return ids, 1, 1
}

// dalekIntensiveCareExile exiles the chosen creature; "if you do" is the
// exile's tail. An on-disk key: never renamed, never reused.
var dalekIntensiveCareExile = effects.RegisterPermanentsPick("dalek-intensive-care/exile", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			return effects.ExileTarget{Target: id, ThenTail: ctx.ItemTail(dalekIntensiveCareCreate)}.Apply(ctx)
		}
		return nil
	},
})

// dalekIntensiveCareCreate is "If you do, create a 3/3 black Dalek
// artifact creature token with menace", told the creature that reached
// exile. An on-disk key.
var dalekIntensiveCareCreate = effects.CardTail("dalek-intensive-care/create-dalek", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/3 black Dalek artifact with menace")), Count: 1}},
	}, ctx.ItemTail(dalekIntensiveCareHaste))
})

// dalekIntensiveCareHaste is "It gains haste until end of turn", over the
// token the creation made. An on-disk key.
var dalekIntensiveCareHaste = effects.CardTail("dalek-intensive-care/haste", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Dalek Intensive Care — haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
