package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "2edd018b-6d82-4ff1-b9d0-5da9b628af3f",
		Name:         "Dalek Intensive Care",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Dalek Intensive Care — exile a non-Dalek creature you control; if you do, create a hasty 3/3 Dalek",
				func(g *game.Game, item *game.StackItem) error {
					return ChoosePermanents{
						Question:   "Dalek Intensive Care — exile a non-Dalek creature you control",
						Candidates: dalekIntensiveCareCandidates,
						Keyed:      dalekIntensiveCareExile,
					}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("Dalek Intensive Care — target Dalek you control deals damage equal to its power to target creature you don't control",
				func(g *game.Game, item *game.StackItem) error { return oneSidedBite(item, NewContext(g, item)) }),
				Clauses(
					TargetPermanent("target Dalek you control", Subtype("Dalek"), YouControl()),
					TargetCreature("target creature you don't control", OpponentControls()),
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
var dalekIntensiveCareExile = RegisterPermanentsPick("dalek-intensive-care/exile", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			return ExileTarget{Target: id, ThenTail: ctx.ItemTail(dalekIntensiveCareCreate)}.Apply(ctx)
		}
		return nil
	},
})

// dalekIntensiveCareCreate is "If you do, create a 3/3 black Dalek
// artifact creature token with menace", told the creature that reached
// exile. An on-disk key.
var dalekIntensiveCareCreate = CardTail("dalek-intensive-care/create-dalek", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("3/3 black Dalek artifact with menace")), Count: 1}},
	}, ctx.ItemTail(dalekIntensiveCareHaste))
})

// dalekIntensiveCareHaste is "It gains haste until end of turn", over the
// token the creation made. An on-disk key.
var dalekIntensiveCareHaste = CardTail("dalek-intensive-care/haste", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Dalek Intensive Care — haste until end of turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
