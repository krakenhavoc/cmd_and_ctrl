package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Master, Transcendent — Legendary Artifact Creature — Mutant
// {1}{B}{G}{U}, 2/4:
//
//	"When The Master enters, target player gets two rad counters.
//	 {T}: Put target creature card in a graveyard that was milled this
//	 turn onto the battlefield under your control. It's a green Mutant
//	 with base power and toughness 3/3. (It loses its other colors and
//	 creature types.)"
//
// #2809. The tap ability targets a creature card in ANY graveyard whose
// most recent arrival there this turn was a mill (MilledThisTurn,
// Tato Farmer's clause). It enters under The Master's controller's
// control, and "it's a green Mutant with base power and toughness 3/3"
// is a continuous effect with no duration (CR 611.2a), pinned to the
// object that entered as Liliana, Death's Majesty's is, so it lasts as
// long as that permanent stays on the battlefield (CR 400.7):
//
//   - layer 4: its creature types become Mutant and nothing else
//     (game.SetCreatureTypesMod); an artifact creature is still an
//     artifact, a legendary one still legendary;
//   - layer 5: it is green and no other colour;
//   - layer 7b: base power and toughness 3/3, so counters and anthems
//     still apply on top. It keeps its abilities.
//
// The words are read through the ability's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40262ae5-1472-4777-b759-d1752c0b0148",
		Name:         "The Master, Transcendent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("The Master, Transcendent — target player gets two rad counters",
				targetPlayerGetsTwoRadCounters), TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put target creature card in a graveyard that was milled this turn onto the battlefield under your control. It's a green Mutant with base power and toughness 3/3.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target creature card in a graveyard that was milled this turn", Creature(), MilledThisTurn()),
			Effect:  theMasterReanimate,
		}},
	})
}

// theMasterReanimate puts the milled creature card onto the battlefield
// under the controller's control, then makes it a 3/3 green Mutant.
func theMasterReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{id},
		Controller: ctx.Controller(),
		ThenTail:   theMasterTranscendentMutant,
	}.Apply(ctx)
}

// theMasterTranscendentMutant is "it's a green Mutant with base power and
// toughness 3/3" as a keyed tail (#4253 slice 6). An on-disk key.
var theMasterTranscendentMutant = TogetherTail("the-master-transcendent/green-mutant",
	func(ctx *Context, entered []uuid.UUID, _ int) error {
		for _, e := range entered {
			if err := (ScopedEffectFor{
				Target: e,
				Mods: []game.Mod{
					game.SetCreatureTypesMod(ctx.TextSubtypes("Mutant")...),
					game.SetColorsMod(ctx.TextColors("G")...),
					game.SetBasePowerMod(3), game.SetBaseToughnessMod(3),
				},
				Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), e),
				Label:    "The Master, Transcendent — a green Mutant with base power and toughness 3/3",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
