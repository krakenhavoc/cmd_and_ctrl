package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kami of Industry — Creature — Spirit {4}{R}, 3/6 (Kamigawa: Neon Dynasty,
// S82):
//
//	"When this creature enters, return target artifact card with mana value
//	 3 or less from your graveyard to the battlefield. It gains haste.
//	 Sacrifice it at the beginning of the next end step."
//
// Ayara, Furnace Queen's return: the card comes back under your control, is
// given haste for as long as it stays, and a delayed trigger sacrifices that
// object at the next end step (CR 603.7); if the artifact left the graveyard
// in response nothing returns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b308fa9a-3f69-4e59-9d4b-8007d9a2b1f1",
		Name:         "Kami of Industry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Kami of Industry — return target artifact card with mana value 3 or less from your graveyard to the battlefield; it gains haste; sacrifice it at the beginning of the next end step",
				kamiOfIndustryReturn),
				TargetCardInGraveyard("target artifact card with mana value 3 or less from your graveyard", YouOwn(), Artifact(), ManaValueLE(3))),
		},
	})
}

func kamiOfIndustryReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false, nil, ctx.ItemTail(kamiOfIndustryEntered))
}

// kamiOfIndustryEntered gives the returned artifact haste and sacrifices it
// at the next end step. An on-disk key: never renamed, never reused.
var kamiOfIndustryEntered = EnteredTail("kami-of-industry/entered", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Kami of Industry — it gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label: "Kami of Industry — sacrifice it",
		Cards: []uuid.UUID{entered},
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
