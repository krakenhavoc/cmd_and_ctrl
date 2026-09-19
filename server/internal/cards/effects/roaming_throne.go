package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roaming Throne — Artifact Creature — Golem {4}, 4/4:
//
//	"Ward {2}
//	 As this creature enters, choose a creature type.
//	 This creature is the chosen type in addition to its other types.
//	 If a triggered ability of another creature you control of the
//	 chosen type triggers, it triggers an additional time."
//
// Four lines, four pieces of vocabulary, none of them new: ward is a
// triggered ability (`Ward`), the type is the CR 614.12 prompt whose
// answer lands on Card.NamedTribe, "is the chosen type in addition to
// its other types" is a layer-4 APPEND, and the last line is #752's
// trigger doubler.
//
// # The doubler names the chosen type, so it reads the SOURCE
//
// `DoublesAbilitiesOf`'s plain filter is handed the ability's source
// and the doubler's controller, but not the doubler itself, and the
// type this card cares about lives on the doubler. `SourceMatch` is
// the option that gets the whole query, so that is what this uses —
// the same door Cloud, Midgar Mercenary needed for its attachment
// relationship.
//
// "ANOTHER creature you control": the Throne's own triggers are not
// doubled (it is its own chosen type, which would otherwise double
// its ward), and the controller check stays on, because the printed
// text says "you control".
//
// The source's type is read through its LAST-KNOWN characteristics,
// not its printed line: a creature that a lord or a Maskwood Nexus
// made the named type counts, and a dies-trigger is measured against
// what the creature was on the battlefield.
//
// Two Thrones naming the same type give three copies of a trigger,
// which is the rules answer and needs no card-side arithmetic — the
// harvester creates independent instances, each with its own optional
// choice and targets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3640c29b-1534-4952-b297-619ade948431",
		Name:         "Roaming Throne",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Roaming Throne"),
		Static:       []game.StaticAbility{IsAlsoTheChosenType()},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Roaming Throne — ward {2}"),
		},
		TriggerDoublers: []game.TriggerDoubler{roamingThroneDoubler()},
	})
}

// roamingThroneDoubler is the last line: another creature you control
// of the chosen type triggers an additional time.
func roamingThroneDoubler() game.TriggerDoubler {
	d := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{
		SourceMatch: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.Source.InstanceID == q.Doubler.InstanceID {
				return false
			}
			tribe := g.NamedTribeOf(q.Doubler.InstanceID)
			if tribe == "" {
				return false
			}
			src := characteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI)
			return src.IsCreature() && src.HasSubtype(tribe)
		},
	})
	d.Label = "Roaming Throne"
	return d
}
