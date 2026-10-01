package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lunarch Veteran // Luminous Phantom (#1855, ADR 0107 §4) — a disturb
// card.
//
// Front face, Creature — Human Cleric {W}, 1/1:
//
//	"Whenever another creature you control enters, you gain 1 life.
//	 Disturb {1}{W} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit Cleric, 1/1:
//
//	"Flying
//	 Whenever another creature you control leaves the battlefield, you
//	 gain 1 life.
//	 If Luminous Phantom would be put into a graveyard from anywhere,
//	 exile it instead."
//
// The front face's trigger is the shared "another creature you control
// enters" condition. The back face's is any departure — dying, being
// exiled, bounced — of another creature its controller controlled as
// it left, read from the last-known information the exit records (CR
// 603.10a), so a creature leaving at the same time as the Phantom still
// counts. The exile clause is DisturbedExile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         lunarchVeteranOracleID,
		Name:             "Lunarch Veteran",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Lunarch Veteran — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
	Register(Spec{
		OracleID:        lunarchVeteranOracleID + "#1",
		Name:            "Luminous Phantom",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, anotherCreatureYouControlLeft, "Luminous Phantom — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Replacements: []game.ReplacementEffect{DisturbedExile("Luminous Phantom")},
	})
}

const lunarchVeteranOracleID = "0761a0e7-d443-4bab-bb15-307c83d4a6a1"

// anotherCreatureYouControlLeft is "whenever another creature you
// control leaves the battlefield": aCreatureYouControlLeft without the
// source itself.
func anotherCreatureYouControlLeft(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ev.CardID != source.InstanceID && aCreatureYouControlLeft(ev, source, lki, g)
}
