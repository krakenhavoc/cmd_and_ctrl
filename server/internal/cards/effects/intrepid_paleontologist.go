package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intrepid Paleontologist — Creature — Human Druid {1}{G}, 1/2 (Lost Caverns
// of Ixalan, #4470):
//
//	"{T}: Add one mana of any color.
//	 {2}: Exile target card from a graveyard.
//	 You may cast Dinosaur creature spells from among cards you own exiled
//	 with this creature. If you cast a spell this way, that creature enters
//	 with a finality counter on it."
//
// The exile is linked to this creature's object (ExileTarget.Linked, CR 607.2a),
// so cards another effect exiled, or that an earlier life of this card exiled
// (CR 400.7), are not its cards. The cast is Valgavoth's standing permission
// over cards exiled with the source (PermissionFilter.ExiledWithSource), narrowed
// to cards its holder owns (OwnedByHolder) and to Dinosaur creature spells, at
// the ordinary timing and for the ordinary cost (the ruling: it adds nothing but
// the zone). The finality counter is the permission's rider
// (CastPermission.EntersWithCounterKind): it rides the spell to its entry event, so
// a Dinosaur hard-cast from hand does not get one, and finality's own
// replacement (CR 122.1h) does the exiling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6026309-aae9-43e7-938b-69db7af7139c",
		Name:         "Intrepid Paleontologist",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Exile target card from a graveyard.",
			Cost:    ManaCost("{2}"),
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  avB7ExileTargetLinked,
		}},
		CastPermissions: []game.CastPermission{{
			Zone: game.ZoneExile,
			Filter: game.PermissionFilter{
				ExiledWithSource: true,
				OwnedByHolder:    true,
				CreatureOnly:     true,
				CreatureType:     "Dinosaur",
			},
			EntersWithCounterKind: game.CounterFinality,
			EntersWithCounters:    1,
			Label:                 "Cast a Dinosaur creature spell from among cards you own exiled with Intrepid Paleontologist",
		}},
	})
}
