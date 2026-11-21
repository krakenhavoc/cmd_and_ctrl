package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extraction Specialist — Creature — Human Rogue {2}{W}, 3/2 (Streets of
// New Capenna):
//
//	"Lifelink
//	 When this creature enters, return target creature card with mana
//	 value 2 or less from your graveyard to the battlefield. That creature
//	 can't attack or block for as long as you control this creature."
//
// The return is a keyed tail so the restriction lands on the permanent
// that actually arrived: a card that stayed in the graveyard (it left in
// response, or a replacement refused it) gets nothing. The duration is
// "for as long as you control this creature" (CR 611.2b), built as the
// ability resolves, so an Extraction Specialist that left the battlefield
// before the trigger resolved, or that is a new object, registers nothing,
// and losing control of it later ends the restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4164034a-5e59-4e40-a150-2c1000b0bd0d",
		Name:            "Extraction Specialist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Extraction Specialist — return target creature card with mana value 2 or less from your graveyard to the battlefield; it can't attack or block for as long as you control this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ReturnFromGraveyardTogether{
						Targets:  effects.LegalTargetCardIDs(ctx),
						ThenTail: extractionSpecialistLock,
					}.Apply(ctx)
				}),
				effects.TargetCardInGraveyard("target creature card with mana value 2 or less from your graveyard",
					effects.Creature(), effects.YouOwn(), effects.ManaValueLE(2))),
		},
	})
}

var extractionSpecialistLock = effects.TogetherTail("extraction-specialist/cant-attack-or-block",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		d, ok := effects.DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
		if !ok {
			return nil
		}
		for _, id := range entered {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
				Duration: d,
				Label:    "Extraction Specialist — can't attack or block for as long as you control it",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
