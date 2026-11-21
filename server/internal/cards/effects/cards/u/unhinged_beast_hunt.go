package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	unhingedBeastHuntLife = "unhinged-beast-hunt/tap-gain-1"
	unhingedBeastHuntTap  = "unhinged-beast-hunt/attacks-tap-matching"
)

// Unhinged Beast Hunt — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Unhinged", "Beast", "Hunt"; three art stickers.
//	{TK}{TK} — {T}: You gain 1 life.
//	{TK}{TK}{TK}{TK} — Whenever this creature attacks, tap each creature an opponent controls with the same power and/or same toughness as this creature.
//	{TK}{TK} — 4/1
//	{TK}{TK}{TK} — 2/6
//
// The first sticker gives an ability that taps the permanent and gains 1 life. The second gives an attack trigger that taps each creature an opponent controls with the same power and/or the same toughness as this creature, read as the ability resolves (the creature as it is then).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea1fb0c3-d52c-4435-af2c-4b74f31189f7",
		Name:         "Unhinged Beast Hunt",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: unhingedBeastHuntLife,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: You gain 1 life.",
					Purpose: game.Purpose{Answers: game.AnswerValue},
					Cost:    effects.TapCost(),
					Effect:  effects.Do(effects.GainLife{Amount: 1}),
				}},
				Text: "{T}: You gain 1 life.",
			},
			{
				Key: unhingedBeastHuntTap,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Unhinged Beast Hunt sticker — tap each creature an opponent controls with the same power and/or same toughness as this creature",
						tapOpponentsCreaturesWithTheSamePowerOrToughness),
				},
				Text: "Whenever this creature attacks, tap each creature an opponent controls with the same power and/or same toughness as this creature.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: unhingedBeastHuntLife},
			{Grant: unhingedBeastHuntTap},
		}},
	})
}

// tapOpponentsCreaturesWithTheSamePowerOrToughness is Unhinged Beast Hunt's
// attack trigger: the creatures the controller's opponents control whose
// power equals the source's, or whose toughness does, are tapped. The set
// is read as the ability resolves.
func tapOpponentsCreaturesWithTheSamePowerOrToughness(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if ctx.IsNewSourceObject(item.SourceCardID) {
		return nil
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	power, toughness := src.CurrentPower(), src.CurrentToughness()
	var victims []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller || !c.IsCreature() {
			continue
		}
		if c.CurrentPower() == power || c.CurrentToughness() == toughness {
			victims = append(victims, c.InstanceID)
		}
	}
	for _, id := range victims {
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
