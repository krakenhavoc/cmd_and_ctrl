package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Covert Cutpurse // Covetous Geist (#1855, ADR 0107 §4) — a disturb
// card.
//
// Front face, Creature — Human Rogue {2}{B}, 2/1:
//
//	"When this creature enters, destroy target creature you don't
//	 control that was dealt damage this turn.
//	 Disturb {4}{B} (You may cast this card from your graveyard
//	 transformed for its disturb cost.)"
//
// Back face, Creature — Spirit Rogue, 2/2:
//
//	"Flying, deathtouch
//	 If Covetous Geist would be put into a graveyard from anywhere,
//	 exile it instead."
//
// ADR 0107 §4 held this card for "was dealt damage this turn". The
// engine already records it: every damage event names the damage
// actually dealt, after prevention, and the turn's events are kept
// (Game.EventsThisTurn). DealtDamageThisTurn reads them, and a creature
// that entered the battlefield after its damage is a new object that
// was never dealt it (CR 400.7). The clause is a target restriction,
// so it is judged as the trigger goes on the stack (CR 603.3d) and
// again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         covertCutpurseOracleID,
		Name:             "Covert Cutpurse",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Disturb("{4}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Covert Cutpurse — destroy target creature you don't control that was dealt damage this turn", destroyChosenPermanent),
				TargetCreature("target creature you don't control that was dealt damage this turn", OpponentControls(), DealtDamageThisTurn())),
		},
	})
	Register(Spec{
		OracleID:        covertCutpurseOracleID + "#1",
		Name:            "Covetous Geist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Replacements:    []game.ReplacementEffect{DisturbedExile("Covetous Geist")},
	})
}

const covertCutpurseOracleID = "5d0b8dc6-f4b6-4650-805d-4240d4a4ab82"
